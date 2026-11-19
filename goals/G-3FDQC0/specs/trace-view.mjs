// The view of trace.tla for the statechart: one command at a time, from the
// moment an agent runs it to the moment hyp prints, and the owner's merge.

export const title = "hyp G-3FDQC0: the trace is not the log";

export const state = (v) =>
  ({
    idle: "No command running",
    probing: "Act: Probes running",
    checked: "Act: outputs hashed, guards checked again",
    "outputs written": "Act: outputs written, event not yet",
    ended: "Command ended, output held back",
    traced: "Trace recorded outside the repository",
  })[v.phase];

export const group = (v) =>
  ["probing", "checked", "outputs written"].includes(v.phase) ? "act" : v.phase === "idle" ? "rest" : "trace";

export const colors = { rest: "teal", act: "orange", trace: "purple" };

export const step = (action, from, to) => {
  switch (action) {
    case "Start":
      return to.phase === "probing"
        ? "an agent runs an act that runs Probes"
        : "an agent reads, is refused, or gives a bad argument";
    case "ProbesEnd":
      return "the Probes end; hyp hashes their outputs";
    case "RecheckRefuses":
      return "a guard checked again refuses; nothing is written";
    case "OutputsWritten":
      return "every guard holds; hyp writes the outputs";
    case "EventWritten":
      return "hyp appends the act's event";
    case "CrashBeforeEvent":
      return "hyp crashes before the event";
    case "TraceWritten":
      return "hyp records what it will print, in the clone's trace folder";
    case "Printed":
      return "hyp prints to the agent";
    case "OwnerMerges":
      return "the owner merges the runner's worktree";
    case "Done":
      return null;
  }
  return action;
};
