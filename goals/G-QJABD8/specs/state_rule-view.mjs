// The view of state_rule.tla: what each command prints in clone A or B.

export const title = "hyp G-QJABD8: git keeps the events that change a Goal's state; the trace keeps the evidence";

const goal = (v) => (v.landed ? "Landed" : v.gitSelect ? "Selected" : v.gitPlace ? "Open" : "Written");

export const state = (v) => {
  let b;
  if (v.trH0.B === "rerun" || v.trCand.B === "rerun") b = "B ran a cell again";
  else if (v.trCand.B === "cell") b = "B ran the candidate";
  else b = "B ran nothing";
  return [goal(v), b + (v.reach ? "" : ", selected commit on no ref")];
};

export const group = (v) => goal(v).toLowerCase();

export const colors = { written: "orange", open: "teal", selected: "purple", landed: "blue" };

export const step = (action, from, to) => {
  switch (action) {
    case "ReadsSpec": return null;
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
  if (o.cmd === "spec") {
    return [
      `result.record  ${o.stateMark ? "changes a Goal's state: no  " : ""}stored in ${o.resultHome}`,
      `select  ${o.stateMark ? "changes a Goal's state: yes  " : ""}stored in git`,
      `Candidate  ${o.replayMark ? "checked at replay: no" : ""}`.trimEnd(),
    ].join("\n");
  }
  if (o.cmd === "run refused") return "refused: no cell of H0";
  if (o.state === "") return "";
  const lines = [`G  ${o.state}`];
  if (o.runs >= 0) lines.push(`runs: ${o.runs} of 3 used`);
  if (o.warn) lines.push("warning: the selected commit is on no branch or tag");
  if (o.refused) lines.push("refused at replay: select");
  return [...lines, ...cellLines(o.cell)].join("\n");
};
