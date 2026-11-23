// The view of channel.tla for the statechart: the runner's command in one
// region, and how the owner's show compares with what the runner got in the
// other.

const seq = (s) =>
  Array.isArray(s) ? s : Object.keys(s ?? {}).sort((a, b) => a - b).map((k) => s[k]);

export const title = "hyp's channel: what the runner gets, what the owner's show holds";

const command = (v) =>
  ({
    idle: "No command running",
    "held back": "Output written, held back",
    traced: "Trace recorded, not yet printed",
  })[v.phase];

const owner = (v) => {
  const got = seq(v.got);
  const results = got.filter((b) => b.kind !== "notice").length;
  const shown = seq(v.trace).length;
  if (got.some((b) => b.kind === "notice")) return "Runner got a notice show lacks";
  if (v.crashes > 0) return "Show holds a read the runner never got";
  if (shown > results) return "Show is one read ahead";
  return "Show holds what the runner got";
};

export const state = (v) => ({ Command: command(v), "Owner's show": owner(v) });

export const group = (v) => {
  const o = owner(v);
  if (o === "Runner got a notice show lacks" || o === "Show holds a read the runner never got") return "differs";
  return "same";
};

export const colors = { same: "teal", differs: "orange" };

export const step = (action, from, to) => {
  switch (action) {
    case "RunnerReads":
      return "the runner runs hyp status";
    case "RunnerIsRefused":
      return "a guard refuses the runner";
    case "TraceWritten":
      return "hyp records the trace";
    case "TraceFails":
      return "the trace cannot be written; hyp prints only a notice";
    case "Printed":
      return "hyp prints the output";
    case "CrashBeforeTrace":
      return "hyp crashes before the trace";
    case "CrashAfterTrace":
      return "hyp crashes after the trace";
    case "OwnerChangesLog":
      return "the owner changes the log";
    case "Done":
      return null;
  }
  return action;
};
