// The view of store.tla: the two branches' acts, the merge, and what hyp
// and git print.

export const title = "hyp's store, two branches and a plain merge";

const acts = (v, b) => {
  if (v.ended[b]) return "G-ONE closed";
  return v.added[b] + v.changed[b] === 0 ? "no act" : "acts";
};

export const state = (v) => {
  if (v.phase === "merged") {
    if (v.conflict) return "merge stopped on a conflict";
    return `merged${v.applied > 1 ? ", both closes applied" : ""}${v.reads ? "" : ", base Goal not read"}`;
  }
  return `main: ${acts(v, "main")}; side: ${acts(v, "side")}`;
};

export const colors = { apart: "blue", merged: "teal" };

export const group = (v) => (v.phase === "merged" ? "merged" : "apart");

export const step = (action) =>
  ({
    StoreAddGoal: "the owner adds a Goal on a branch",
    StoreClose: "the owner closes G-ONE on a branch",
    GitMerge: "git merge, with no merge driver",
  })[action] ?? action;

export const print = (action, from, to) => to.out;
