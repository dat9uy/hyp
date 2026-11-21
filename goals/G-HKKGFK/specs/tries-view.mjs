// The view of tries.tla for the statechart: how far the grill's tries of one
// variant went, and what the report says, in the owner's words.

export const title = "hyp G-HKKGFK: the grill's tries of one variant and the report";

const n = (v) => Object.keys(v.seq ?? {}).length;
const failed = (v) => Object.values(v.tried).includes("fail");

export const state = (v) => {
  if (v.phase === "reported") return v.gap ? "Report: no example separates it" : "Report: separated";
  if (n(v) === 0) return "Variant written as a patch";
  return failed(v) ? "An example failed" : "Every try so far met";
};

export const group = (v) => (v.phase === "reported" ? "report" : "grill");
export const colors = { grill: "orange", report: "teal" };

export const step = (action, from, to) => {
  switch (action) {
    case "GrillTries": return "the grill tries the next example";
    case "GrillSpotChecks": return "the grill tries a predicted fail once more";
    case "ToolReports": return "the tool writes the report";
    case "Done": return null;
  }
  return action;
};

export const print = (action, from, to) => {
  if (action === "ToolReports") return `separated by: ${to.sep.length ? [...to.sep].sort().join(" ") : "none"}`;
  if (action === "GrillTries" || action === "GrillSpotChecks") {
    const e = Object.keys(to.tried).find((k) => from.tried[k] === "no" && to.tried[k] !== "no");
    return `${e}: ${to.tried[e]}`;
  }
  return null;
};
