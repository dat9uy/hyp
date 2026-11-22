// The view of git_keeps.tla: what each command prints in clone A or B.

export const title = "hyp G-G4NNT7: git keeps the owner's acts; the trace keeps the evidence";

const goal = (v) => (v.landed ? "Landed" : v.gitSelect !== "none" ? "Selected" : v.gitPlace ? "Open" : "Written");

export const state = (v) => {
  const b = [];
  if (v.trH0.B === "rerun" || v.trCand.B === "rerun") b.push("B ran a cell again");
  else if (v.trCand.B === "cell") b.push("B ran the candidate");
  else b.push("B ran nothing");
  return [goal(v), b[0] + (v.reach ? "" : ", selected commit on no ref")];
};

export const group = (v) => goal(v).toLowerCase();

export const colors = { written: "orange", open: "teal", selected: "purple", landed: "blue" };

export const step = (action, from, to) => {
  switch (action) {
    case "GrillPlaces": return "the grill places the set";
    case "RunnerRuns": return to.refusedB ? "a runner in B runs [refused]" : "a runner runs a line";
    case "RunnerSelects": return "the runner selects";
    case "OwnerLands": return "the owner lands";
    case "BranchGoes": return "the selected commit's branch goes";
    case "Status": return null;
    case "ReaderTries": return "a reader in B runs the try line";
  }
  return action;
};

const cellLines = (c) => {
  switch (c) {
    case "met": return ["line-1/a: met"];
    case "met, run again": return ["line-1/a: met, run again in this clone"];
    case "not run": return ["line-1/a: not run"];
    case "missing": return ["line-1/a: not in this clone's trace"];
    case "missing, try line": return ["line-1/a: not in this clone's trace", "hyp try G line-1/a"];
  }
  return [];
};

export const print = (action, from, to) => {
  const o = to.out;
  if (o.cmd === "run refused") return "refused: no cell of H0";
  if (o.state === "") return "";
  const lines = [`G  ${o.state}`];
  if (o.runs >= 0) lines.push(`runs: ${o.runs} of 3 used`);
  if (o.warn) lines.push("warning: the selected commit is on no branch or tag");
  return [...lines, ...cellLines(o.cell)].join("\n");
};
