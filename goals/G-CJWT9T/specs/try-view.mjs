// The view of try.tla for the statechart: where the example is, what the
// checkout holds, what hyp try has shown, and what the log records. A try
// changes only what it has shown: the log, the checkout and the example stay
// as they were.

export const title = "hyp G-CJWT9T: hyp try shows a verdict before place or run records it";

const isH0 = (s) => !s.edit && !s.new;
const same = (a, b) => a.edit === b.edit && a.new === b.new;

export const state = (v) => {
  const log = Object.values(v.log ?? {});
  const placed = log.some((e) => e.type === "place");
  const ran = log.some((e) => e.type === "run");
  const shows = v.shows ?? [];
  const onH0 = shows.some((s) => s.on === "H0");
  const onTree = shows.some((s) => s.on === "worktree");
  return {
    Example: placed ? "Placed" : v.comEx !== "none" ? "Committed" : v.diskEx !== "none" ? "On disk only" : "Not written",
    Checkout: isH0(v.diskSubj) ? "As H0" : same(v.diskSubj, v.headSubj) ? "Change committed" : "Change not committed",
    Try: onH0 && onTree ? "Shown on H0 and on the checkout" : onH0 ? "Shown on H0" : onTree ? "Shown on the checkout" : "Nothing shown",
    Log: ran ? "A run recorded" : placed ? "H0's cell recorded" : "Nothing recorded",
  };
};

export const group = (v) => ((v.shows ?? []).length > 0 ? "tried" : "untried");

export const colors = { tried: "teal", untried: "tan" };

export const step = (action, from, to) => {
  switch (action) {
    case "GrillWritesExample":
      return "the grill writes the example folder";
    case "GrillCommitsExample":
      return "the grill commits it";
    case "AgentEditsFile":
      return "an agent edits a tracked file";
    case "AgentAddsFile":
      return "an agent adds a new file";
    case "AgentCommitsChange":
      return "the agent commits the change";
    case "AgentTriesOnH0":
      return "hyp try G LINE/EXAMPLE";
    case "AgentTriesOnWorktree":
      return "hyp try G LINE/EXAMPLE --worktree";
    case "GrillPlaces":
      return "hyp place records H0's cell";
    case "RunnerRuns":
      return "hyp run records a Result";
    case "Done":
      return null;
  }
  return action;
};
