// The owner's words for the states and steps of the Spec guard model.
type V = { files: string; named: string[]; chart: boolean; outcome: string; names: string[] };

const folder = (v: V) => {
  if (v.files === "none") return "no spec";
  if (v.files === "plain") return "a plain spec";
  const all = v.named.length === 2 ? "a switch for every rule" : "a rule with no switch";
  return `a model, ${all}, ${v.chart ? "a chart" : "no chart"}`;
};

export const title = "hyp G-G2R3VG: the Spec guard reads the model";

export const state = (v: V) => {
  if (v.outcome === "selected") return "Selected";
  if (v.outcome === "refused") {
    if (v.files === "none") return "Refused: no spec";
    return v.names.includes("chart") ? "Refused: names the missing chart" : "Refused: names each rule with no switch";
  }
  return `Spec folder: ${folder(v)}`;
};

export const group = (v: V) => (v.outcome === "editing" ? "commit" : v.outcome === "selected" ? "selected" : "refused");
export const colors = { commit: "teal", selected: "blue", refused: "orange" };

export const step = (action: string) =>
  ({
    CommitPlain: "the runner commits a plain spec",
    CommitModel: "the runner commits a model",
    CommitSwitch: "the runner commits a switch config",
    CommitChart: "the runner commits the chart",
    Select: "the runner asks hyp to select",
    Again: "the runner commits again",
    Done: null,
  })[action] ?? action;
