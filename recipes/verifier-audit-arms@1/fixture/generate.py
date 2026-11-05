#!/usr/bin/env python3
"""generate.py HYP SCRATCH [DEST]: build the synthetic store of verifier-audit-arms.

HYP is the hyp binary and SCRATCH an empty directory. The store holds one plain
select on a fictional question, with eight Experiments decided by reading. Two
of its contradicts are known wrong, one is known right, and no text in the store
says which. Everything but event ids is deterministic. With DEST (the bundle's
fixture directory), it copies the store to DEST/tree and writes DEST/ids.env and
DEST/items.tsv. Only the variable names of this script use the words a Doctor
looks for in the store.
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

HYPOTHESES = [
    ("h-outbox", "Durable outbox table, retry worker",
     "Write each delivery to a durable outbox table in the same transaction as the business change. A retry worker reads the table, delivers, and marks each row when the receiver answers."),
    ("h-redis", "Redis delayed queue, AOF everysec",
     "Keep pending deliveries in a Redis sorted set scored by the next attempt time, on a Redis server with appendonly yes and appendfsync everysec, drained by a retry worker."),
    ("h-inline", "In-process retry, persisted attempt log",
     "Retry inside the web process with a timer per delivery, and write every attempt to a persisted attempt log on local disk, replayed at start-up to rebuild the pending set."),
    ("h-broker", "Managed broker, at-most-once delivery",
     "Publish each delivery to a managed broker queue, with consumers on at-most-once delivery: a message is removed from the queue when it is handed to the consumer."),
]


def cfg_outbox(batch):
    return f"outbox_table=webhook_outbox outbox_commit=same_transaction worker_batch={batch} worker_poll_ms=200"


CFG_REDIS = "redis_version=7.2.4 appendonly=yes appendfsync=everysec maxmemory_policy=noeviction"
CFG_REDIS_OFF = "redis_version=7.2.4 appendonly=no maxmemory_policy=noeviction"
CFG_INLINE_MEM = "retry_timers=per_delivery attempt_log_backend=memory attempt_log_replay=on_start"
CFG_BROKER = "queue=webhooks-eu delivery_mode=at-most-once ack_mode=auto prefetch=20"

# alias, parent, target, verdicts, title, numbers, config line, method sentences
# in the middle of method.md, observation, whether the summary reads as a failure
RUNS = [
    dict(name="sup1", parent="h-outbox", target="h-redis", verdicts={"h-outbox": "supports", "h-redis": "inconclusive"},
         title="Outbox survives a worker kill during retries",
         total=500, lost=0, restart="kill -9 of the retry worker at delivery 180", cfg=cfg_outbox(50), side="redis",
         middle="The service wrote rows to `webhook_outbox` in the same transaction as the order change, and one retry worker drained them in batches of 50. The Redis path ran alongside on only 20 deliveries, which is too few to judge it.",
         obs="All 500 deliveries reached the receiver after the retry worker was killed at delivery 180; 0 lost, 37 delivered twice and removed by the receiver's idempotency key. The Redis path saw only 20 deliveries, too few to judge."),
    dict(name="sup2", parent="h-outbox", target="h-inline", verdicts={"h-outbox": "supports", "h-inline": "inconclusive"},
         title="Outbox survives a full service restart",
         total=500, lost=0, restart="systemctl restart of the whole service at delivery 240", cfg=cfg_outbox(50), side="in-process",
         middle="The whole service, retry worker included, was restarted while 62 rows were due. The in-process path ran alongside on only 20 deliveries, which is too few to judge it.",
         obs="All 500 deliveries reached the receiver across a full service restart at delivery 240; 0 lost. The 62 rows due at the restart were delivered within 4 seconds of start-up. The in-process path saw only 20 deliveries, too few to judge."),
    dict(name="sup3", parent="h-outbox", target="h-broker", verdicts={"h-outbox": "supports", "h-broker": "inconclusive"},
         title="Outbox survives a database failover during retries",
         total=500, lost=0, restart="database failover to the standby at delivery 300", cfg=cfg_outbox(25), side="broker",
         middle="The primary database was failed over to its synchronous standby at delivery 300, with a batch of 25 in flight. The broker path ran alongside on only 20 deliveries, which is too few to judge it.",
         obs="All 500 deliveries reached the receiver across a database failover at delivery 300 and a worker restart; 0 lost. Rows of the batch in flight were delivered again and removed by the idempotency key. The broker path saw only 20 deliveries, too few to judge."),
    dict(name="rsup", parent="h-redis", target="h-outbox", verdicts={"h-redis": "supports", "h-outbox": "inconclusive"},
         title="Redis queue survives a graceful restart",
         total=500, lost=0, restart="graceful restart of redis-server at delivery 210", cfg=CFG_REDIS, side="outbox",
         middle="Redis ran from `redis.conf` with `appendonly yes` and `appendfsync everysec`, and was restarted with SHUTDOWN at delivery 210, after which the AOF was replayed. The outbox path ran alongside on only 20 deliveries, which is too few to judge it.",
         obs="All 500 deliveries reached the receiver across a graceful Redis restart at delivery 210; 0 lost, and the sorted set came back with 41 pending members from the AOF. The outbox path saw only 20 deliveries, too few to judge."),
    dict(name="binc", parent="h-broker", target="h-outbox", verdicts={"h-broker": "inconclusive", "h-outbox": "inconclusive"},
         title="Broker and outbox under a short receiver outage",
         total=200, lost=0, restart="receiver unavailable for 90 seconds at delivery 80", cfg=CFG_BROKER, side="outbox",
         middle="No process was restarted in this run: only the receiver went away for 90 seconds, so it cannot speak to a restart on either path. Both paths retried and delivered everything.",
         obs="With the receiver away for 90 seconds, both paths delivered all 200 deliveries; 0 lost. No process restarted in this run, so it says nothing about a restart on either path."),
    dict(name="seed_deciding", parent="h-redis", target="h-outbox", verdicts={"h-redis": "contradicts"},
         title="Redis queue under a crash of the server",
         total=500, lost=143, restart="kill -9 of redis-server at delivery 260", cfg=CFG_REDIS_OFF, side=None, failed=True,
         middle="Redis ran from the stock test image with `appendonly no` added to its command line, and the retry worker read the sorted set every 200 ms. The server was killed at delivery 260 with 143 members in the set.",
         obs="143 of 500 deliveries never reached the receiver after redis-server was killed at delivery 260. The sorted set came back empty after the restart, and the retry worker found nothing to retry. Pending deliveries do not survive a restart on this path."),
    dict(name="seed_lone", parent="h-inline", target="h-outbox", verdicts={"h-inline": "contradicts"},
         title="In-process retry under a restart of the web process",
         total=500, lost=118, restart="restart of the web process at delivery 275", cfg=CFG_INLINE_MEM, side=None, failed=True,
         middle="Each delivery had its own retry timer in the web process, and each attempt was appended to the attempt log, which the process kept in memory as an in-process array and read back at start-up. The web process was restarted at delivery 275.",
         obs="118 of 500 deliveries never reached the receiver after the web process restarted at delivery 275. The attempt log replay at start-up found 0 entries, so no retry timer came back. Pending deliveries do not survive a restart on this path."),
    dict(name="ctrl", parent="h-broker", target="h-outbox", verdicts={"h-broker": "contradicts"},
         title="Broker queue under a consumer restart",
         total=500, lost=20, restart="restart of the consumer at delivery 310", cfg=CFG_BROKER, side=None, failed=True,
         middle="The consumer ran with at-most-once delivery, with `ack_mode=auto` and a prefetch of 20, so the broker removed each message as it handed it over. The consumer was restarted at delivery 310 with 20 messages handed over and not yet delivered.",
         obs="20 of 500 deliveries never reached the receiver after the consumer restarted at delivery 310. The broker had already removed those 20 messages when it handed them over, and nothing retried them. Pending deliveries do not survive a restart on this path."),
]

ORDER = ["sup1", "sup2", "sup3", "rsup", "binc", "seed_deciding", "seed_lone", "ctrl"]


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
    place_cfg = shown // 2 - rng.randint(4, 10)
    place_restart = shown // 2 + rng.randint(4, 10)
    missing = set(rng.sample(range(shown // 2, shown), min(lost // 12, shown // 4)))
    for i in range(shown):
        if i == place_cfg:
            lines.append(f"{clock(rng, state)} config: {run['cfg']}")
        if i == place_restart:
            lines.append(f"{clock(rng, state)} step: {run['restart']}")
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
        f"# {run['title']}\n\n## Setup\n\n{fill[0]}\n\n{fill[1]}\n\n"
        f"## Run\n\n{run['middle']}\n\n{fill[4]}\n\n"
        f"## Reading the result\n\n{fill[2]}\n\n{fill[3]}\n"
    )


def main():
    if len(sys.argv) not in (3, 4):
        sys.exit("usage: generate.py HYP SCRATCH [DEST]")
    hyp, scratch = os.path.realpath(sys.argv[1]), os.path.realpath(sys.argv[2])
    dest = os.path.realpath(sys.argv[3]) if len(sys.argv) == 4 else None
    env = dict(os.environ, HYP_ACTOR=ACTOR)
    rng = random.Random(2611050916)

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
        args = ["record", "--exp", e, "--observation", run["obs"], "--why", "the run finished and its log was read",
                "--evidence", f".tree/evidence/{e}/run.log", "--evidence", f".tree/evidence/{e}/method.md"]
        for ref, word in run["verdicts"].items():
            args += ["--verdict", f"{ids[ref]}={word}"]
        results[name] = new(*args)
        exps[name] = e
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
        with open(os.path.join(dest, "ids.env"), "w") as f:
            f.write(f"QUESTION={q}\nSELECTED={ids['h-outbox']}\nSELECT_EVENT={select_event}\n"
                    f"TARGET_DECIDING={results['seed_deciding']}\nTARGET_LONE={results['seed_lone']}\nCONTROL={results['ctrl']}\n"
                    f"EXP_DECIDING={exps['seed_deciding']}\nEXP_LONE={exps['seed_lone']}\nEXP_CONTROL={exps['ctrl']}\n")
        with open(os.path.join(dest, "items.tsv"), "w") as f:
            for name in ORDER:
                for ref in runs[name]["verdicts"]:
                    f.write(f"{results[name]}\t{ids[ref]}\n")


if __name__ == "__main__":
    main()
