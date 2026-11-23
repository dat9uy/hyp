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

const word = (v) =>
  v.endBy !== "none" ? "ended" : v.selected ? "selected" : v.placed ? "open" : "written";

// The state block hyp prints after an act or a status, read from the state.
const block = (v, o) => {
  const lines = [`G  ${o.state || word(v)}`];
  if (o.endLine) lines.push(o.endLine);
  const bCounts = o.rows >= 0 ? o.rows - (v.runA ? 1 : 0) > 0 : false;
  if (v.runB === "merged" && v.endBy !== "none" && !bCounts) lines.push("refused at replay");
  if (v.endBy === "none") lines.push("owner may:", "  hyp close G");
  return lines;
};

export const print = (action, from, to) => {
  const o = to.out;
  switch (o.cmd) {
    case "close":
      if (o.refused) return `${o.by} close exit 1`;
      return [...block(to, o), "close exit 0"].join("\n");
    case "land":
      if (o.refused) return [...block(to, o), "land exit 1"].join("\n");
      return [...block(to, o), "land exit 0"].join("\n");
    case "run":
      return o.refused ? "run exit 1" : "run exit 0";
    case "add goal --from":
      return o.refused ? "add exit 1" : "recorded G-NEW\nadd exit 0";
    case "status, old log":
      return o.refused ? "refused at replay" : [
        "stop: selected, by select 01M7S4F7Y677ZNGG9Y0QQJMWSX",
        "landed: by owner, land 01M7YS8QZ3D2XNPY6EC2XAEKMS",
        "verdict: met, exit 0"].join("\n");
    case "status G":
    case "select":
    case "place":
      return block(to, o).join("\n");
  }
  return "";
};
