// The owner's words for the states and steps of the unselect model.
type V = {
  st: string; cand: boolean; pending: boolean; used: number; since: number;
  unselected: boolean; fresh: boolean; newCell: boolean; by: string; landedEver: boolean;
  last: string; lastOK: boolean; lastWho: string; prevSt: string;
};

export const title = "hyp G-VDG87A: the owner's unselect";

export const state = (v: V) => {
  if (v.st === "landed") return "Landed";
  if (v.st === "selected") return v.unselected ? "Selected again" : "Selected";
  if (!v.unselected) return v.cand ? "Open, a candidate" : "Open, no candidate";
  if (!v.cand) return "Unselected, no candidate";
  return v.newCell ? "Unselected, a candidate with a new cell" : "Unselected, the old candidate only";
};

export const group = (v: V) => (v.st === "landed" ? "landed" : v.st === "selected" ? "selected" : v.unselected ? "unselected" : "open");
export const colors = { open: "teal", selected: "blue", unselected: "orange", landed: "purple" };

export const step = (action: string, from: V, to: V) => {
  if (action === "Unselect") {
    if (!to.lastOK) return "hyp refuses an unselect";
    return to.lastWho === "owner" ? "the owner unselects" : "someone else unselects";
  }
  if (action === "RunnerSelect") return to.lastOK ? "the runner selects" : "hyp refuses the runner's select";
  return ({
    RunNewRow: "the runner runs a new commit",
    RunOldRow: from.pending ? "the runner runs the new version on the row" : "the runner reruns the row",
    Place: "the grill places a new version",
    OwnerSelect: "the owner selects",
    Land: "the owner lands",
  } as Record<string, string>)[action] ?? action;
};
