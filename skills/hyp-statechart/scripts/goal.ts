// The Goal part of the statechart tool: when the spec is in a hyp Goal's spec
// folder, goals/<G>/specs/, it runs the model once for each variant, writes
// one variant config per variant that lists exactly the rules that variant
// breaks, finds a walk of the model for each example of the Goal, predicts
// from the model alone whether each example is met under each variant, and
// writes a report of all that as goal.md.
//
// A variant is a boolean CONSTANT that the check config sets FALSE. A walk is
// a path from a start state; its output is what the view's print() returns
// for each step, joined by newlines, and it meets an example as hyp compares
// an output: every "floor: contains" is a substring, every "floor: equals"
// equals the output, and no "ceiling" is a substring.
import fs from "node:fs";
import path from "node:path";
import { readConfig } from "./frontends/config.ts";
import type { CheckedView, CheckResult, Graph, Vars } from "./ir.ts";
import { check } from "./run.ts";

const SEARCH_LIMIT = 500_000;

// ---------------------------------------------------------------- examples

type Floor = { op: "contains" | "equals"; text: string };
export type Example = { key: string; words: string; floors: Floor[]; ceilings: string[] };

// parseExample reads an example.txtar as hyp does: each section starts with a
// line "-- name --", and a body drops exactly one final newline.
export function parseExample(key: string, text: string): Example {
  const ex: Example = { key, words: "", floors: [], ceilings: [] };
  let name: string | undefined;
  let body: string[] = [];
  const flush = () => {
    if (name === undefined) return;
    const t = body.join("");
    const b = t.endsWith("\n") ? t.slice(0, -1) : t;
    if (name === "words") ex.words = b;
    else if (name === "floor: contains") ex.floors.push({ op: "contains", text: b });
    else if (name === "floor: equals") ex.floors.push({ op: "equals", text: b });
    else if (name === "ceiling") ex.ceilings.push(b);
  };
  for (const raw of text.split(/(?<=\n)/)) {
    const m = /^-- (.+) --$/.exec(raw.replace(/\n$/, ""));
    if (m) {
      flush();
      name = m[1];
      body = [];
    } else if (name !== undefined) body.push(raw);
  }
  flush();
  return ex;
}

const byLine = (a: string, b: string) => {
  const n = (s: string) => Number(/^line-(\d+)/.exec(s)?.[1] ?? 0);
  return n(a) - n(b) || (a < b ? -1 : a > b ? 1 : 0);
};

const dirs = (dir: string) => {
  try {
    return fs
      .readdirSync(dir, { withFileTypes: true })
      .filter((d) => d.isDirectory())
      .map((d) => d.name);
  } catch {
    return [];
  }
};

export function readGoal(goalDir: string) {
  const lines = dirs(goalDir)
    .filter((d) => /^line-\d+$/.test(d))
    .sort(byLine);
  const examples: Example[] = [];
  for (const line of lines)
    for (const ex of dirs(path.join(goalDir, line)).sort()) {
      const file = path.join(goalDir, line, ex, "example.txtar");
      if (fs.existsSync(file))
        examples.push(parseExample(`${line}/${ex}`, fs.readFileSync(file, "utf8")));
    }
  return { lines, examples };
}

// ----------------------------------------------------------------- matcher

// A matcher state holds, for each floor and ceiling, how far the output so
// far goes into it: for a contains floor or a ceiling the KMP position (or
// DONE once it is found), for an equals floor the length of the output so far
// (or DEAD once the output is no longer a prefix of it).
const DONE = -1;
const DEAD = -2;
type MState = { started: boolean; pos: number[] };

const kmp = (p: string) => {
  const f = new Int32Array(p.length + 1);
  f[0] = -1;
  for (let i = 1, k = -1; i <= p.length; i++) {
    while (k >= 0 && p[k] !== p[i - 1]) k = f[k];
    f[i] = ++k;
  }
  return f;
};

export class Matcher {
  items: { kind: "contains" | "equals" | "ceiling"; text: string; fail: Int32Array }[] = [];
  constructor(ex: Example) {
    for (const f of ex.floors) this.items.push({ kind: f.op, text: f.text, fail: kmp(f.text) });
    for (const c of ex.ceilings) this.items.push({ kind: "ceiling", text: c, fail: kmp(c) });
  }
  start(): MState {
    return { started: false, pos: this.items.map((it) => (it.kind !== "equals" && !it.text ? DONE : 0)) };
  }
  // feed appends one print to the output; undefined when the output can no
  // longer meet the example.
  feed(m: MState, print: string): MState | undefined {
    const s = (m.started ? "\n" : "") + print;
    const pos = m.pos.slice();
    for (let i = 0; i < this.items.length; i++) {
      const it = this.items[i];
      let p = pos[i];
      if (p === DONE) {
        if (it.kind === "ceiling") return undefined;
        continue;
      }
      if (it.kind === "equals") {
        if (p === DEAD || it.text.slice(p, p + s.length) !== s) return undefined;
        pos[i] = p + s.length;
        continue;
      }
      for (const ch of s) {
        while (p >= 0 && it.text[p] !== ch) p = it.fail[p];
        p++;
        if (p === it.text.length) {
          if (it.kind === "ceiling") return undefined;
          p = DONE;
          break;
        }
      }
      pos[i] = p;
    }
    return { started: true, pos };
  }
  accepts(m: MState) {
    return this.items.every((it, i) =>
      it.kind === "equals" ? m.pos[i] === it.text.length : it.kind === "ceiling" || m.pos[i] === DONE,
    );
  }
  key(m: MState) {
    return `${m.started ? 1 : 0}:${m.pos.join(",")}`;
  }
}

// ------------------------------------------------------------------- walks

type Edge = { action: string; from: string; to: string; print: string; label: string };
export type Model = {
  starts: string[];
  out: Map<string, Edge[]>;
  vars: Map<string, Vars>;
  key: Map<string, string>;
  name: (id: string) => string;
};

const text = (v: unknown) =>
  typeof v === "string" ? v : v === undefined || v === null ? "" : JSON.stringify(v);

export function modelOf(graph: Graph, view: CheckedView): Model {
  const out = new Map<string, Edge[]>();
  const vars = new Map<string, Vars>();
  const key = new Map<string, string>();
  const starts: string[] = [];
  for (const [id, s] of graph.states) {
    if (s.excluded) continue;
    vars.set(id, s.vars);
    key.set(id, JSON.stringify(s.vars));
    out.set(id, []);
    if (s.start) starts.push(id);
  }
  for (const st of graph.steps) {
    const from = vars.get(st.from);
    const to = vars.get(st.to);
    if (!from || !to) continue;
    const label = view.step ? text(view.step(st.action, from, to)) || st.action : st.action;
    const print = view.print ? text(view.print(st.action, from, to)) : label;
    out.get(st.from)!.push({ action: st.action, from: st.from, to: st.to, print, label });
  }
  const name = (id: string) => (view.state ? text(view.state(vars.get(id)!)) : key.get(id)!);
  return { starts, out, vars, key, name };
}

export type Walk = { edges: Edge[] } | { none: true } | { stopped: true };

// findWalk is the shortest walk of the model whose output meets the example.
export function findWalk(model: Model, ex: Example): Walk {
  const mt = new Matcher(ex);
  type Node = { id: string; m: MState; back?: Node; edge?: Edge };
  const seen = new Set<string>();
  let queue: Node[] = [];
  for (const id of model.starts) {
    const n = { id, m: mt.start() };
    const k = `${id}|${mt.key(n.m)}`;
    if (!seen.has(k)) {
      seen.add(k);
      queue.push(n);
    }
  }
  while (queue.length) {
    const next: Node[] = [];
    for (const n of queue) {
      if (mt.accepts(n.m)) {
        const edges: Edge[] = [];
        for (let x: Node | undefined = n; x?.edge; x = x.back) edges.unshift(x.edge);
        return { edges };
      }
      for (const e of model.out.get(n.id) ?? []) {
        const m = mt.feed(n.m, e.print);
        if (!m) continue;
        const k = `${e.to}|${mt.key(m)}`;
        if (seen.has(k)) continue;
        if (seen.size >= SEARCH_LIMIT) return { stopped: true };
        seen.add(k);
        next.push({ id: e.to, m, back: n, edge: e });
      }
    }
    queue = next;
  }
  return { none: true };
}

// replay follows the actions of a walk in another model and says whether
// some path with those actions gives an output that meets the example.
export function replay(model: Model, actions: string[], ex: Example): boolean {
  const mt = new Matcher(ex);
  let front = new Map<string, { id: string; m: MState }>();
  for (const id of model.starts) front.set(`${id}|${mt.key(mt.start())}`, { id, m: mt.start() });
  for (const a of actions) {
    const next = new Map<string, { id: string; m: MState }>();
    for (const { id, m } of front.values())
      for (const e of model.out.get(id) ?? []) {
        if (e.action !== a) continue;
        const m2 = mt.feed(m, e.print);
        if (m2) next.set(`${e.to}|${mt.key(m2)}`, { id: e.to, m: m2 });
      }
    front = next;
  }
  for (const { m } of front.values()) if (mt.accepts(m)) return true;
  return false;
}

export type Difference = {
  actions: string[];
  labels: string[];
  // the step where the variant first goes to another state than the model
  turn: { index: number; action: string; label: string; from: string };
  modelPrint: string | undefined;
  variantPrint: string | undefined;
};

// difference is the shortest walk, taken by the model and the variant along
// the same actions and the same prints, after which the two can print
// something else on the same action. undefined when along every walk the two
// print the same.
export function difference(base: Model, variant: Model): Difference | undefined | "stopped" {
  type Pair = { b: string; v: string; back?: Pair; action?: string; label?: string };
  const seen = new Set<string>();
  let queue: Pair[] = [];
  const add = (p: Pair) => {
    const k = `${p.b}|${p.v}`;
    if (seen.has(k)) return true;
    if (seen.size >= SEARCH_LIMIT) return false;
    seen.add(k);
    queue.push(p);
    return true;
  };
  const same: [string, string][] = [];
  for (const b of base.starts)
    for (const v of variant.starts) if (base.key.get(b) === variant.key.get(v)) same.push([b, v]);
  if (!same.length) for (const b of base.starts) for (const v of variant.starts) same.push([b, v]);
  for (const [b, v] of same) add({ b, v });
  while (queue.length) {
    const now = queue;
    queue = [];
    for (const p of now) {
      const be = base.out.get(p.b) ?? [];
      const ve = variant.out.get(p.v) ?? [];
      const actions = [...new Set([...be.map((e) => e.action), ...ve.map((e) => e.action)])].sort();
      for (const a of actions) {
        const bs = be.filter((e) => e.action === a);
        const vs = ve.filter((e) => e.action === a);
        const bp = new Set(bs.map((e) => e.print));
        const vp = new Set(vs.map((e) => e.print));
        const onlyB = [...bp].filter((x) => !vp.has(x)).sort();
        const onlyV = [...vp].filter((x) => !bp.has(x)).sort();
        if (onlyB.length || onlyV.length) {
          const chain: Pair[] = [];
          for (let x: Pair | undefined = p; x?.action; x = x.back) chain.unshift(x);
          const actions = [...chain.map((x) => x.action!), a];
          const labels = [...chain.map((x) => x.label!), (bs[0] ?? vs[0]).label];
          let turn: Difference["turn"] | undefined;
          for (let i = 0; i < chain.length; i++)
            if (base.key.get(chain[i].b) !== variant.key.get(chain[i].v)) {
              const from = chain[i].back!;
              turn = { index: i, action: chain[i].action!, label: chain[i].label!, from: base.name(from.b) };
              break;
            }
          turn ??= { index: chain.length, action: a, label: labels[chain.length], from: base.name(p.b) };
          return { actions, labels, turn, modelPrint: onlyB[0], variantPrint: onlyV[0] };
        }
        for (const eb of bs)
          for (const ev of vs)
            if (eb.print === ev.print && !add({ b: eb.to, v: ev.to, back: p, action: a, label: eb.label }))
              return "stopped";
      }
    }
  }
  return undefined;
}

// ----------------------------------------------------------------- sources

// sourcesOf reads, for each rule, the "source:" tag of the line comments
// right above its definition in the module.
export function sourcesOf(tla: string, rules: string[]) {
  const lines = tla.split(/\r?\n/);
  const found = new Map<string, string>();
  for (const rule of rules) {
    const def = lines.findIndex((l) => new RegExp(`^${rule}\\b[^=]*==`).test(l));
    if (def < 0) continue;
    for (let i = def - 1; i >= 0 && /^\s*\\\*/.test(lines[i]); i--) {
      const m = /source:\s*(.+?)\s*$/i.exec(lines[i]);
      if (m) {
        found.set(rule, m[1]);
        break;
      }
    }
  }
  return found;
}

// ---------------------------------------------------------------- variants

export function variantsOf(sections: { key: string; tokens: string[] }[]) {
  const found: string[] = [];
  for (const s of sections) {
    if (s.key !== "CONSTANT" && s.key !== "CONSTANTS") continue;
    for (let i = 0; i + 2 < s.tokens.length; i++)
      if (s.tokens[i + 1] === "=" && s.tokens[i + 2] === "FALSE" && /^\w+$/.test(s.tokens[i]))
        found.push(s.tokens[i]);
  }
  return found;
}

const RULE_KEYS = new Set(["INVARIANT", "INVARIANTS", "PROPERTY", "PROPERTIES"]);

// variantConfig is the check config with the variant TRUE; it checks the
// given invariants and properties.
export function variantConfig(
  sections: { key: string; tokens: string[] }[],
  variant: string,
  invariants: string[],
  properties: string[],
  head: string[] = [],
) {
  const out = [...head];
  for (const s of sections) {
    if (RULE_KEYS.has(s.key)) continue;
    if (s.key === "CONSTANT" || s.key === "CONSTANTS") {
      out.push("CONSTANTS");
      for (let i = 0; i < s.tokens.length;) {
        if (s.tokens[i + 1] === "=" && /^\w+$/.test(s.tokens[i]) && /^\w+$/.test(s.tokens[i + 2] ?? "")) {
          const value = s.tokens[i] === variant ? "TRUE" : s.tokens[i + 2];
          out.push(`    ${s.tokens[i]} = ${value}`);
          i += 3;
        } else {
          // a value written with more than one token, such as a set
          let j = i + 1;
          while (j < s.tokens.length && !(s.tokens[j + 1] === "=" || s.tokens[j + 1] === "<-")) j++;
          out.push(`    ${s.tokens.slice(i, j).join(" ")}`);
          i = j;
        }
      }
      continue;
    }
    out.push(`${s.key} ${s.tokens.join(" ")}`.trimEnd());
  }
  if (invariants.length) out.push("", "INVARIANTS", ...invariants.map((r) => `    ${r}`));
  if (properties.length) out.push("", "PROPERTIES", ...properties.map((r) => `    ${r}`));
  return `${out.join("\n")}\n`;
}

// ------------------------------------------------------------------ report

type Cell = { predicted: "met" | "fail" | "no walk" | "not predicted"; tried?: "met" | "fail" };

const fence = (s: string) => {
  const ticks = s.includes("```") ? "````" : "```";
  return [ticks, s, ticks].join("\n");
};

const readTried = (specDir: string, variant: string, key: string): Cell["tried"] => {
  const file = path.join(specDir, "tried", variant, `${key}.txt`);
  if (!fs.existsSync(file)) return undefined;
  const m = /verdict:\s*(met|failed|fail)/i.exec(fs.readFileSync(file, "utf8"));
  if (!m) return undefined;
  return m[1].toLowerCase() === "met" ? "met" : "fail";
};

const cellText = (c: Cell) =>
  c.tried === undefined ? c.predicted : `predicted ${c.predicted}, tried ${c.tried}`;

const ordinal = (n: number) => {
  const s = ["th", "st", "nd", "rd"];
  const v = n % 100;
  return n + (s[(v - 20) % 10] || s[v] || s[0]);
};

const isGoalConfig = (configFile: string) => {
  const parent = path.basename(path.dirname(path.resolve(configFile)));
  return parent !== "variants" && parent !== "switches";
};

// goalReport runs the Goal part when the spec has variants or is in a Goal's
// spec folder. It writes the variant configs into specs/variants/ and returns
// the report, or undefined when there is nothing to report.
export async function goalReport(
  specFile: string,
  configFile: string,
  view: CheckedView,
  base: CheckResult,
  scratch: string,
): Promise<string | undefined> {
  if (!isGoalConfig(configFile)) return undefined;
  const cfgText = fs.readFileSync(configFile, "utf8");
  const cfg = readConfig({ file: configFile, text: cfgText });
  const variants = variantsOf(cfg.sections);
  const specDir = path.dirname(path.resolve(specFile));
  const goalDir = path.basename(specDir) === "specs" ? path.dirname(specDir) : undefined;
  const goal = goalDir ? readGoal(goalDir) : { lines: [], examples: [] };
  if (!variants.length && !goalDir) return undefined;
  const rules = [...cfg.invariants, ...cfg.properties];
  const isProperty = new Set(cfg.properties);

  // Run each variant with every rule of the check config.
  const runs = await Promise.all(
    variants.map(async (v) => {
      const file = path.join(scratch, `variant-${v}.cfg`);
      fs.writeFileSync(file, variantConfig(cfg.sections, v, cfg.invariants, cfg.properties));
      try {
        return { v, result: await check(specFile, file) };
      } catch (error) {
        return { v, error: error instanceof Error ? error.message : String(error) };
      }
    }),
  );
  const broken = new Map<string, string[]>();
  const notChecked = new Map<string, string[]>();
  const models = new Map<string, Model | undefined>();
  const failed: string[] = [];
  for (const run of runs) {
    if (!("result" in run) || !run.result) {
      failed.push(`${run.v}: ${run.error}`);
      continue;
    }
    const r = run.result;
    const out: string[] = [];
    const skipped: string[] = [];
    for (const verdict of [...r.rules, ...r.properties]) {
      if (!rules.includes(verdict.rule.name) || verdict.rule.kind === "deadlock") continue;
      if (verdict.outcome === "broken") out.push(verdict.rule.name);
      if (verdict.outcome === "not checked") skipped.push(verdict.rule.name);
    }
    broken.set(run.v, rules.filter((x) => out.includes(x)));
    notChecked.set(run.v, rules.filter((x) => skipped.includes(x)));
    models.set(run.v, r.paths ? undefined : modelOf(r.graph, view));
  }

  // Write one variant config per variant.
  const written: string[] = [];
  if (variants.length) {
    const dir = path.join(specDir, "variants");
    fs.mkdirSync(dir, { recursive: true });
    for (const v of variants) {
      const list = broken.get(v);
      if (!list) continue;
      const head = [
        `\\* Variant ${v} of ${path.basename(configFile)}: the check config with ${v} TRUE.`,
        `\\* The statechart tool wrote this file. INVARIANTS and PROPERTIES list exactly`,
        `\\* the rules ${v} breaks${list.length ? "" : "; it breaks none"}.`,
      ];
      const text = variantConfig(
        cfg.sections,
        v,
        list.filter((r) => !isProperty.has(r)),
        list.filter((r) => isProperty.has(r)),
        head,
      );
      fs.writeFileSync(path.join(dir, `${v}.cfg`), text);
      written.push(path.relative(process.cwd(), path.join(dir, `${v}.cfg`)));
    }
  }

  // Walks of the check config and predictions.
  const baseModel = base.paths ? undefined : modelOf(base.graph, view);
  const walks = new Map<string, Walk | undefined>();
  for (const ex of goal.examples) walks.set(ex.key, baseModel && findWalk(baseModel, ex));
  const grid = new Map<string, Map<string, Cell>>();
  const separatedBy = new Map<string, string[]>();
  for (const v of variants) {
    const row = new Map<string, Cell>();
    const sep: string[] = [];
    const vm = models.get(v);
    for (const ex of goal.examples) {
      const w = walks.get(ex.key);
      let predicted: Cell["predicted"];
      if (!w || !("edges" in w)) predicted = w && "none" in w ? "no walk" : "not predicted";
      else if (!vm) predicted = "not predicted";
      else
        predicted = replay(
          vm,
          w.edges.map((e) => e.action),
          ex,
        )
          ? "met"
          : "fail";
      if (predicted === "fail") sep.push(ex.key);
      row.set(ex.key, { predicted, tried: readTried(specDir, v, ex.key) });
    }
    grid.set(v, row);
    separatedBy.set(v, sep);
  }

  const out: string[] = [];
  const name = goalDir ? path.basename(goalDir) : path.basename(specFile, ".tla");
  out.push(`# Variants and examples of ${name}`, "");
  out.push(
    `The check config is ${path.basename(configFile)}. A variant is a constant it sets FALSE; each variant ran with every rule of the check config.`,
  );
  if (!baseModel)
    out.push(
      "",
      "The model has more states than the chart budget, so the tool finds no walk and predicts no example.",
    );
  for (const f of failed) out.push("", `The tool could not run variant ${f}`);

  // Rules
  const sources = sourcesOf(fs.readFileSync(specFile, "utf8"), rules);
  const brokenBy = (r: string) => variants.filter((v) => broken.get(v)?.includes(r));
  out.push("", "## Rules", "");
  out.push("| Rule | Source | Broken by | An example sees it break |");
  out.push("|---|---|---|---|");
  for (const r of rules) {
    const by = brokenBy(r);
    const skipped = variants.filter((v) => notChecked.get(v)?.includes(r));
    const seen = by.some((v) => (separatedBy.get(v) ?? []).length > 0);
    const byText = by.length
      ? by.join(", ")
      : skipped.length
        ? `not checked under ${skipped.join(", ")}`
        : "not shown to fail";
    out.push(`| ${r} | ${sources.get(r) ?? "no source"} | ${byText} | ${seen ? "yes" : "no"} |`);
  }
  const unshown = rules.filter((r) => !brokenBy(r).length && !variants.some((v) => notChecked.get(v)?.includes(r)));
  out.push("");
  for (const r of unshown)
    out.push(`- ${r} is not shown to fail: no variant breaks it, so it can hold whatever the system does.`);
  const noSource = rules.filter((r) => !sources.has(r));
  out.push(`- Rules with no source: ${noSource.length ? noSource.join(", ") : "none"}.`);
  if (goalDir) {
    const named = new Set<string>();
    for (const s of sources.values()) for (const m of s.matchAll(/line-\d+/g)) named.add(m[0]);
    const unnamed = goal.lines.filter((l) => !named.has(l));
    out.push(`- Lines of the Goal that no rule names: ${unnamed.length ? unnamed.join(", ") : "none"}.`);
  }
  if (written.length) out.push(`- Variant configs written: ${written.join(", ")}.`);

  // Walks
  if (goal.examples.length) {
    out.push("", "## Walks of the examples", "");
    for (const ex of goal.examples) {
      const w = walks.get(ex.key);
      if (!w) out.push(`- ${ex.key}: not searched.`);
      else if ("none" in w)
        out.push(`- ${ex.key}: no walk of the model gives an output that meets this example.`);
      else if ("stopped" in w) out.push(`- ${ex.key}: the search stopped at ${SEARCH_LIMIT} states.`);
      else {
        const acts = w.edges.map((e) => e.action);
        out.push(`- ${ex.key}: walk ${acts.length ? acts.join(", ") : "(no step)"}`);
        if (w.edges.length) out.push("", fence(w.edges.map((e) => e.print).join("\n")), "");
      }
    }
  }

  // Grid
  if (variants.length && goal.examples.length) {
    out.push("", "## Examples by variants", "");
    out.push(
      "Each cell is predicted from the model alone: the example's walk, followed under the variant. A tried cell, from specs/tried/<variant>/<example>.txt, is the verdict `hyp try --worktree` printed on the variant's patch.",
      "",
    );
    out.push(`| Example | ${variants.join(" | ")} |`);
    out.push(`|---|${variants.map(() => "---|").join("")}`);
    for (const ex of goal.examples)
      out.push(`| ${ex.key} | ${variants.map((v) => cellText(grid.get(v)!.get(ex.key)!)).join(" | ")} |`);
    out.push(
      `| Separated by | ${variants.map((v) => (separatedBy.get(v)!.length ? separatedBy.get(v)!.join(" ") : "none")).join(" | ")} |`,
    );
    const differ: string[] = [];
    for (const v of variants)
      for (const ex of goal.examples) {
        const c = grid.get(v)!.get(ex.key)!;
        if (c.tried && (c.predicted === "met" || c.predicted === "fail") && c.tried !== c.predicted)
          differ.push(
            `- ${v} on ${ex.key}: predicted ${c.predicted}, tried ${c.tried}; the tried verdict differs from the prediction, so fix the model.`,
          );
      }
    out.push("");
    if (differ.length) out.push("Tried cells that differ from their prediction:", "", ...differ);
    else out.push("Each tried cell agrees with its prediction.");
  }

  // One part per variant that no example separates
  if (baseModel)
    for (const v of variants) {
      if ((separatedBy.get(v) ?? []).length) continue;
      const vm = models.get(v);
      if (!vm) continue;
      out.push("", `## ${v}: no example separates it`, "");
      const list = broken.get(v) ?? [];
      out.push(`Rules it breaks: ${list.length ? list.join(", ") : "none"}.`);
      const d = difference(baseModel, vm);
      if (d === "stopped") out.push("", `The search for a step where it acts otherwise stopped at ${SEARCH_LIMIT} pairs of states.`);
      else if (!d) out.push("", `Along every walk ${v} prints what the model prints.`);
      else {
        out.push(
          "",
          `The step no example takes: ${d.turn.label} (${d.turn.action}) from "${d.turn.from}", the ${ordinal(d.turn.index + 1)} step of the walk ${d.actions.join(", ")}.`,
          "",
          `On the last step of that walk (${d.labels[d.labels.length - 1]}) the model prints:`,
          "",
          d.modelPrint === undefined ? "(nothing: the step cannot happen)" : fence(d.modelPrint),
          "",
          `and ${v} prints:`,
          "",
          d.variantPrint === undefined ? "(nothing: the step cannot happen)" : fence(d.variantPrint),
        );
      }
    }
  return out.join("\n");
}
