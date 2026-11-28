// The view of rest.tla: a Goal B that rests on the select of a Goal A.

export const title = "A Goal that rests on another Goal's select";

const bText = (v) => {
  if (v.b === "absent") return "B not added";
  if (v.b === "written") {
    if (v.bWait === "wrong verdict") return "B written, an example wrong on A's select";
    if (v.bWait === "A closed") return "B written, A was closed";
    return "B written, waits for A's select";
  }
  return `B ${v.b}${v.bH0 === "own" ? " on its own H0" : v.bH0 !== "none" ? ` on ${v.bH0}` : ""}`;
};

export const state = (v) => `A ${v.a}${v.aSel !== "none" ? ` on ${v.aSel}` : ""}; ${bText(v)}`;

export const step = (action) => ({
  AddB: "the owner adds B --from A",
  ASelect: "A is selected",
  AUnselect: "the owner unselects A",
  AClose: "the owner closes A",
  ALand: "the owner lands A",
  BRun: "B's runner runs",
  BLand: "the owner lands B",
}[action] ?? action);

export const print = (action, from, to) => {
  if (to.addRefused && !from.addRefused) return "refused: --from names a Goal in work";
  if (action === "BLand") return "landed B";
  return bText(to);
};
