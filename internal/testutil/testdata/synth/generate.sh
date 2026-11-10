#!/bin/sh
# generate.sh HYP: rebuild the synthetic store in this directory with the hyp
# binary HYP. Every Experiment declares decided_by or pins the instrument
# recipes/count-check@1, so the store has the shape #230 and later write.
#
# Event ids are new on every run, and tests name them (the display ids and the
# cut events), so a regenerated store means updating those tests. The ids are
# fixed once the store is committed.
#
# The committed store is history: it was written by a hyp from before the
# select became the gate (ADR 0015). A current hyp refuses its verifier
# confirms (the role is retired) and the q-cache select (h-none and h-cdn are
# undecided), so this script no longer runs to the end; replay reads every
# event it wrote, and the verifier confirms apply nothing.
#
# Rounds, in order, and what hyp signals reports for them:
#   e-one     targets none: one-rival
#   e-redis   targets h-none and gives it inconclusive; a Result that leaves a
#             target without a verdict is refused (ADR 0018), so the unjudged
#             mark it once showed is gone
#   e-split   moves h-none and h-cdn to inconclusive
#   e-lru-2..4  move nothing; progress is the ledger's, so no signal (ADR 0017)
#   q-cache   two insufficient verifier verdicts, once verifier-loop. Two
#             dq.request_verify events, which older binaries wrote and replay
#             reads as history, keep that old shape in the log
#   e-short   pins count-check@1 and cites its command record
#   short     amended after its select, then confirmed again
set -eu
hyp=${1:?usage: generate.sh HYP}
case $hyp in /*) ;; *) hyp=$(pwd)/$hyp ;; esac
here=$(cd "$(dirname "$0")" && pwd)
cd "$here"
rm -rf .tree
"$hyp" init >/dev/null

h() { "$hyp" "$@" >/dev/null; }
run() { HYP_ACTOR=runner h "$@"; }

# old_request WHY appends the dq.request_verify an older hyp wrote on the last
# select, since no command writes one now.
old_request() {
	python3 - "$1" <<'PY'
import datetime, json, os, sys
path = ".tree/events/runner.jsonl"
sel = [e for e in map(json.loads, open(path)) if e["type"] == "dq.select"][-1]
now = datetime.datetime.now(datetime.timezone.utc)
abc = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
n = (int(now.timestamp() * 1000) << 80) | int.from_bytes(os.urandom(10), "big")
ulid = "".join(abc[(n >> (5 * i)) & 31] for i in reversed(range(26)))
ev = {"v": 1, "id": ulid, "ts": now.strftime("%Y-%m-%dT%H:%M:%S.%fZ"), "actor": "runner", "type": "dq.request_verify",
      "node": sel["node"], "payload": {"select": sel["id"]}, "why": sys.argv[1]}
open(path, "a").write(json.dumps(ev, separators=(",", ":"), ensure_ascii=False) + "\n")
PY
}

run add goal --title "Pick a cache design for the synthetic service" --alias g --why "the service is slow under load"
run add dq --parent g --title "Which cache layer should the service use?" --alias q-cache --why "the design question"
run add hyp --parent q-cache --title "A process-local LRU cache" --alias h-lru \
	--body "Each process keeps an LRU of 10k entries; no shared state." --why "cheapest rival"
run add hyp --parent q-cache --title "A shared Redis cache" --alias h-redis --why "a shared rival"
run add hyp --parent q-cache --title "No cache; tune the queries" --alias h-none --why "the null rival"
run add hyp --parent q-cache --title "A CDN in front of the reads" --alias h-cdn --why "an edge rival"

run add exp --parent h-lru --title "Load test the LRU alone" --alias e-one --recipe none --decided-by reading \
	--target none --why "a first look"
run record --exp e-one --verdict h-lru=supports --observation "p99 fell from 900ms to 120ms" --why "read from the load report"

run add exp --parent h-redis --title "Load test Redis against no cache" --alias e-redis --recipe none --decided-by reading \
	--target h-none --why "two rivals in one run"
run record --exp e-redis --verdict h-redis=contradicts --verdict h-none=inconclusive --observation "the network hop cost more than the cache saved" \
	--why "h-none was not read"

run add exp --parent h-none --title "Compare tuned queries with a CDN" --alias e-split --recipe none --decided-by reading \
	--target h-cdn --why "the two rivals left"
run record --exp e-split --verdict h-none=inconclusive --verdict h-cdn=inconclusive \
	--observation "both within noise of the baseline" --why "read from the load report"

for n in 2 3 4; do
	run add exp --parent h-lru --title "Load test the LRU again, run $n" --alias e-lru-$n --recipe none --decided-by reading \
		--target h-cdn --why "repeat the LRU run"
	run record --exp e-lru-$n --verdict h-lru=supports --verdict h-cdn=inconclusive \
		--observation "p99 near 120ms again" --why "read from the load report"
done

run select q-cache h-lru --why "h-lru is the one supported rival"
HYP_ACTOR=verifier-1 h confirm q-cache --role verifier --verdict insufficient --why "no run measured memory per process"
old_request "asked again"
HYP_ACTOR=verifier-2 h confirm q-cache --role verifier --verdict insufficient --why "still no memory measurement"
run add exp --parent h-lru --title "Measure memory per process with the LRU" --alias e-mem --recipe none --decided-by reading \
	--target h-cdn --why "the verifiers' gap"
run record --exp e-mem --verdict h-lru=supports --verdict h-cdn=inconclusive \
	--observation "the LRU adds 40MB per process" --limit "one host size only" --why "read from the profiler"
old_request "memory is measured now"
HYP_ACTOR=verifier-3 h confirm q-cache --role verifier --verdict sufficient --why "latency and memory are both measured"

run add dq --parent g --title "Is the hot key set short?" --alias q-keys --why "sizing the LRU"
run add hyp --parent q-keys --title "Two or fewer hot keys per tenant" --alias short --why "the short rival"
run add hyp --parent q-keys --title "More than two hot keys per tenant" --alias long --why "the long rival"
run add exp --parent short --title "Judge the measured hot key count" --alias e-short --recipe count-check@1 \
	--target long --why "count-check decides it"
exp=$("$hyp" show e-short | sed -n 's/^Experiment \(E-[A-Z0-9]*\)$/\1/p')
test -n "$exp"
"$hyp" recipe run --record ".tree/evidence/$exp/judge" count-check@1 judge 2 >/dev/null
run record --exp e-short --verdict short=supports --verdict long=contradicts --observation "count 2" \
	--evidence recipes/count-check@1/judge.sh --evidence ".tree/evidence/$exp/judge/command.json" \
	--why "count-check printed verdict: short=supports"
run select q-keys short --why "the instrument decided it"
HYP_ACTOR=verifier-1 h confirm q-keys --role verifier --verdict sufficient --why "the command record shows the gate passed"
run amend short --body "At most two keys per tenant are hot; count-check@1 judged a count of 2." --why "state the measured count"
HYP_ACTOR=verifier-1 h confirm q-keys --role verifier --verdict sufficient --why "the amended body states the count the record shows"

run checkpoint --summary "Both questions answered: h-lru for the cache, short for the hot keys" --why "end of the synthetic run"
