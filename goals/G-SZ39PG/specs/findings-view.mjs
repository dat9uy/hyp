// The view of findings.tla: what hyp prints to an agent for each finding act.

export const title = "hyp G-SZ39PG, a finding hangs on its Goal and is folded into an issue or refuted";

const label = (s) => ({ none: null, notfolded: "not folded", folded: "folded", refuted: "refuted" })[s];

export const state = (v) => {
  const ids = Object.keys(v.f);
  const parts = ids.map((i) => label(v.f[i])).filter(Boolean);
  const fs = parts.length === 0 ? "no finding" : [...new Set(parts)].sort().join(" + ");
  const g = v.goal === "ended" ? "Goal landed" : v.goal === "selected" ? "Goal selected" : "Goal open";
  return [g, fs];
};

export const group = (v) => v.goal;
export const colors = { open: "teal", selected: "purple", ended: "blue" };

export const step = (action, from, to) => {
  switch (action) {
    case "AgentAdds": return to.out.ok ? "an agent adds a finding" : "an agent adds a finding [refused]";
    case "AgentAddsNoGoal": return null;
    case "AgentFolds": return to.out.ok ? "an agent folds into an issue" : "an agent folds [refused]";
    case "AgentFoldsNoIssue": return to.out.ok ? "an agent folds with no issue" : null;
    case "AgentRefutes": return to.out.ok ? "an agent refutes a finding" : "an agent refutes [refused]";
    case "AgentReadsStatus": return null;
    case "AgentReadsSpec": return null;
    case "RunnerSelects": return "the runner selects";
    case "OwnerLands": return to.out.ok ? "the owner lands" : "the owner lands [refused]";
  }
  return action;
};

const shown = (v, i) => {
  const s = v.f[i];
  if (s === "none") return null;
  return s;
};

export const print = (action, from, to) => {
  const o = to.out;
  switch (o.cmd) {
    case "add": return o.ok ? "recorded F-" : "refused: the Goal has ended";
    case "add-no-goal": return o.ok ? "recorded F-" : "refused: no Goal G-NOGOAL";
    case "fold": return o.ok ? "folded into #" : (o.why === "refuted" ? "refused: the finding is refuted" : "refused: the finding is folded");
    case "fold-no-issue": return o.ok ? "folded" : "refused: a fold names an issue";
    case "refute": return o.ok ? "refuted" : "refused: the finding is refuted";
    case "land": return o.ok ? "landed" : "refused: a finding is not folded";
    case "select": return "selected";
    case "spec": return o.states ? "finding.add\nfinding.fold\nfinding.refute\nnot folded: fold, refute\nfolded: fold, refute\nrefuted: none" : "finding.add\nfinding.fold\nfinding.refute";
    case "status": {
      const lines = [];
      for (const i of Object.keys(to.f)) {
        const s = o.shown[i];
        if (s === "absent") continue;
        if (s === "no state") { lines.push("F-" + i); continue; }
        if (s === "notfolded") lines.push("F-" + i + "  not folded");
        else if (s === "folded") lines.push("F-" + i + "  folded into #" + to.issue[i]);
        else lines.push("F-" + i + "  refuted");
      }
      return lines.join("\n");
    }
  }
  return "";
};
