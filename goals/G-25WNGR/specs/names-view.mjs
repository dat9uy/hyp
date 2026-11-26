// The view of names.tla: what hyp prints for each act and read.

export const title = "hyp G-25WNGR, a Goal names its findings";

const one = (v) => Object.keys(v.f)[0];
const fstate = (v) => {
  const i = one(v);
  const s = v.f[i];
  if (s === "none") return "no finding";
  if (s === "refuted" || s === "declined") return s;
  const named = Object.keys(v.gs).filter((g) =>
    v.own[g].includes?.(i) || v.ex[g].includes?.(i) || v.old[g].includes?.(i));
  if (named.some((g) => v.gs[g] === "work")) return "named by a Goal in work";
  if (named.some((g) => v.gs[g] === "landed")) return "settled";
  if (named.length > 0) return "ended, not settled";
  return "waiting";
};

export const state = (v) => fstate(v);
export const group = (v) => {
  const s = fstate(v);
  return s === "no finding" ? "none" : s === "waiting" ? "wait" : s === "named by a Goal in work" ? "work" : "end";
};
export const colors = { none: "blue", wait: "teal", work: "purple", end: "orange" };

export const step = (action, from, to) => {
  const o = to.out;
  if (!o.ok) return null;
  switch (action) {
    case "OwnerAddsGoal": return o.ids.length ? "the owner adds a Goal that names it" : null;
    case "GrillPlaces": return o.ids.length ? "the grill places an example that names it" : null;
    case "OwnerLands": return to.gs[o.g] === "landed" ? "the owner lands with a select" : "the owner lands with no select";
    case "OwnerCloses": return "the owner closes the Goal";
    case "AgentAdds": return "an agent records a finding";
    case "AgentRefutes": return "an agent refutes it";
    case "OwnerDeclines": return "the owner declines it";
  }
  return null;
};

const text = (s) => ({ folded: "folded into G-", settled: "settled", endednot: "ended, not settled",
  waiting: "not folded", refuted: "refuted", declined: "declined", nottrace: "not in this clone's trace" }[s] ?? s);

const refusal = (o) => ({
  unknown: "hyp: refused: F-ONE is not in this clone's trace",
  final: "hyp: refused: F-ONE is refuted or declined",
  twowork: "hyp: refused: F-ONE is named by another Goal in work",
  uncovered: "hyp: refused: F-ONE is named by G-" + o.g + " and no example names it",
}[o.why]);

export const print = (action, from, to) => {
  const o = to.out;
  switch (o.cmd) {
    case "add": return o.ok ? "recorded G-" + o.g + (o.ids.length ? "\nfindings: F-ONE" : "") : refusal(o) + "\nexit 1";
    case "place": return o.ok ? "placed G-" + o.g : refusal(o) + "\nexit 1";
    case "end": return to.gs[o.g] === "landed" ? "landed G-" + o.g : "ended G-" + o.g;
    case "addf": return "recorded F-ONE  not folded";
    case "refute": return "recorded F-ONE  refuted";
    case "decline": return o.ok ? "recorded F-ONE  declined" : "hyp: refused: F-ONE is folded\nexit 1";
    case "fold": return o.ok ? "recorded F-ONE  folded into G-" + o.g : "hyp: unknown command: finding fold\nexit 2";
    case "list": {
      const s = o.shown[one(to)];
      return (s === "absent" ? "" : "F-ONE  " + text(s) + (s === "folded" ? "" : "") + "\n") + "exit 0";
    }
    case "status": return o.count + " not folded: hyp finding list --not-folded\nexit 0";
    case "statusg": {
      const lines = ["G-" + o.g];
      if (o.ids.length) {
        lines.push("findings:");
        const s = o.shown[one(to)];
        lines.push("  F-ONE  " + text(s));
      }
      lines.push("exit 0");
      return lines.join("\n");
    }
  }
  return "";
};
