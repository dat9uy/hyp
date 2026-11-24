// The view of learned.tla: what hyp prints to an agent for each act and read.

export const title = "hyp G-T30ZAY, a finding is free until folded into a Goal in work";

const ended = (v, i) => v.f[i] === "folded" && v.into[i] !== "none" && v.gs[v.into[i]] === "ended";
const label = (v, i) => {
  const s = v.f[i];
  if (s === "none") return null;
  if (s === "notfolded") return v.draft[i] ? "not folded, drafted" : "not folded";
  if (s === "refuted") return "refuted";
  return ended(v, i) ? "folded into an ended Goal" : "folded into a Goal in work";
};

export const state = (v) => {
  const parts = Object.keys(v.f).map((i) => label(v, i)).filter(Boolean);
  return parts.length === 0 ? "no finding" : [...new Set(parts)].sort().join(" + ");
};

export const group = (v) => {
  const s = Object.keys(v.f).map((i) => v.f[i]);
  if (s.includes("refuted")) return "refuted";
  if (s.includes("folded")) return "folded";
  return s.includes("notfolded") ? "free" : "none";
};
export const colors = { none: "blue", free: "teal", folded: "purple", refuted: "orange" };

export const step = (action, from, to) => {
  const o = to.out;
  switch (action) {
    case "AgentAdds": return o.ok ? (to.draft[Object.keys(to.f).find((i) => from.f[i] === "none" && to.f[i] !== "none")] ? "an agent records a finding with a draft" : "an agent records a finding") : "an agent records a finding [refused]";
    case "AgentFolds": return o.ok ? "an agent folds it into a Goal" : "an agent folds it [refused]";
    case "AgentRefutes": return o.ok ? "an agent refutes it" : null;
    case "GoalEnds": return "the Goal ends";
    case "GoalPlaced": case "GoalSelected": return null;
    case "AgentLists": case "AgentReadsStatus": case "AgentReadsGoal": return null;
  }
  return action;
};

const stateText = (v, o, i) => {
  const s = o.shown[i];
  if (s === "notfolded") return "not folded";
  if (s === "refuted") return "refuted";
  if (s === "folded-ended") return "folded into G-" + v.into[i] + ", ended";
  if (s === "folded") return "folded into G-" + v.into[i];
  return null;
};

export const print = (action, from, to) => {
  const o = to.out;
  switch (o.cmd) {
    case "add": return o.ok ? "recorded F-" : "refused: finding add needs a Goal";
    case "fold":
      if (o.ok) return "folded into G-";
      return { missing: "refused: no Goal", ended: "refused: the Goal has ended",
               refuted: "refused: the finding is refuted", folded: "refused: the finding is folded",
               frozen: "refused: the finding's Goal has ended" }[o.why];
    case "refute": return o.ok ? "refuted" : "refused: the finding is refuted";
    case "list": {
      const lines = [];
      for (const i of Object.keys(to.f)) {
        const t = stateText(to, o, i);
        if (t) lines.push("F-" + i + "  " + t);
      }
      if (o.drafts) lines.push("draft by the agent: hyp add goal");
      if (o.acts) lines.push("hyp finding add --fact");
      return lines.join("\n");
    }
    case "status": {
      const lines = [];
      for (const g of Object.keys(o.words)) if (o.words[g]) lines.push("G-" + g + " words");
      if (o.count >= 0) lines.push(o.count + " not folded: hyp finding list --not-folded");
      return lines.join("\n");
    }
    case "status-g": {
      const lines = [];
      if (o.goalWhy) lines.push("why:");
      for (const i of Object.keys(to.f)) {
        const t = stateText(to, o, i);
        if (t) lines.push("F-" + i + "  " + t);
      }
      return lines.join("\n");
    }
  }
  return "";
};
