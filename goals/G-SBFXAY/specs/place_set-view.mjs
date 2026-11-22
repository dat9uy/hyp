// The view of place_set.tla: the Goal's state, and what the last place did.

export const title = "hyp G-SBFXAY: the grill places a Goal's whole example set";

const goal = (v) => (v.selected ? "Selected" : v.open ? "Open" : "Written");

const lastPlace = (v) => {
  switch (v.last) {
    case "placed": return "set placed";
    case "part": return "part of a set placed";
    case "refused": return "set refused";
  }
  return "nothing placed yet";
};

export const state = (v) => [goal(v), lastPlace(v)];

export const group = (v) => goal(v).toLowerCase();

export const colors = { written: "orange", open: "teal", selected: "purple" };

export const step = (action, from, to) => {
  switch (action) {
    case "GrillEditsExamples": return null;
    case "GrillEditsModel": return null;
    case "GrillPlaces":
      if (to.last === "placed") return "the grill places the set";
      if (to.last === "part") return "the grill places the set [an example disagrees]";
      return "the grill places the set [a cause to refuse]";
    case "RunnerRuns": return "the runner runs a Hypothesis";
    case "Selects": return "a candidate is selected";
    case "Status": return null;
  }
  return action;
};

const names = (s) => [...s].sort().join(", ");

export const print = (action, from, to) => {
  if (action !== "GrillPlaces") return "";
  const grill = to.block === "per line" ? "grill may: hyp place G line-1/<example>" : "grill may: hyp place G";
  if (to.last === "placed") return "placed G\n" + grill;
  if (to.last === "part") return "placed part of G; refused: " + names(to.named) + "\n" + grill;
  return "refused: " + names(to.named) + "\n" + grill;
};
