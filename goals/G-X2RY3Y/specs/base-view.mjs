// The view of base.tla: one Goal in work whose rows may stand on newer code.

export const title = "A Goal moves onto newer code only when nothing it measures changed";

const name = (s) => (s.length ? s.slice().sort().join("+") : "H0");

export const state = (v) => {
  if (!v.inWork) return "stopped after a refused move";
  const sel = v.sel.length === 0 ? "open" : `selected on ${name(v.sel.filter((x) => x !== "g"))}`;
  const ref = v.refused.length ? ", a move refused" : "";
  return `${sel}${ref}`;
};

export const step = (action, from, to) => {
  if (action === "EnvChange") return "the environment changes";
  if (to.belowRefused && !from.belowRefused) return "a run above the select is refused";
  if (to.refused.length > from.refused.length) return "a run on newer code is refused";
  return "a run is recorded";
};

export const print = (action, from, to) => {
  if (action === "EnvChange") return "";
  if (to.belowRefused && !from.belowRefused) return "refused: the Goal is selected";
  if (to.refused.length > from.refused.length) {
    const r = to.refused.find((x) => !from.refused.some((y) => JSON.stringify(y) === JSON.stringify(x)));
    const base = name(r.base);
    return r.why === "environment"
      ? `refused: the environment changed; base ${base} and H0 both measure differently`
      : `refused: an example measures differently on base ${base}`;
  }
  const base = to.prNamesBase ? ` on base ${name(to.sel.filter((x) => x !== "g"))}` : "";
  return `recorded: selected${base}`;
};
