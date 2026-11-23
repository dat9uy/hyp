// The view of goal_end.tla: what the owner and the runners see in clone A.

export const title = "hyp G-GF11ED: the owner ends a Goal with land or close";

const goal = (v) =>
  v.endBy === "land" ? "Ended by land"
  : v.endBy === "close" ? "Ended by close"
  : v.selected ? "Selected"
  : v.placed ? "Open"
  : "Written";

export const state = (v) => {
  const b = v.runB === "none" ? "B ran nothing"
    : v.runB === "unmerged" ? "B ran, not merged"
    : "B's run merged";
  return [goal(v), b];
};

export const group = (v) =>
  v.endBy !== "none" ? "ended" : v.selected ? "selected" : v.placed ? "open" : "written";

export const colors = { written: "orange", open: "teal", selected: "purple", ended: "blue" };

export const step = (action, from, to) => {
  switch (action) {
    case "GrillPlaces": return "the grill places the set";
    case "RunnerRunsA": return to.out.refused ? null : "a runner in A runs a line";
    case "RunnerSelects": return "the runner selects";
    case "RunnerRunsB": return "a runner in B runs a line";
    case "LogsMerge": return "the logs of A and B merge";
    case "Closes": return to.out.refused ? null : "the owner closes";
    case "OwnerLands": return to.out.refused ? null : "the owner lands";
    case "AddFrom": return null;
    case "OldLogReplays": return null;
    case "ShowStatus": return "you run hyp status G";
  }
  return action;
};

export const print = (action, from, to) => {
  const o = to.out;
  if (o.refused) return `${o.cmd}: refused: ${o.refused}`;
  if (o.cmd === "run in B" || o.cmd === "merge" || o.cmd === "none") return "";
  if (o.cmd === "add goal --from") return "recorded G2";
  const lines = [`G  ${o.state}`];
  if (o.endLine) lines.push(o.endLine);
  if (o.rows >= 0) lines.push(`rows: ${o.rows}`);
  return lines.join("\n");
};
