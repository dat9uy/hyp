// The view of prompt.tla: what hyp status G prints in each state of the
// Goal, and what a reader finds in the skill and the owner guide.

export const title = "hyp G-4GV6W7, hyp status G prints the runner's Start prompt of an open Goal";

const label = { written: "Written", open: "Open", selected: "Selected", landed: "Landed", closed: "Closed" };

export const state = (v) => label[v.state];

export const group = (v) => v.state;

export const colors = { written: "orange", open: "teal", selected: "purple", landed: "blue", closed: "pink" };

export const step = (action) => {
  switch (action) {
    case "GrillPlaces": return "the grill places the examples";
    case "RunnerSelects": return "the runner selects";
    case "OwnerUnselects": return "the owner unselects";
    case "OwnerLands": return "the owner lands";
    case "OwnerCloses": return "the owner closes";
    case "OwnerReadsStatus": return null;
    case "ReaderSearchesCopy": return null;
    case "ReaderReadsDocs": return null;
  }
  return action;
};

const header = (s) => {
  switch (s) {
    case "landed": return ["G  ended", "landed: by owner, land E at T"];
    case "closed": return ["G  ended", "closed: by owner, close E at T"];
  }
  return [`G  ${s}  runs: 0 of 3 used`];
};

const prompt = (o) => {
  const g = o.filled ? "G" : "<G>";
  const actor = o.filled ? "runner-g" : "<actor>";
  const lines = [
    "The runner's Start prompt, ready to copy:",
    `Run the hyp Goal ${g} as its runner.`,
    `- Your actor is ${actor}. Set HYP_ACTOR=${actor} and HYP_ROLE=runner each`,
  ];
  if (o.closes) lines.push("  Closes #N for each of these issues: <issues>.");
  return lines;
};

export const print = (action, from, to) => {
  const o = to.out;
  if (o.cmd === "status") return [...header(o.state), ...(o.prompt ? prompt(o) : [])].join("\n");
  if (o.cmd === "copy") return o.copy ? "skills/hyp-grill/run-rules.md" : "no copy";
  if (o.cmd === "docs") {
    const v = o.says ? "says" : "silent";
    return ["skills/hyp-grill/SKILL.md", "skills/hyp-grill/run-rules.md", "docs/owner-guide.md"].map((f) => `${f}: ${v}`).join("\n");
  }
  return "";
};
