export const title = "G-5N7SRH: the select's report";

export const state = (v) => {
  const g = { open: "Open", selected: "Selected", ended: "Landed" }[v.gs];
  return v.clone === "fresh" ? `${g}, read in a fresh clone` : g;
};

export const step = (action, from, to) => {
  if (action.startsWith("RunnerSelects")) return "the runner selects the one candidate";
  if (action.startsWith("OwnerReadsStatus")) return "the owner reads hyp status G";
  if (action.startsWith("OwnerLands")) return "the owner lands the Goal";
  if (action.startsWith("OwnerClones")) return "the owner reads in a fresh clone";
  if (action.startsWith("ReaderGreps")) return null;
  return action;
};

const has = (o, p) => (o.parts || []).includes(p);

function report(o, ended) {
  const l = ["# G-ONE: Greet.", "", "## line-1 (gap): the subject says bye."];
  if (has(o, "examples")) {
    l.push("", "### line-1/a: Fixture line-1/a.", "");
    if (has(o, "try"))
      l.push("verdict: not in this clone's trace; run it again in this clone with:",
        "  hyp try G-ONE line-1/a --commit COMMIT-ONE");
    else if (has(o, "verdict"))
      l.push(ended ? "verdict: met, by select SELECT-ONE at TIME" : "verdict: met, exit 0");
    if (has(o, "output")) l.push("", "output, 9 bytes:", "", "```", "hello bye", "```");
    if (has(o, "stderr")) l.push("", "stderr, 16 bytes:", "", "```", "fixture-warning", "```");
  }
  if (has(o, "charts")) l.push("", "## Chart: goals/G-ONE/specs/chart.mmd", "", "```mermaid", "stateDiagram-v2", "```");
  l.push("", "selected commit: COMMIT-ONE", "selected row: ROW-ONE");
  return l.join("\n");
}

export const print = (action, from, to) => {
  const o = to.out;
  if (o.cmd === "select") return report(o, false);
  if (o.cmd === "status") {
    const ended = to.gs === "ended";
    const head = ended ? "G-ONE  ended\nstop: selected, by select SELECT-ONE at TIME\n\nPR body:\n\n"
      : "G-ONE  selected  runs: 1 of 3 used\nstop: selected, by select SELECT-ONE at TIME\n\nPR body:\n\n";
    return head + report(o, ended);
  }
  if (o.cmd === "grep") {
    if (o.doc === "guide") return o.says ? "1" : "0";
    return o.says
      ? "433:**PR body**:\n434-The select's report: each line, each example's words, output, stderr, exit code and verdict, and the selected commit and row."
      : "433:**PR body**:\n434-The select's report: the Goal's words, each line with its kind, each placed example's key and words, each chart of the model, and the selected commit and row.";
  }
  return "";
};
