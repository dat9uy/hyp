// The view of decline.tla: what hyp prints for each act and read.

export const title = "hyp G-QC8DF5, the decline of a finding";

const ids = (v) => Object.keys(v.f);

export const state = (v) => {
  const i = ids(v)[0];
  const s = v.f[i];
  if (s === "none") return "no finding";
  if (s === "notfolded") return "not folded";
  if (s === "refuted") return "refuted";
  if (s === "declined") return "declined";
  return v.gs[v.into[i]] === "ended" ? "folded into an ended Goal" : "folded into a Goal in work";
};

export const group = (v) => {
  const s = v.f[ids(v)[0]];
  return s === "declined" || s === "refuted" ? "final" : s === "none" ? "none" : "open";
};
export const colors = { none: "blue", open: "teal", final: "orange" };

export const step = (action, from, to) => {
  const o = to.out;
  switch (action) {
    case "AgentAdds": return "an agent records a finding";
    case "AgentFolds": return o.ok ? "an agent folds it into a Goal in work" : null;
    case "AgentRefutes": return o.ok ? "an agent refutes it" : null;
    case "Declines": return o.ok ? "the owner declines it, with a why" : null;
    case "GoalEnds": return "the Goal ends";
  }
  return null;
};

const stateText = (s) =>
  s === "notfolded" ? "not folded"
  : s === "held" ? "folded into G-ONE"
  : s === "folded" ? "folded into G-ONE, ended"
  : s;

export const print = (action, from, to) => {
  const o = to.out;
  switch (o.cmd) {
    case "add": return "recorded F-ONE";
    case "end": return "ended G-ONE";
    case "fold":
      return o.ok ? "recorded F-ONE  folded into G-ONE" : "hyp: refused: F-ONE is " + stateText(o.prior) + "\nexit 1";
    case "refute":
      return o.ok ? "recorded F-ONE  refuted" : "hyp: refused: F-ONE is " + stateText(o.prior) + "\nexit 1";
    case "decline": {
      if (o.ok) return "recorded F-ONE  " + to.f[o.id];
      if (o.role !== "owner") return "hyp: refused: only the owner declines a finding\nexit 1";
      if (!o.hasWhy) return "hyp: refused: a decline needs --why\nexit 1";
      return "hyp: refused: F-ONE is " + stateText(o.prior) + "\nexit 1";
    }
    case "list": {
      const lines = [];
      for (const i of ids(to)) {
        const s = o.shown[i];
        if (s === "absent") continue;
        lines.push("F-" + i + "  " + stateText(s) + "  by agent at");
        lines.push("  fact: The subject says hello.");
        if (s === "declined") lines.push("  why: not wanted");
      }
      lines.push("exit 0");
      return lines.join("\n");
    }
    case "status":
      return o.count + " not folded: hyp finding list --not-folded\nexit 0";
  }
  return "";
};
