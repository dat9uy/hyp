// The view of goal.tla: a Goal's lifecycle when git holds only the owner's
// acts, and what hyp prints on each act.

export const title = "A Goal's lifecycle, owner acts in git, Results in each clone";

const placed = (v) => [...v.atAdd, ...v.later];

export const state = (v) => {
  const store = v.form === "old" ? "old store: " : "";
  if (!v.added) return "no Goal";
  if (v.landed !== "none") return `${store}landed on ${v.landed}`;
  if (v.closed) return `${store}closed`;
  if (placed(v).length === 0) return "written, no example";
  const sel = [...new Set(Object.keys(v.results).map((k) => selected(v, k)))];
  const shown = sel.filter((s) => s !== "none");
  const unsel = v.rejected.length ? `, ${v.rejected.join(" ")} unselected` : "";
  if (shown.length) return `${store}selected on ${shown.join(" / ")}${unsel}`;
  return `${store}open${unsel}`;
};

// Mirrors SelectedIn of goal.tla only for the labels; the rules use TLC's.
const selected = (v, k) => {
  if (v.form === "old" && !v.anyUnselect) return "c1";
  const ex = placed(v);
  const cands = Object.keys(v.results[k]).filter(
    (c) =>
      !v.unselected.includes(c) &&
      ex.length > 0 &&
      ex.every((e) => v.results[k][c][e] === "met"),
  );
  return cands.length === 1 ? cands[0] : "none";
};

export const group = (v) => {
  if (!v.added) return "start";
  if (v.landed !== "none" || v.closed) return "ended";
  return "open";
};

export const colors = { start: "blue", open: "teal", ended: "tan" };

export const step = (action, from, to) =>
  ({
    AddGoal: "the owner adds the Goal with its examples",
    Place: "the owner adds an example",
    Run: "the runner runs an example on a commit",
    Status: "you read hyp status",
    Land: "the owner lands",
    Unselect: "the owner unselects the commit shown selected",
    Close: "the owner closes the Goal",
  })[action] ?? action;

export const print = (action, from, to) => to.out;
