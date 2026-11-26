// The view of saywhy.tla: the candidates of an open Goal, and what the
// last hyp status or hyp show printed.

export const title = "hyp G-W8MR53: say why a select is held back, and what show takes";

const holdOf = (v, c) => v.cand[c];
const held = (v) => ["c1", "c2"].filter((c) => !["absent", "free"].includes(holdOf(v, c)));
const present = (v) => ["c1", "c2"].filter((c) => holdOf(v, c) !== "absent");

export const state = (v) => {
  const n = present(v).length;
  const stop = n === 2 ? "Two candidates" : n === 1 ? "one candidate" : "no candidate";
  const h = held(v).length;
  return [stop, h === 0 ? "no select held back" : `${h} held back`];
};

export const group = (v) => (held(v).length ? "held" : "clear");
export const colors = { held: "orange", clear: "teal" };

const fixes = {
  "no-spec": "commit the spec, then run one line on that commit",
  "rule-no-variant": "commit a variant config under goals/G-ONE/specs/variants/ that lists each rule it breaks, then run one line on that commit",
  "no-chart": "commit the chart (.mmd), then run one line on that commit",
};

export const step = (action, from, to) => {
  switch (action) {
    case "RunnerRunsRow": return "the runner runs a row to a candidate";
    case "RunnerFixesHold": return "the runner commits the fix";
    case "ReadsStatus": return `${to.out.role} reads hyp status${to.out.withG ? " G" : ""}`;
    case "RunsShow": return `hyp show with ${to.out.arg === "actor" ? "an actor" : to.out.arg === "goal-id" ? "a Goal id" : "another word"}`;
  }
  return action;
};

export const print = (action, from, to) => {
  const o = to.out;
  if (action === "ReadsStatus") {
    const lines = ["G-ONE  open"];
    for (const c of [...o.selects].sort()) lines.push(`  hyp select G-ONE ${c}`);
    for (const n of [...o.notes].sort((a, b) => (a.c < b.c ? -1 : 1))) {
      if (n.hold === "free") { lines.push(`select of ${n.c}: ready`); continue; }
      lines.push(`select of ${n.c} held back: ${n.guard} ${n.hold}` + (n.fix ? `; ${fixes[n.hold]}` : ""));
    }
    return lines.join("\n");
  }
  if (action === "RunsShow") {
    if (o.arg === "actor") return "no read of runner-1 is recorded\nshow exit: 0";
    const arg = o.arg === "goal-id" ? "G-ONE" : "Not_An_Actor";
    let msg = o.actorword
      ? `hyp: show ${arg}: show takes an actor name`
      : `hyp: show ${arg}: an actor matches ^[a-z0-9][a-z0-9-]{0,39}$`;
    if (o.statushint) msg += "; hyp status G-ONE prints that Goal";
    return `${msg}\nshow exit: ${o.exit}`;
  }
  return "";
};
