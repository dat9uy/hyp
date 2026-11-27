// The view of store.tla: the two branches' acts, the merge, and what hyp
// and git print.

export const title = "hyp's store, two branches and a plain merge";

const acts = (v, b) => {
  const n = v.added[b] + v.changed[b];
  if (n === 0) return "no act";
  return `${n} act${n > 1 ? "s" : ""}${v.ended[b] ? ", G-ONE ended" : ""}`;
};

export const state = (v) => {
  const store = v.form === "old" ? "older store" : "store";
  if (v.phase === "merged") {
    if (v.conflict) return `${store}, merge stopped on a conflict`;
    return `${store}, merged${v.applied > 1 ? ", both ends applied" : ""}${v.reads ? "" : ", base Goal not read"}`;
  }
  return `${store}, main: ${acts(v, "main")}; side: ${acts(v, "side")}`;
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
