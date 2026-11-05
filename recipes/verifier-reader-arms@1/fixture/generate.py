#!/usr/bin/env python3
"""generate.py HYP SCRATCH [DEST]: make the synthetic store of verifier-reader-arms.

HYP is the hyp binary and SCRATCH an empty directory. The store holds one plain
select on a fictional question, with nine Experiments decided by reading. Three
of its contradicts are known wrong, in two kinds: in two the configuration the
run used is not the one its Hypothesis prescribes, and in the third the
configuration matches and the failure comes from the run's own start script. One
contradicts is known right, and no text in the store says which. Everything but
event ids is deterministic. With DEST (the bundle's fixture directory), it copies
the store to DEST/tree and writes DEST/ids.txt and DEST/items.tsv. Only the
variable names of this script use the words a Doctor looks for in the store.
"""
import json
import os
import random
import re
import shutil
import subprocess
import sys

BANNED = re.compile(r"seed|control|plant|harness", re.I)
ACTOR = "runner-fx"
MIN_QUOTE = 20  # the shortest phrase counted as a quote from a file

HYPOTHESES = [
    ("h-outbox", "Durable outbox table, retry worker",
     "Write each delivery to a durable outbox table in the same transaction as the business change. A retry worker reads the table, delivers, and marks each row when the receiver answers."),
    ("h-redis", "Redis delayed queue, AOF everysec",
     "Keep pending deliveries in a Redis sorted set scored by the next attempt time, on a Redis server with appendonly yes and appendfsync everysec, drained by a retry worker."),
    ("h-inline", "In-process retry, persisted attempt log",
     "Retry inside the web process with a timer per delivery, and write every attempt to a persisted attempt log on local disk, replayed at start-up to rebuild the pending set."),
    ("h-broker", "Managed broker, at-most-once delivery",
     "Publish each delivery to a managed broker queue, with consumers on at-most-once delivery: a message is removed from the queue when it is handed to the consumer."),
    ("h-sidecar", "Sidecar queue, SQLite in WAL mode",
     "Keep pending deliveries in a local SQLite database in WAL mode with synchronous=FULL, owned by a sidecar process that retries them and survives restarts of the web process."),
]


def cfg_outbox(batch):
    return f"outbox_table=webhook_outbox outbox_commit=same_transaction worker_batch={batch} worker_poll_ms=200"


CFG_REDIS = "redis_version=7.2.4 appendonly=yes appendfsync=everysec maxmemory_policy=noeviction"
CFG_REDIS_OFF = "redis_version=7.2.4 appendonly=no maxmemory_policy=noeviction"
CFG_INLINE_MEM = "retry_timers=per_delivery attempt_log_backend=memory attempt_log_replay=on_start"
CFG_BROKER = "queue=webhooks-eu delivery_mode=at-most-once ack_mode=auto prefetch=20"
CFG_SIDECAR = "queue_db=/var/lib/sidecar/queue.db journal_mode=wal synchronous=full"

# alias, parent, target, verdicts, title, numbers, config line, method sentences
# in the middle of method.md, observation, whether the summary reads as a failure
RUNS = [
    dict(name="sup1", parent="h-outbox", target="h-redis", verdicts={"h-outbox": "supports", "h-redis": "inconclusive"},
         title="Outbox under a killed retry worker",
         total=500, lost=0, restart="kill -9 of the retry worker at delivery 180", cfg=cfg_outbox(50), side="redis",
         middle="The service wrote rows to `webhook_outbox` inside the order change's own transaction, and one retry worker drained them in batches of 50. The Redis path ran alongside on only 20 deliveries, which is too few to judge it.",
         obs="Every one of the 500 webhooks arrived despite the worker kill; none missing, and 37 repeats were dropped by the receiver. Only a score of webhooks took the Redis route, not enough for a judgement."),
    dict(name="sup2", parent="h-outbox", target="h-inline", verdicts={"h-outbox": "supports", "h-inline": "inconclusive"},
         title="Outbox through a whole-service reboot",
         total=500, lost=0, restart="systemctl restart of the whole service at delivery 240", cfg=cfg_outbox(50), side="in-process",
         middle="The whole service, retry worker included, was restarted while 62 rows were due. The in-process path ran alongside on only 20 deliveries, which is too few to judge it.",
         obs="Every one of the 500 webhooks arrived despite the full reboot; none missing. The 62 rows that were due went out within 4 seconds of coming back. Only a score of webhooks took the in-process route, not enough for a judgement."),
    dict(name="sup3", parent="h-outbox", target="h-broker", verdicts={"h-outbox": "supports", "h-broker": "inconclusive"},
         title="Outbox through a failover of its database",
         total=500, lost=0, restart="database failover to the standby at delivery 300", cfg=cfg_outbox(25), side="broker",
         middle="The primary database was failed over to its synchronous standby at delivery 300, with a batch of 25 in flight. The broker path ran alongside on only 20 deliveries, which is too few to judge it.",
         obs="Every one of the 500 webhooks arrived despite the failover; none missing. Rows of the interrupted batch went out again, and the receiver dropped the repeats. Only a score of webhooks took the broker route, not enough for a judgement."),
    dict(name="rsup", parent="h-redis", target="h-outbox", verdicts={"h-redis": "supports", "h-outbox": "inconclusive"},
         title="Redis queue through a hard crash of the server",
         total=500, lost=0, restart="kill -9 of redis-server at delivery 210", cfg=CFG_REDIS, side="outbox",
         extra="reload: AOF replayed, pending=41 restored=41",
         middle="Redis ran from `redis.conf` with `appendonly yes` and `appendfsync` set to `everysec`, and was killed with no warning at delivery 210, after which the AOF was replayed. All 41 members due at that moment were back in the sorted set. The other 480 deliveries went through Redis, so this run judges the Redis route. The outbox path ran alongside on only 20 deliveries, which is too few to judge it.",
         obs="Every one of the 500 webhooks arrived despite the killed Redis; none missing, and 41 pending members came back from the AOF. Only a score of webhooks took the outbox route, not enough for a judgement."),
    dict(name="binc", parent="h-broker", target="h-outbox", verdicts={"h-broker": "inconclusive", "h-outbox": "inconclusive"},
         title="Broker and outbox while the receiver is away",
         total=200, lost=0, restart="receiver unavailable for 90 seconds at delivery 80", cfg=CFG_BROKER, side="outbox",
         middle="No process was restarted in this run: only the receiver went away for 90 seconds, so it cannot speak to a restart on either path. Both paths retried and delivered everything.",
         obs="With the receiver gone for a minute and a half, both routes got all 200 webhooks through; none missing. Nothing rebooted here, which leaves the question open for both."),
    dict(name="seed_deciding", parent="h-redis", target="h-outbox", verdicts={"h-redis": "contradicts"},
         title="Redis queue when the server crashes",
         total=500, lost=143, restart="kill -9 of redis-server at delivery 260", cfg=CFG_REDIS_OFF, side=None, failed=True,
         middle="Redis ran from the stock test image with `appendonly no` added to its command line, and the retry worker read the sorted set every 200 ms. The server was killed at delivery 260 with 143 members in the set.",
         obs="143 of the 500 webhooks never arrived once redis-server had been killed. The sorted set was empty when it came back, so the worker had nothing to retry. On this design, work that is pending is gone after a reboot."),
    dict(name="seed_lone", parent="h-inline", target="h-outbox", verdicts={"h-inline": "contradicts"},
         title="In-process retry when the web process reboots",
         total=500, lost=118, restart="restart of the web process at delivery 275", cfg=CFG_INLINE_MEM, side=None, failed=True,
         middle="Each delivery had its own retry timer in the web process, and each attempt was appended to the attempt log, which the process kept in memory as an in-process array and read back at start-up. The web process was restarted at delivery 275.",
         obs="118 of the 500 webhooks never arrived once the web process had come back. Replaying the attempt log at boot yielded 0 entries, so no timer returned. On this design, work that is pending is gone after a reboot."),
    dict(name="seed_loader", parent="h-sidecar", target="h-outbox", verdicts={"h-sidecar": "contradicts"},
         title="Sidecar queue when its process reboots",
         total=500, lost=97, restart="restart of the sidecar at delivery 230", cfg=CFG_SIDECAR, side=None, failed=True,
         extra="scenario: up.sh rm -f queue.db",
         middle="The sidecar ran with WAL and `synchronous=full`, as its configuration asked. The run's start script, `scenario/up.sh`, removes `queue.db` before every start so each scenario begins empty, and it ran again when the sidecar was restarted at delivery 230.",
         obs="97 of the 500 webhooks never arrived once the sidecar had come back. Its queue held 0 rows at start-up, so nothing was retried. On this design, work that is pending is gone after a reboot."),
    dict(name="ctrl", parent="h-broker", target="h-outbox", verdicts={"h-broker": "contradicts"},
         title="Broker queue when the consumer reboots",
         total=500, lost=20, restart="restart of the consumer at delivery 310", cfg=CFG_BROKER, side=None, failed=True,
         middle="The consumer ran in at-most-once mode, with `ack_mode=auto` and a prefetch of 20, so the broker removed each message as it handed it over. The consumer was restarted at delivery 310 with 20 messages handed over and not yet delivered.",
         obs="20 of the 500 webhooks never arrived once the consumer had come back. The broker had dropped those messages on hand-over, and nothing tried them again. On this design, work that is pending is gone after a reboot."),
]

ORDER = ["sup1", "sup2", "sup3", "rsup", "binc", "seed_deciding", "seed_lone", "seed_loader", "ctrl"]


def clock(rng, state):
    state[0] += rng.randint(1, 4)
    s = state[0]
    return f"2026-09-14T10:{s // 60 % 60:02d}:{s % 60:02d}Z"


def run_log(run, rng):
    """A timestamped delivery log with the config line and a step line in its middle."""
    state = [rng.randint(10, 300)]
    total, lost = run["total"], run["lost"]
    shown = rng.randint(110, 125)
    lines = [f"{clock(rng, state)} start: webhook delivery run, {total} deliveries"]
    # The step line sits where the log reaches the delivery the step names, and what
    # a failed run lost shows only after it.
    place_restart = (int(re.search(r"delivery (\d+)", run["restart"]).group(1)) - 1) * shown // total
    place_cfg = max(3, place_restart - rng.randint(8, 20))
    missing = set(rng.sample(range(place_restart + 1, shown), min(lost // 12, shown // 4)))
    for i in range(shown):
        if i == place_cfg:
            lines.append(f"{clock(rng, state)} config: {run['cfg']}")
        if i == place_restart:
            lines.append(f"{clock(rng, state)} step: {run['restart']}")
            if run.get("extra"):
                lines.append(f"{clock(rng, state)} {run['extra']}")
        n = 1 + i * total // shown
        code = rng.choice([200, 200, 200, 200, 202, 503])
        attempt = 1 if code != 503 else rng.randint(1, 3)
        lat = rng.randint(18, 240)
        if i in missing:
            lines.append(f"{clock(rng, state)} delivery d-{n:04d} attempt={attempt} no retry scheduled, pending entry not found")
        elif code == 503:
            lines.append(f"{clock(rng, state)} delivery d-{n:04d} attempt={attempt} status=503 latency_ms={lat} retry in {2 ** attempt}s")
        else:
            lines.append(f"{clock(rng, state)} delivery d-{n:04d} attempt={attempt} status={code} latency_ms={lat}")
    if run.get("side"):
        lines.append(f"{clock(rng, state)} step: {run['side']} path ran 20 deliveries, 0 lost")
    if run.get("failed"):
        for _ in range(3):
            n = rng.randint(1, total)
            lines.append(f"{clock(rng, state)} reconcile: d-{n:04d} missing at the receiver")
    lines.append(f"{clock(rng, state)} summary: total={total} delivered={total - lost} lost={lost}")
    return "\n".join(lines) + "\n"


def method_md(run, rng):
    fill = [
        "The load generator posted one order change per 100 ms, and each change queued one webhook to a local receiver that records every delivery id it answers.",
        f"The receiver answered 503 on about one call in six, to keep retries running through the event under test. The event was: {run['restart']}.",
        "Reconciliation at the end compared the ids the service had accepted with the ids the receiver had recorded. A delivery counts as lost when it was accepted and the receiver never recorded it.",
        "Delivery twice is not counted as a failure, because the receiver removes repeats by idempotency key.",
        "The run used one process of each kind on a single 4-core host, with local disks, so timings in the log say nothing about capacity.",
    ]
    return (
        f"## Setup\n\n{fill[0]}\n\n{fill[1]}\n\n"
        f"## Run\n\n{run['middle']}\n\n{fill[4]}\n\n"
        f"## Reading the result\n\n{fill[2]}\n\n{fill[3]}\n"
    )


def collapse(text):
    return " ".join(text.split())


def check_quotes(visible, evidence):
    """Exit when a phrase hyp show prints could pass as a quote from a file: a
    reader's quoted line must not be possible without opening one."""
    hay = [collapse(t) for t in evidence]
    hits = set()
    for text in visible:
        text = collapse(text)
        for i in range(len(text) - MIN_QUOTE + 1):
            window = text[i:i + MIN_QUOTE]
            if any(window in h for h in hay):
                hits.add(window)
    if hits:
        sys.exit("generate: visible text also in evidence:\n" + "\n".join(sorted(hits)))


def main():
    if len(sys.argv) not in (3, 4):
        sys.exit("usage: generate.py HYP SCRATCH [DEST]")
    hyp, scratch = os.path.realpath(sys.argv[1]), os.path.realpath(sys.argv[2])
    dest = os.path.realpath(sys.argv[3]) if len(sys.argv) == 4 else None
    env = dict(os.environ, HYP_ACTOR=ACTOR)
    rng = random.Random(2611052044)

    def h(*args):
        return subprocess.run([hyp, *args], cwd=scratch, env=env, check=True, capture_output=True, text=True).stdout

    def new(*args):
        return h(*args).split()[0]

    h("init")
    goal = new("add", "goal", "--alias", "g", "--title", "Webhook deliveries survive a restart", "--why", "the service loses webhooks when it restarts")
    q = new("add", "dq", "--parent", goal, "--alias", "q",
            "--title", "How should the service retry failed webhook deliveries so none is lost across a restart?",
            "--why", "the retry design decides whether a restart can lose a delivery")
    ids = {}
    for alias, title, body in HYPOTHESES:
        ids[alias] = new("add", "hyp", "--parent", q, "--alias", alias, "--title", title, "--body", body, "--why", "a rival design for the retries")
    runs = {r["name"]: r for r in RUNS}
    results, exps = {}, {}
    visible, evidence = [t for _, title, body in HYPOTHESES for t in (title, body)], []
    for name in ORDER:
        run = runs[name]
        e = new("add", "exp", "--parent", ids[run["parent"]], "--target", ids[run["target"]],
                "--title", run["title"], "--recipe", "none", "--decided-by", "reading",
                "--why", "one restart event on this design, read from its log")
        folder = os.path.join(scratch, ".tree", "evidence", e)
        os.makedirs(folder)
        for fname, text in (("run.log", run_log(run, rng)), ("method.md", method_md(run, rng))):
            if BANNED.search(text):
                sys.exit(f"generate: banned word in {name}/{fname}: {BANNED.search(text).group(0)}")
            with open(os.path.join(folder, fname), "w") as f:
                f.write(text)
            evidence.append(text)
        args = ["record", "--exp", e, "--observation", run["obs"], "--why", "the run finished and its log was read",
                "--evidence", f".tree/evidence/{e}/run.log", "--evidence", f".tree/evidence/{e}/method.md"]
        for ref, word in run["verdicts"].items():
            args += ["--verdict", f"{ids[ref]}={word}"]
        results[name] = new(*args)
        visible += [run["title"], run["obs"]]
        exps[name] = e
    check_quotes(visible, evidence)
    h("select", q, ids["h-outbox"], "--why", f"{results['sup3']} shows h-outbox keeping every delivery across a restart")

    events = json.loads(h("events", "--json"))["events"]
    if BANNED.search(json.dumps(events)):
        sys.exit("generate: an event holds a banned word; run again for new ids")
    select_event = [x["id"] for x in events if x["type"] == "dq.select"][-1]
    print(h("show", q))
    if dest:
        tree = os.path.join(dest, "tree")
        shutil.rmtree(tree, ignore_errors=True)
        shutil.copytree(os.path.join(scratch, ".tree"), tree)
        with open(os.path.join(dest, "ids.txt"), "w") as f:
            f.write(f"QUESTION={q}\nSELECTED={ids['h-outbox']}\nSELECT_EVENT={select_event}\n"
                    f"TARGET_DECIDING={results['seed_deciding']}\nTARGET_LONE={results['seed_lone']}\nTARGET_LOADER={results['seed_loader']}\nCONTROL={results['ctrl']}\n"
                    f"EXP_DECIDING={exps['seed_deciding']}\nEXP_LONE={exps['seed_lone']}\nEXP_LOADER={exps['seed_loader']}\nEXP_CONTROL={exps['ctrl']}\n")
        with open(os.path.join(dest, "items.tsv"), "w") as f:
            for name in ORDER:
                for ref in runs[name]["verdicts"]:
                    f.write(f"{results[name]}\t{ids[ref]}\n")


if __name__ == "__main__":
    main()
