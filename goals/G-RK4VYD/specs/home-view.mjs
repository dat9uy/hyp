// The view of home.tla for the statechart: the trace of clone c2 under
// $HOME/.hyp, from its first read to the owner's prune. Clone c1, which has
// two worktrees, stays live in every chart state a prune can reach; its own
// steps show only when they change what the chart shows of c2.

export const title = "hyp G-RK4VYD: one clone's trace in the home folder, and hyp prune";

export const state = (v) => {
  const traced = (v.trace.hyp.c2 ?? []).length > 0;
  if (v.live.includes("c2")) return traced ? "Clone live, its reads traced" : "Clone live, no read yet";
  if (!traced) return "Clone gone, no trace left";
  return v.last === "prune" ? "Clone gone, trace named by hyp prune" : "Clone gone, trace kept";
};

export const group = (v) => (v.live.includes("c2") ? "live" : "gone");

export const colors = { live: "teal", gone: "orange" };

export const step = (action, from, to) => {
  switch (action) {
    case "AgentReads":
      return (to.got.c2 ?? []).length > (from.got.c2 ?? []).length
        ? "an agent reads in the clone; hyp records it under $HOME/.hyp"
        : "an agent reads in the other clone";
    case "EnvChanges":
      return "$XDG_STATE_HOME is set or unset";
    case "CloneDeleted":
      return "the clone is deleted or moved";
    case "OwnerShows":
      return "the owner runs hyp show";
    case "OwnerPrunes":
      return to.last === "remove" ? "the owner runs hyp prune --remove" : "the owner runs hyp prune";
  }
  return action;
};
