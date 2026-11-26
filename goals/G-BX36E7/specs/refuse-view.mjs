// The view of refuse.tla: what the statechart tool prints on each Goal model,
// and what a runner reads in the run rules.

export const title = "hyp G-BX36E7, the statechart tool says when hyp select refuses the check config";

const label = {
  none: "Nothing run",
  unnamed: "Tool ran on a rule no variant breaks",
  named: "Tool ran on a model every rule of which a variant breaks",
  breaks: "Tool ran on a model that breaks a rule",
  rules: "Runner read the run rules",
};

export const state = (v) => label[v.out.cmd === "rules" ? "rules" : v.out.case];

const stepLabel = { RunToolOnUnnamed: "run the tool on a rule no variant breaks", RunToolOnNamed: "run the tool on a model whose every rule a variant breaks", RunToolOnBreaks: "run the tool on a model that breaks a rule", ReadRunRules: "read the run rules" };

export const step = (action) => stepLabel[action] ?? action;

export const print = (action, from, to) => {
  const o = to.out;
  if (o.cmd === "rules") {
    const lines = ["Run only the lines hyp prints."];
    if (o.silent) lines.push("`hyp status` does not say so.");
    if (o.points) lines.push("make the fix that hyp status names");
    return lines.join("\n");
  }
  const lines = [];
  if (o.case === "unnamed") lines.push("| TypeOK | model only | not shown to fail | no |");
  if (o.case === "breaks") lines.push("Positive breaks");
  if (o.wrote) lines.push("- Variant configs written: goals/G-ONE/specs/variants/Breaks.cfg.");
  if (o.says) lines.push("- hyp select refuses this check config");
  if (o.wrote) lines.push("Wrote out");
  lines.push(`tool exit: ${o.exit}`);
  return lines.join("\n");
};
