// The view of foldlines.tla: what hyp prints to an agent for each act and read.

export const title = "hyp G-G5F1JD, the ready lines of a finding";

const ids = (v) => Object.keys(v.f);
const inWork = (v) => Object.keys(v.gs).some((g) => v.gs[g] === "work");

export const state = (v) => {
  const i = ids(v)[0];
  const s = v.f[i];
  const work = inWork(v) ? "a Goal in work" : "no Goal in work";
  if (s === "none") return "no finding, " + work;
  if (s === "notfolded") return "not folded, " + work;
  if (s === "refuted") return "refuted";
  const ended = v.gs[v.into[i]] === "ended";
  return (ended ? "folded into an ended Goal, " : "folded into a Goal in work, ") + work;
};

export const group = (v) => {
  const s = v.f[ids(v)[0]];
  return s === "none" ? "none" : s === "notfolded" ? "free" : s === "folded" ? "folded" : "refuted";
};
export const colors = { none: "blue", free: "teal", folded: "purple", refuted: "orange" };

export const step = (action, from, to) => {
  const o = to.out;
  switch (action) {
    case "AgentAdds": return "an agent records a finding";
    case "AgentFolds": return o.ok ? "an agent folds it into a Goal" : null;
    case "AgentRefutes": return "an agent refutes it";
    case "GoalEnds": return "a Goal ends";
  }
  return null;
};

const stateText = (v, i, s) =>
  s === "notfolded" ? "not folded"
  : s === "refuted" ? "refuted"
  : "folded into G-" + v.into[i] + (v.gs[v.into[i]] === "ended" ? ", ended" : "");

const mayBlock = (o, i) => {
  const lines = [];
  const fold = [...o.fold[i]].sort();
  if (fold.length === 0 && !o.refute[i]) return lines;
  lines.push("  may:");
  for (const g of fold) lines.push("    hyp finding fold F-" + i + " --goal " + (g === "G" ? "G" : "G-" + g));
  if (o.refute[i]) lines.push("    hyp finding refute F-" + i);
  return lines;
};

export const print = (action, from, to) => {
  const o = to.out;
  switch (o.cmd) {
    case "add": return "recorded F-";
    case "fold":
      if (o.ok) return "recorded F-ONE  folded into G-" + o.why;
      return "hyp: refused: In work: G-" + o.why + " is ended; a fold names a Goal in work: written, open or selected\nexit 1";
    case "refute": return "recorded F-ONE  refuted";
    case "list": {
      const lines = [];
      for (const i of ids(to)) {
        const s = o.shown[i];
        if (s === "absent") continue;
        lines.push("F-" + i + "  " + stateText(to, i, s) + "  by agent at");
        lines.push("  fact: The subject says hello.");
        lines.push("  evidence: greet.txt line 1");
        lines.push(...mayBlock(o, i));
      }
      lines.push("exit 0");
      return lines.join("\n");
    }
    case "status":
      return o.count + " not folded: hyp finding list --not-folded\nexit 0";
    case "status-g": {
      const lines = [];
      const shown = ids(to).filter((i) => o.shown[i] !== "absent");
      if (shown.length > 0) lines.push("findings:");
      for (const i of shown) {
        lines.push("  F-" + i + "  folded into G-" + o.why + "  by agent at");
        lines.push("    evidence: greet.txt line 1");
        lines.push("    fact: The subject says hello.");
        lines.push(...mayBlock(o, i).map((l) => "  " + l));
      }
      lines.push("exit 0");
      return lines.join("\n");
    }
  }
  return "";
};
