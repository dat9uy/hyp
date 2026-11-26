// The view of predict.tla: what the statechart tool prints for the fixture
// example, in the case its Probe and run section set up.
export const title = "The statechart tool predicts the fixture example";

const words = {
  setup: "the Probe adds and folds, the run lists",
  setupPrint: "the ceiling is a setup print",
  probeFloor: "the floor is the run section's echo",
  noStep: "the Probe calls toy refute",
  plain: "no setup, one command",
};

const result = (s) => {
  if (s.includes("outside the model")) return "outside the model's prints";
  if (s.includes("no step")) return "names toy refute";
  if (s.includes("no walk")) return "no walk";
  if (s.includes("| fail |")) return "predicts fail";
  if (s.includes("| met |")) return "predicts met";
  return "";
};

export const state = (v) =>
  v.ran ? `${words[v.case]}: ${result(v.shown)}` : `fixture set up: ${words[v.case]}`;

export const step = () => "the tool runs";

export const print = (action, from, to) => to.shown;
