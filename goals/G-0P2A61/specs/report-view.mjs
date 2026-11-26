// The view of report.tla: what hyp prints for each act and read.

export const title = "hyp G-0P2A61, the select's report";

export const state = (v) =>
  ({ open: "Open", selected: "Selected", landed: "Landed" })[v.gs] +
  (v.clone === "fresh" ? ", fresh clone" : "");
export const group = (v) => v.gs;
export const colors = { open: "blue", selected: "purple", landed: "orange" };

export const step = (action) =>
  ({
    RunnerSelects: "the runner selects",
    OwnerReadsStatus: "the owner reads hyp status G",
    OwnerLands: "the owner lands",
    OwnerClones: "the owner clones the store",
  })[action] ?? null;

const has = (o, p) => (o.parts.includes ? o.parts.includes(p) : false);

// The report as hyp prints it, in the words the examples' floors use.
const report = (o, to) => {
  const l = [];
  if (o.cmd === "status") {
    l.push("G-ONE  " + (o.gs === "landed" ? "ended" : "selected"));
    l.push("stop: selected, by select SELECT-ONE at TIME");
    l.push("PR body:");
  }
  l.push("# G-ONE: Greet.");
  l.push("## line-1 (gap): the subject says bye.");
  l.push("### line-1/a: Fixture line-1/a.");
  if (has(o, "verdict")) {
    if (o.gs === "landed") l.push("verdict: met, by select SELECT-ONE at TIME");
    else if (has(o, "try")) l.push("verdict: not in this clone's trace; run it again in this clone with:\n  hyp try G-ONE line-1/a --commit COMMIT-ONE");
    else l.push("verdict: met, exit 0");
  }
  if (has(o, "output")) l.push("output, 10 bytes:\nhello bye");
  if (has(o, "stderr")) l.push("stderr, 16 bytes:\nfixture-warning");
  if (has(o, "chart")) l.push("## Chart: goals/G-ONE/specs/chart.mmd\n```mermaid\nstateDiagram-v2\n```");
  l.push("selected commit: COMMIT-ONE");
  l.push("selected row: ROW-ONE");
  return l.join("\n");
};

export const print = (action, from, to) => {
  const o = to.out;
  switch (o.cmd) {
    case "select": return "selected G-ONE\n" + report(o, to);
    case "status": return report(o, to);
    case "land": return "landed G-ONE";
    case "clone": return "cloned";
  }
  return null;
};
