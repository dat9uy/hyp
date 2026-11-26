// The view of meetable.tla: the Goal's state, and what the place did.

export const title = "hyp G-MNSE33: place refuses an example no build can meet";

const goal = (v) => (v.goal === "open" ? "Open" : "Written");

const lastPlace = (v) => {
  switch (v.last) {
    case "placed": return "set placed";
    case "refused": return "set refused";
  }
  return "nothing placed yet";
};

export const state = (v) => [goal(v), lastPlace(v)];

export const group = (v) => goal(v).toLowerCase();

export const colors = { written: "orange", open: "teal" };

export const step = (action, from, to) => {
  if (action === "Status") return null;
  if (action === "GrillFixes") return "the grill fixes the set";
  if (action !== "GrillPlaces") return action;
  if (to.last === "placed") return "the grill places the set [some output meets each example]";
  return "the grill places the set [an example no output meets]";
};

export const print = (action, from, to) => {
  if (action !== "GrillPlaces") return "";
  if (to.last === "placed") return "placed G-ONE\nplace exit: 0\nG-ONE  open";
  const keys = [...to.named].sort().map((k) => "line-1/" + k + ": Meetable");
  return ["refused: none of the example set of G-ONE is placed", ...keys,
    "place exit: 1", "G-ONE  written"].join("\n");
};
