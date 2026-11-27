// The view of change.tla: what hyp and the statechart tool print on each step.

export const title = "G-XVME1C: the kept change at the Selected stop";

export const state = (v) => (v.st === "none" ? `${v.case}` : `${v.case}: ${v.st}`);

export const print = (action, from, to) => [...to.out.items].sort().join("\n");
