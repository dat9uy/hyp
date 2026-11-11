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
	TypeHypothesisSubject  = "hypothesis.subject"
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
	TypeAnswer             = "answer"
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
	RoleVerifier = "verifier" // retired by ADR 0015: only old logs hold it, and it applies nothing
	RoleOwner    = "owner"

	ConfirmSufficient   = "sufficient"
	ConfirmInsufficient = "insufficient"
)

// RecipeNone marks an Experiment that follows no recipe.
const RecipeNone = "none"

// DecidedReading is the one value of NodeAdd.DecidedBy, an older
// spelling of RubricReading: replay reads it as rubric reading, and no
// command writes it (ADR 0018).
const DecidedReading = "reading"

// RubricReading is the NodeAdd.Rubric value that says the Experiment's
// verdicts are decided by reading the prose rubric in its body (ADR 0018).
const RubricReading = "reading"

// MaxRecords bounds an Experiment's pinned records.
const MaxRecords = 500

var (
	EventTypes = set(TypeNodeAdd, TypeResultRecord, TypeResultRetract, TypeHypothesisAbandon, TypeHypothesisAmend, TypeHypothesisSubject,
		TypeDQSelect, TypeDQUnselect, TypeDQDefer, TypeDQWithdraw, TypeDQResume, TypeDQRequestVerify,
		TypeExperimentWithdraw, TypeConfirm, TypeReaffirm, TypeCheckpoint, TypeAnswer, TypeEdgeAdd, TypeEdgeRemove)
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
	MaxReadingItem  = 600
	MaxSubjectFiles = 50
	MaxFingerprints = 2000
	MaxRunSubjects  = 100 // a Result's subjects: one per cited run, and a sandbox record holds one run per Arm
	MaxLine         = 600 // a line on a Goal's node.add (ADR 0020); a line an old checkpoint wrote holds at most MaxLimitItem
	MaxCites        = 10
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
	// Rubric is how the Experiment's measures become verdicts: a rubric
	// bundle's NAME@VERSION, pinned with RubricHash, or RubricReading.
	Rubric     string `json:"rubric,omitempty"`
	RubricHash string `json:"rubric_hash,omitempty"`
	// Records pins the earlier records the Experiment reads beyond its own
	// run, by path and sha256 (ADR 0018).
	Records []SubjectFile `json:"records,omitempty"`
	// Aims lists the limit and reading keys an Experiment means to settle.
	// It is intent: only a Result's closes or applies settles one.
	Aims []string `json:"aims,omitempty"`
	// Lines are a Goal's lines, the cases every candidate must meet, keyed
	// <goal-id>#<n>. They come with the Goal's add and no later event adds
	// one (ADR 0020).
	Lines []GoalLine `json:"lines,omitempty"`
}

// GoalLine is one line of a Goal. Cites are plain keys, each settled when
// the line is written: a line, an accepted or closed limit, an answered
// reading, or a selected or contradicted Hypothesis. What a cite means is
// derived from the key's state at the line's place in the log, never stored.
type GoalLine struct {
	Text  string   `json:"text"`
	Cites []string `json:"cites,omitempty"`
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
	// Readings states how the run read owner words a verdict rests on: a line
	// of the Goal, or an answer the owner gave to an earlier reading (Of).
	Readings []ReadingItem `json:"readings,omitempty"`
	// Applies lists the closed readings whose owner answers this Result's
	// verdicts apply. Each must be one its Experiment aims at.
	Applies []string `json:"applies,omitempty"`
	// Subjects keeps one subject per cited run: the base commit and the
	// sha256 of each file its cited command or sandbox record says the run
	// was given over it (ADR 0020). hyp record copies it from those records,
	// as it copies EvidenceHashes; no runner types it.
	Subjects []RunSubject `json:"subjects,omitempty"`
	// Fingerprints is an old Result's one merged set of file sha256s. No
	// command writes it; replay reads it as one subject with no base.
	Fingerprints []string `json:"fingerprints,omitempty"`
}

// RunSubject is the subject of one cited run: the base commit it ran at, ""
// when its record names none, and the sorted, unique sha256s of the files
// given over that base (a patch and input files).
type RunSubject struct {
	Base  string   `json:"base,omitempty"`
	Files []string `json:"files,omitempty"`
}

// IsCommit reports whether s is a full git object id: 40 or 64 lowercase
// hex digits.
func IsCommit(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// HypothesisSubject is the payload of hypothesis.subject: what embodies the
// Hypothesis's claim, a base commit, files named by their sha256, or both
// (ADR 0020). Path is a pointer for readers, relative to the store root; the
// sha256 is the identity.
type HypothesisSubject struct {
	Base  string        `json:"base,omitempty"`
	Files []SubjectFile `json:"files,omitempty"`
}

// SubjectFile is one file of a Hypothesis's subject.
type SubjectFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// IsSHA256 reports whether s is 64 lowercase hex digits.
func IsSHA256(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
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
	Hypothesis string `json:"hypothesis"`
	PrevSelect string `json:"prev_select"`
	// Tradeoff is kept so that older events decode; nothing reads it. A
	// trade-off is derived from the question's candidates.
	Tradeoff    bool `json:"tradeoff,omitempty"`
	ScopeChange bool `json:"scope_change"`
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

// Checkpoint is the payload of checkpoint: the runner's report on its Goal.
// It holds a summary, refers, or both (ADR 0020). Role, Closes, Revises,
// Accepts, Returns and Lines are read-only history: replay applies them as
// an old owner checkpoint wrote them, and no command writes them (owner
// decision of 2026-11-11); the owner's acts are an Answer.
type Checkpoint struct {
	Summary string `json:"summary,omitempty"`
	// Role is RoleOwner on an old checkpoint that closes or revises readings,
	// accepts limits or writes lines, and empty otherwise.
	Role string `json:"role,omitempty"`
	// Closes lists the open reading keys the owner's summary answers.
	Closes []string `json:"closes,omitempty"`
	// Revises lists the answered reading keys whose answer the owner's
	// summary replaces from this checkpoint on.
	Revises []string `json:"revises,omitempty"`
	// Accepts lists the limit keys the owner accepts untested.
	Accepts []string `json:"accepts,omitempty"`
	// Refers lists the limit keys the runner hands to the owner, because no
	// test within the Goal's lines can settle them, each with why. An old
	// checkpoint stored each as a bare key, and its Summary says why.
	Refers []KeyText `json:"refers,omitempty"`
	// Returns lists the referred or accepted limit keys the owner hands
	// back to the runner, because the Goal covers their case; Summary says
	// why. It is the mirror of Refers.
	Returns []string `json:"returns,omitempty"`
	// Lines lists an old Goal's lines, keyed <goal-id>#<n>. Only the first
	// owner checkpoint on a Goal that carries any wrote them; a Goal's lines
	// now come with its add (NodeAdd.Lines).
	Lines []string `json:"lines,omitempty"`
}

// Answer is the payload of answer, the owner's act (ADR 0020). It has no
// node, no summary, no why and no role: the owner's reason is each item's
// text. Answers settles an open reading or replaces an earlier answer of an
// answered one; Accepts accepts a referred limit, so this Goal does not run
// its case; Returns hands a referred or accepted limit back to the runner.
type Answer struct {
	Answers []KeyText `json:"answers,omitempty"`
	Accepts []KeyText `json:"accepts,omitempty"`
	Returns []KeyText `json:"returns,omitempty"`
}

// KeyText is one keyed item of an answer or a refer: the key it moves and
// the writer's words for it. A bare string decodes as a key with no text,
// which is how an old checkpoint stored its refers.
type KeyText struct {
	Key  string `json:"key"`
	Text string `json:"text,omitempty"`
}

type keyText KeyText

func (k KeyText) MarshalJSON() ([]byte, error) {
	if k.Text == "" {
		return json.Marshal(k.Key)
	}
	return json.Marshal(keyText(k))
}

func (k *KeyText) UnmarshalJSON(b []byte) error {
	if t := bytes.TrimSpace(b); len(t) > 0 && t[0] == '"' {
		*k = KeyText{}
		return json.Unmarshal(t, &k.Key)
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	var p keyText
	if err := d.Decode(&p); err != nil {
		return err
	}
	*k = KeyText(p)
	return nil
}

// Keys lists the keys of items, in order.
func Keys(items []KeyText) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.Key
	}
	return out
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

// KeyKind is the kind of item a key names inside a Result.
type KeyKind string

const (
	KeyLimit   KeyKind = "limit"   // <result>#<n>
	KeyReading KeyKind = "reading" // <result>#r<n>
)

// Key names item N (counted from 1) of one kind in Result's list: a limit,
// 01J…W6V06D#1, or a reading, 01J…W6V06D#r1. A Goal's line is a limit
// keyed on the Goal, 01J…WH6GK5#1. Result is the full id of the node that
// holds the item once stored, or any node reference on the command line.
type Key struct {
	Result string
	Kind   KeyKind
	N      int
}

// LimitKey and ReadingKey spell the key of limit or reading n of Result r.
func LimitKey(r string, n int) string   { return Key{r, KeyLimit, n}.String() }
func ReadingKey(r string, n int) string { return Key{r, KeyReading, n}.String() }

func (k Key) String() string {
	if k.Kind == KeyReading {
		return fmt.Sprintf("%s#r%d", k.Result, k.N)
	}
	return fmt.Sprintf("%s#%d", k.Result, k.N)
}

// ParseKey splits "<ref>#<n>" or "<ref>#r<n>"; ok is false unless ref is not
// empty and n >= 1. The two forms never overlap.
func ParseKey(s string) (Key, bool) {
	ref, num, found := strings.Cut(s, "#")
	kind := KeyLimit
	if rest, isReading := strings.CutPrefix(num, "r"); isReading {
		kind, num = KeyReading, rest
	}
	n, err := strconv.Atoi(num)
	if !found || ref == "" || err != nil || n < 1 {
		return Key{}, false
	}
	return Key{ref, kind, n}, true
}

// KeyResult returns the node a stored key names: its Result, or the Goal
// for a line.
func KeyResult(s string) string {
	ref, _, _ := strings.Cut(s, "#")
	return ref
}

// DisplayKey shows key k of a node of kind as R-W6V06D#1, R-W6V06D#r1 or
// G-WH6GK5#1.
func DisplayKey(kind, k string) string {
	r := KeyResult(k)
	return DisplayID(kind, r) + k[len(r):]
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

// ReadingItem is one reading of a Result. Of names the reading whose owner
// answer this reading reads, or is empty when it reads a line of the Goal.
// A reading with no Of is stored as a bare string, so older events decode
// and encode unchanged.
type ReadingItem struct {
	Of   string `json:"of,omitempty"`
	Text string `json:"text"`
}

type readingItem ReadingItem

func (r ReadingItem) MarshalJSON() ([]byte, error) {
	if r.Of == "" {
		return json.Marshal(r.Text)
	}
	return json.Marshal(readingItem(r))
}

func (r *ReadingItem) UnmarshalJSON(b []byte) error {
	if t := bytes.TrimSpace(b); len(t) > 0 && t[0] == '"' {
		*r = ReadingItem{}
		return json.Unmarshal(t, &r.Text)
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	var p readingItem
	if err := d.Decode(&p); err != nil {
		return err
	}
	*r = ReadingItem(p)
	return nil
}
