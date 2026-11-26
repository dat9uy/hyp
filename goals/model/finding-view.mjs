// The view of finding.tla: the state of the finding, and what hyp prints.

export const title = "hyp, the acts on one finding";

export const state = (v) => (v.st === "none" ? "no finding" : v.st);

export const step = (action) =>
  ({
    FindingAdd: "an agent adds the finding",
    AddGoal: "the owner adds a Goal that names it",
    Land: "the owner lands the Goal with a select",
    Close: "the owner closes the Goal",
    FindingRefute: "an agent refutes it",
    FindingDecline: "the owner declines it",
    FindingList: "anyone lists the findings",
  })[action] ?? action;

export const print = (action, from, to) => to.out;
