// Package event defines the hyp event envelope, its closed vocabularies and
// content limits. Derived state never appears here: no payload carries a
// writable status.
package event

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"
)

const Version = 1

// Event is one line of a .tree/events/<actor>.jsonl file.
type Event struct {
	V       int             `json:"v"`
	ID      string          `json:"id"`
	TS      string          `json:"ts"`
	Actor   string          `json:"actor"`
	Type    string          `json:"type"`
	Node    string          `json:"node"`
	Payload json.RawMessage `json:"payload"`
	Why     string          `json:"why"`
}

// Event types.
const (
	TypeNodeAdd            = "node.add"
	TypeResultRecord       = "result.record"
	TypeResultRetract      = "result.retract"
	TypeHypothesisAbandon  = "hypothesis.abandon"
	TypeHypothesisAmend    = "hypothesis.amend"
	TypeDQSelect           = "dq.select"
	TypeDQUnselect         = "dq.unselect"
	TypeDQDefer            = "dq.defer"
	TypeDQWithdraw         = "dq.withdraw"
	TypeExperimentWithdraw = "experiment.withdraw"
	TypeDQResume           = "dq.resume"
	TypeDQRequestVerify    = "dq.request_verify"
	TypeConfirm            = "confirm"
	TypeReaffirm           = "reaffirm"
	TypeCheckpoint         = "checkpoint"
	TypeEdgeAdd            = "edge.add"
	TypeEdgeRemove         = "edge.remove"
)

// Node kinds.
const (
	KindGoal           = "Goal"
	KindDesignQuestion = "DesignQuestion"
	KindHypothesis     = "Hypothesis"
	KindExperiment     = "Experiment"
	KindResult         = "Result"
)

// Edge types.
const (
	EdgeDependsOn   = "depends_on"
	EdgeSupersedes  = "supersedes"
	EdgeExternalRef = "external_ref"
)

// Result verdicts.
const (
	VerdictSupports     = "supports"
	VerdictContradicts  = "contradicts"
	VerdictInconclusive = "inconclusive"
)

// Confirm roles and confirm verdicts.
const (
	RoleVerifier = "verifier"
	RoleOwner    = "owner"

	ConfirmSufficient   = "sufficient"
	ConfirmInsufficient = "insufficient"
)

// RecipeNone marks an Experiment that follows no recipe.
const RecipeNone = "none"

// DecidedReading is the one value of NodeAdd.DecidedBy: the Experiment's
// verdicts are decided by reading, not by a measure.
const DecidedReading = "reading"

var (
	EventTypes = set(TypeNodeAdd, TypeResultRecord, TypeResultRetract, TypeHypothesisAbandon, TypeHypothesisAmend,
		TypeDQSelect, TypeDQUnselect, TypeDQDefer, TypeDQWithdraw, TypeDQResume, TypeDQRequestVerify,
		TypeExperimentWithdraw, TypeConfirm, TypeReaffirm, TypeCheckpoint, TypeEdgeAdd, TypeEdgeRemove)
	Kinds           = set(KindGoal, KindDesignQuestion, KindHypothesis, KindExperiment, KindResult)
	EdgeTypes       = set(EdgeDependsOn, EdgeSupersedes, EdgeExternalRef)
	Verdicts        = set(VerdictSupports, VerdictContradicts, VerdictInconclusive)
	Roles           = set(RoleVerifier, RoleOwner)
	ConfirmVerdicts = set(ConfirmSufficient, ConfirmInsufficient)
)

// Content limits, counted in runes.
const (
	MaxTitle        = 120
	MaxText         = 600
	MaxGoalBody     = 2000
	MaxWhy          = 300
	MaxEvidence     = 10
	MaxEvidenceItem = 300
	MaxExternalRef  = 300
	MaxLimits       = 10
	MaxLimitItem    = 300
)

// NodeAdd is the payload of node.add.
type NodeAdd struct {
	Kind       string   `json:"kind"`
	Parent     string   `json:"parent"`
	Title      string   `json:"title"`
	Body       string   `json:"body,omitempty"`
	Alias      string   `json:"alias,omitempty"`
	Recipe     string   `json:"recipe,omitempty"`
	RecipeHash string   `json:"recipe_hash,omitempty"`
	Targets    []string `json:"targets,omitempty"`
	DecidedBy  string   `json:"decided_by,omitempty"`
}

// ResultRecord is the payload of result.record.
type ResultRecord struct {
	Parent      string            `json:"parent"`
	Verdicts    map[string]string `json:"verdicts"`
	Observation string            `json:"observation"`
	Evidence    []string          `json:"evidence,omitempty"`
	// EvidenceHashes maps an evidence pointer into a versioned recipe to the
	// bundle hash recorded for that recipe when the Result was written.
	EvidenceHashes map[string]string `json:"evidence_hashes,omitempty"`
	Limits         []string          `json:"limits,omitempty"`
	Closes         []string          `json:"closes,omitempty"`
	// Readings states how the run read a line of the Goal that a verdict rests on.
	Readings []string `json:"readings,omitempty"`
}

// EvidenceRef returns the recipe ref (name@version) that an evidence pointer
// recipes/<name>@<version>/<file> names, after cleaning the path. ok is false
// for any other pointer.
func EvidenceRef(e string) (Ref, bool) {
	ref, err := ParseRef(evidenceDir(e))
	if err != nil || !ref.Versioned() {
		return Ref{}, false
	}
	return ref, true
}

// BadEvidenceRef returns the parse error for an evidence pointer
// recipes/<dir>/... whose dir holds an @ but is not a ref name@version, and
// nil for any other pointer.
func BadEvidenceRef(e string) error {
	dir := evidenceDir(e)
	if !MeantAsRef(dir) {
		return nil
	}
	_, err := ParseRef(dir)
	return err
}

// evidenceDir returns the folder under recipes/ that a cleaned evidence
// pointer names, or "" for a pointer outside recipes/.
func evidenceDir(e string) string {
	rest, found := strings.CutPrefix(path.Clean(e), "recipes/")
	if !found {
		return ""
	}
	dir, _, _ := strings.Cut(rest, "/")
	return dir
}

// HypothesisAmend is the payload of hypothesis.amend: the Hypothesis's
// revised body, which replaces the current one.
type HypothesisAmend struct {
	Body string `json:"body"`
}

// Empty is the payload of result.retract, hypothesis.abandon, dq.defer,
// dq.withdraw, dq.resume and reaffirm.
type Empty struct{}

// DQSelect is the payload of dq.select.
type DQSelect struct {
	Hypothesis  string `json:"hypothesis"`
	PrevSelect  string `json:"prev_select"`
	Tradeoff    bool   `json:"tradeoff"`
	ScopeChange bool   `json:"scope_change"`
}

// DQUnselect is the payload of dq.unselect.
type DQUnselect struct {
	PrevSelect string `json:"prev_select"`
}

// DQRequestVerify is the payload of dq.request_verify: the select whose
// insufficient verifier verdict has been addressed.
type DQRequestVerify struct {
	Select string `json:"select"`
}

// Confirm is the payload of confirm.
type Confirm struct {
	Select  string   `json:"select"`
	Role    string   `json:"role"`
	Verdict string   `json:"verdict"`
	Accepts []string `json:"accepts,omitempty"`
}

// Checkpoint is the payload of checkpoint.
type Checkpoint struct {
	Summary string `json:"summary"`
	// Role is RoleOwner on a checkpoint that closes readings, and empty otherwise.
	Role string `json:"role,omitempty"`
	// Closes lists the reading keys the owner's summary answers.
	Closes []string `json:"closes,omitempty"`
}

// Edge is the payload of edge.add and edge.remove.
type Edge struct {
	Type string `json:"type"`
	To   string `json:"to"`
}

// DecodePayload decodes e.Payload into dst, refusing unknown fields. A
// missing or null payload decodes as {}.
func DecodePayload(e Event, dst any) error {
	raw := bytes.TrimSpace(e.Payload)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		raw = []byte("{}")
	}
	return strictDecode(raw, dst)
}

// VerdictOrder lists a result.record's verdict targets in the order the
// payload stores them; decoding into a map would lose that order.
func VerdictOrder(raw json.RawMessage) []string {
	var top map[string]json.RawMessage
	if json.Unmarshal(raw, &top) != nil {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(top["verdicts"]))
	if _, err := dec.Token(); err != nil {
		return nil
	}
	var keys []string
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return keys
		}
		keys = append(keys, t.(string))
		var skip json.RawMessage
		if dec.Decode(&skip) != nil {
			return keys
		}
	}
	return keys
}

// ParseLine decodes one JSONL line, refusing unknown top-level keys.
func ParseLine(line []byte) (Event, error) {
	var e Event
	err := strictDecode(line, &e)
	return e, err
}

func strictDecode(raw []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if dec.More() {
		return errors.New("trailing data after JSON object")
	}
	return nil
}

var kindLetters = map[string]string{
	KindGoal:           "G",
	KindDesignQuestion: "Q",
	KindHypothesis:     "H",
	KindExperiment:     "E",
	KindResult:         "R",
}

// DisplayID returns <letter>-<last 6 chars of id>, for example H-7QK2ZD.
func DisplayID(kind, id string) string {
	tail := id
	if len(tail) > 6 {
		tail = tail[len(tail)-6:]
	}
	letter, ok := kindLetters[kind]
	if !ok {
		letter = "?"
	}
	return fmt.Sprintf("%s-%s", letter, tail)
}

// LimitKey names limit n (counted from 1) of Result r, for example
// 01J…W6V06D#1.
func LimitKey(r string, n int) string { return fmt.Sprintf("%s#%d", r, n) }

// ParseLimitKey splits "<ref>#<n>" into ref and n; ok is false unless n >= 1.
func ParseLimitKey(k string) (ref string, n int, ok bool) {
	ref, num, found := strings.Cut(k, "#")
	n, err := strconv.Atoi(num)
	if !found || ref == "" || err != nil || n < 1 {
		return "", 0, false
	}
	return ref, n, true
}

// ReadingKey names reading n (counted from 1) of Result r, for example
// 01J…W6V06D#r1.
func ReadingKey(r string, n int) string { return fmt.Sprintf("%s#r%d", r, n) }

// ParseReadingKey splits "<ref>#r<n>" into ref and n; ok is false unless n >= 1.
func ParseReadingKey(k string) (ref string, n int, ok bool) {
	ref, num, found := strings.Cut(k, "#r")
	n, err := strconv.Atoi(num)
	if !found || ref == "" || err != nil || n < 1 {
		return "", 0, false
	}
	return ref, n, true
}

// Runes counts runes; every content limit is measured with it.
func Runes(s string) int { return utf8.RuneCountInString(s) }

func set(items ...string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, it := range items {
		m[it] = true
	}
	return m
}
