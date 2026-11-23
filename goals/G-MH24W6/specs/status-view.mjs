// The view of status.tla: what hyp prints to the owner for each read and act.

export const title = "hyp G-MH24W6, hyp status lists the Goals in work and counts the ended ones";

const ended = (s) => s === "landed" || s === "closed";

export const state = (v) => {
  const ids = Object.keys(v.g);
  const work = ids.filter((i) => !ended(v.g[i])).length;
  const done = ids.length - work;
  const goals = done === 0 ? "no Goal ended" : work === 0 ? "every Goal ended" : "some Goals ended";
  return goals;
};

export const group = (v) => {
  const ids = Object.keys(v.g);
  const done = ids.filter((i) => ended(v.g[i])).length;
  return done === 0 ? "work" : done === ids.length ? "ended" : "mixed";
};
export const colors = { work: "teal", mixed: "purple", ended: "blue" };

export const step = (action, from, to) => {
  switch (action) {
    case "GrillPlaces": return null;
    case "RunnerSelects": return null;
    case "OwnerLands": return "the owner lands a Goal";
    case "OwnerCloses": return "the owner closes a Goal";
    case "ReplayRefuses": return null;
    case "LineUnread": return null;
    case "OwnerReadsStatus": return null;
    case "OwnerReadsStatusAll": return null;
    case "OwnerReadsGoal": return null;
    case "OwnerReadsUsage": return null;
  }
  return action;
};

const block = (v, i) => {
  const s = v.g[i];
  if (s === "landed") return `G-${i}  ended\nstop: selected\nlanded: by owner\nwaits: none`;
  if (s === "closed") return `G-${i}  ended\nstop: none\nclosed: by owner\nwaits: none`;
  const stop = s === "selected" ? "selected" : "none";
  return `G-${i}  ${s}  runs: 0 of 3 used\nstop: ${stop}\nwaits: none`;
};

const status = (to) => {
  const o = to.out;
  const lines = [];
  for (const i of Object.keys(to.g)) if (o.shown[i]) lines.push(block(to, i));
  if (o.refusedShown) lines.push("refused at replay:\n  E: an event replay refused");
  if (o.unreadShown) lines.push("not read at replay:\n  .tree/events/x.jsonl line 1: not read");
  if (o.noWork) lines.push("no Goal is in work.");
  if (o.countLine) {
    const n = o.count === 1 ? "1 ended Goal" : `${o.count} ended Goals`;
    lines.push(`${n} not shown; hyp status --all lists every Goal.`);
  }
  return lines.join("\n");
};

export const print = (action, from, to) => {
  const o = to.out;
  switch (o.cmd) {
    case "place": return `placed G-${o.arg}`;
    case "select": return `selected G-${o.arg}`;
    case "land": return `landed G-${o.arg}`;
    case "close": return `closed G-${o.arg}`;
    case "merge": return "";
    case "status": case "status-all": return status(to);
    case "status-all-refused": return "flag provided but not defined: -all";
    case "status-g": {
      const b = block(to, o.arg);
      return o.full ? b : b.split("\n")[0];
    }
    case "help": case "spec":
      return o.usageAll ? "  status [G] [--role ROLE] [--all]" : "  status [G] [--role ROLE]";
  }
  return "";
};
