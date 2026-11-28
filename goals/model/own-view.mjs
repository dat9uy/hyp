// The view of own.tla: Own path at one run of a Goal.

export const title = "A run is judged by what its runner wrote";

export const state = (v) =>
  v.result === "none" ? "before the run"
  : v.result === "recorded" ? "the run is recorded"
  : v.result === "two bases" ? "the run is refused by Same base"
  : "the run is refused by Own path";

export const step = (action, from, to) =>
  to.result === "recorded" ? "the runner runs its commit"
  : to.result === "two bases" ? "the runner runs a commit on two bases"
  : to.result === "other goal" ? "the runner runs a commit that writes another Goal's folder"
  : "the runner runs a commit that writes a folder of no Goal";

export const print = (action, from, to) =>
  to.result === "recorded" ? "run exit 0"
  : to.result === "two bases" ? "Same base: stands on two bases\nrun exit 1"
  : to.result === "other goal" ? "Own path: a path of another Goal\nrun exit 1"
  : "Own path: a folder of no Goal\nrun exit 1";
