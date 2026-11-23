// The view of ended_goal.tla: what each step prints, in the words of the
// examples' Probes.

export const title = "hyp G-KSCGXV, an ended Goal reads the same in every clone";

export const state = (v) => {
  const g = v.landed ? "Ended" : v.selected ? "Selected" : v.ran ? "Run in clone A" : "Open";
  return v.gitResult ? `${g}, git holds an older hyp's Result` : g;
};

export const group = (v) => (v.landed ? "ended" : v.selected ? "selected" : "open");

export const colors = { open: "teal", selected: "purple", ended: "blue" };

export const step = (action) => {
  switch (action) {
    case "RunnerRuns": return "the runner runs a line in clone A";
    case "RunnerSelects": return "the runner selects";
    case "OwnerLands": return "the owner lands";
    case "OlderHypWrites": return "an older hyp's Result line is in git";
    default: return null;
  }
};

const yn = (b) => (b ? "yes" : "no");

export const print = (action, from, to) => {
  const o = to.out;
  switch (o.cmd) {
    case "compare":
      return [
        `same status in clone A and clone B: ${yn(o.same)}`,
        `each example names the select: ${yn(o.names)}`,
        "hyp try line: no",
        "output block: no",
        `runs count in clone A: ${yn(o.runsA)}`,
        `runs count in clone B: ${yn(o.runsB)}`,
      ].join("\n");
    case "gitresult":
      return `same output with and without the event in git: ${yn(o.same)}`;
    case "table":
      return `bytes from .tree/outputs/ printed: ${yn(o.oldBytes)}`;
    case "spec":
      return o.earlier
        ? "mentions an earlier hyp: yes\nit still reads the outputs an earlier hyp stored there"
        : "mentions an earlier hyp: no";
    case "replay":
      return ["ended Goals: 20", `exit 0 for every Goal: ${yn(o.exitAll)}`, "refused at replay: none"].join("\n");
  }
  return "";
};
