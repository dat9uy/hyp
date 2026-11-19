// The view of place.tla for the statechart: the Goal's state and stop, and
// what the grill's block offered after the last command.

export const title = "hyp G-1908D1: the grill's place lines in the state block";

const offered = (v) => {
  const b = [...(v.block ?? [])].map(String).sort();
  const ls = [...v.lines].map(String).sort();
  if (b.length === 0) return "grill may: (none)";
  if (b.join() === ls.join()) return "grill may place on each line";
  return "grill may place on " + b.map((x) => (x === "0" ? "<line>" : "line-" + x)).join(", ");
};

const goal = (v) => {
  if (v.selected) return "Selected";
  const ls = [...v.lines];
  const open = ls.every((l) => v.cell[l] === "live");
  if (open) return v.stop === "none" ? "Open" : "Open, stop: " + v.stop;
  const empty = ls.some((l) => v.cell[l] === "none");
  return empty ? "Written, a line holds no example" : "Written, every line placed, H0 moved";
};

export const state = (v) => [goal(v), offered(v)];

export const group = (v) => (v.selected ? "selected" : goal(v).startsWith("Open") ? "open" : "written");

export const colors = { written: "orange", open: "teal", selected: "purple" };

export const step = (action, from, to) => {
  switch (action) {
    case "GrillPlaces":
      return "the grill places an example";
    case "PlaceRefused":
      return "a place is refused";
    case "OwnerReplaces":
      return "the owner replaces the Goal; H0 moves";
    case "RunnerRunsH0":
      return "the runner runs a line on H0";
    case "RunnerRuns":
      return "the runner runs a candidate";
    case "Selects":
      return "a candidate is selected";
    case "Status":
      return "anyone reads status";
  }
  return action;
};
