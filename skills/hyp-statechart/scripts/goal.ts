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
export type Example = {
  key: string;
  words: string;
  run: string;
  floors: Floor[];
  ceilings: string[];
  // the example's Probe, probe/run, when it has one
  probe?: string;
};

// parseExample reads an example.txtar as hyp does: each section starts with a
// line "-- name --", and a body drops exactly one final newline.
export function parseExample(key: string, text: string): Example {
  const ex: Example = { key, words: "", run: "", floors: [], ceilings: [] };
  let name: string | undefined;
  let body: string[] = [];
  const flush = () => {
    if (name === undefined) return;
    const t = body.join("");
    const b = t.endsWith("\n") ? t.slice(0, -1) : t;
    if (name === "words") ex.words = b;
    else if (name === "run") ex.run = b;
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
      if (!fs.existsSync(file)) continue;
      const example = parseExample(`${line}/${ex}`, fs.readFileSync(file, "utf8"));
      const probe = path.join(goalDir, line, ex, "probe", "run");
      if (fs.existsSync(probe)) example.probe = fs.readFileSync(probe, "utf8");
      examples.push(example);
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

// ---------------------------------------------------------- command walks

// commandsOf reads, for each action, the "command:" tag of the line comments
// right above its definition in the module: the command of the system that
// takes that step, such as "\* command: hyp select".
export function commandsOf(tla: string) {
  const lines = tla.split(/\r?\n/);
  const found = new Map<string, string>();
  for (let d = 0; d < lines.length; d++) {
    const def = /^(\w+)\s*(?:\([^)]*\))?\s*==/.exec(lines[d]);
    if (!def) continue;
    for (let i = d - 1; i >= 0 && /^\s*\\\*/.test(lines[i]); i--) {
      const m = /command:\s*(.+?)\s*$/i.exec(lines[i]);
      if (m) {
        found.set(def[1], m[1]);
        break;
      }
    }
  }
  return found;
}

// shellCommands splits a shell script into its simple commands, in the order
// they are written: it drops comments and the bodies of here-documents,
// splits at ; & | && || and newlines outside quotes, and drops leading
// assignments and keywords. Each command is a list of words with their
// quotes removed; raw keeps the text, so that a reference to $HYP_RUN shows.
export function shellCommands(script: string) {
  const cmds: { words: string[]; raw: string }[] = [];
  const lines = script.replace(/\\\n/g, " ").split("\n");
  const keywords = new Set(["if", "then", "else", "elif", "fi", "do", "done", "while", "until", "!", "{", "}", "(", ")", "exec", "command"]);
  let doc: { tag: string; strip: boolean } | undefined;
  for (const line of lines) {
    if (doc) {
      if ((doc.strip ? line.replace(/^\t+/, "") : line) === doc.tag) doc = undefined;
      continue;
    }
    let words: string[] = [];
    let word: string | undefined;
    let raw = "";
    let quote: string | undefined;
    const end = () => {
      if (word !== undefined) words.push(word);
      word = undefined;
    };
    const flush = () => {
      end();
      let i = 0;
      while (i < words.length && (keywords.has(words[i]) || /^\w+=/.test(words[i]))) i++;
      words = words.slice(i);
      if (words.length) cmds.push({ words, raw: raw.trim() });
      words = [];
      raw = "";
    };
    for (let i = 0; i < line.length; i++) {
      const c = line[i];
      if (quote) {
        raw += c;
        if (c === quote) quote = undefined;
        else if (c === "\\" && quote === '"' && i + 1 < line.length) word = (word ?? "") + line[++i];
        else word = (word ?? "") + c;
        continue;
      }
      if (c === "'" || c === '"') {
        quote = c;
        raw += c;
        word ??= "";
        continue;
      }
      if (c === "#" && word === undefined) break;
      const heredoc = /^<<(-?)\s*(['"]?)(\w+)\2/.exec(line.slice(i));
      if (heredoc) {
        doc = { tag: heredoc[3], strip: heredoc[1] === "-" };
        end();
        raw += heredoc[0];
        i += heredoc[0].length - 1;
        continue;
      }
      if (c === ";" || c === "&" || c === "|") {
        flush();
        continue;
      }
      raw += c;
      if (/\s/.test(c)) end();
      else word = (word ?? "") + c;
    }
    flush();
  }
  return cmds;
}

// A command step of an example: the action of the model that a command of
// its Probe or of its run section takes. run is true for a command of the
// run section: only its prints are matched against the example.
type CommandStep = { action: string; command: string; run: boolean };

export type CommandWalk =
  | { steps: CommandStep[] }
  // a command of a program the model holds steps for, with no step of its own
  | { unknown: string[] };

// commandWalk is the walk of the commands an example's Probe and its run
// section call, in the order they call them. A Probe command that refers to
// $HYP_RUN stands for the commands of the run section; with no Probe, the
// walk is the run section's commands. A command of another program, such as
// echo or mkdir, takes no step of the model.
export function commandWalk(ex: Example, commands: Map<string, string>): CommandWalk {
  const tags = [...commands].map(([action, command]) => ({ action, words: command.split(/\s+/) }));
  tags.sort((a, b) => b.words.length - a.words.length);
  const programs = new Set(tags.map((t) => t.words[0]));
  const steps: CommandStep[] = [];
  const unknown: string[] = [];
  const take = (script: string, run: boolean) => {
    for (const c of shellCommands(script)) {
      if (!run && /\$\{?HYP_RUN\b/.test(c.raw)) {
        take(ex.run, true);
        continue;
      }
      const tag = tags.find((t) => t.words.every((w, i) => c.words[i] === w));
      if (tag) steps.push({ action: tag.action, command: tag.words.join(" "), run });
      else if (programs.has(c.words[0])) {
        const name = c.words.slice(0, 2).join(" ");
        if (!unknown.includes(name)) unknown.push(name);
      }
    }
  };
  if (ex.probe !== undefined) take(ex.probe, false);
  else take(ex.run, true);
  return unknown.length ? { unknown } : { steps };
}

// writesBeforeRun reports whether an example's Probe writes a file before it
// calls its run section: a command with an output redirect (> or >>, not to
// /dev/null or to a file descriptor) or a program that writes files, such as
// cp or touch. Such a write can set up the example's case, and it takes no
// step of the model.
const WRITERS = new Set(["cp", "mv", "touch", "tee", "mkdir", "ln", "rm", "install", "dd", "tar", "unzip"]);
export function writesBeforeRun(ex: Example): boolean {
  if (ex.probe === undefined) return false;
  for (const c of shellCommands(ex.probe)) {
    if (/\$\{?HYP_RUN\b/.test(c.raw)) return false;
    const unquoted = c.raw.replace(/'[^']*'|"(?:[^"\\]|\\.)*"/g, "");
    if (/>>?\s*(?![&>]|\/dev\/null\b)\S/.test(unquoted)) return true;
    if (WRITERS.has(c.words[0])) return true;
  }
  return false;
}

// pathOf is a path of the model that takes the steps of a command walk, from
// a start state; undefined when the model cannot take them.
export function pathOf(model: Model, steps: CommandStep[]): Edge[] | undefined {
  const seen = new Set<string>();
  const go = (id: string, i: number): Edge[] | undefined => {
    if (i === steps.length) return [];
    const k = `${id}|${i}`;
    if (seen.has(k)) return undefined;
    seen.add(k);
    for (const e of model.out.get(id) ?? []) {
      if (e.action !== steps[i].action) continue;
      const rest = go(e.to, i + 1);
      if (rest) return [e, ...rest];
    }
    return undefined;
  };
  for (const id of model.starts) {
    const p = go(id, 0);
    if (p) return p;
  }
  return undefined;
}

// follow takes the steps of a command walk in a model, from each start state
// and along each edge of each step's action, feeds only the prints of the run
// section's steps to the matcher, and says whether some path meets the
// example.
export function follow(model: Model, steps: CommandStep[], ex: Example): boolean {
  const mt = new Matcher(ex);
  let front = new Map<string, { id: string; m: MState }>();
  for (const id of model.starts) front.set(`${id}|${mt.key(mt.start())}`, { id, m: mt.start() });
  for (const s of steps) {
    const next = new Map<string, { id: string; m: MState }>();
    for (const { id, m } of front.values())
      for (const e of model.out.get(id) ?? []) {
        if (e.action !== s.action) continue;
        const m2 = s.run ? mt.feed(m, e.print) : m;
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

// sourceLines returns the lines a source names, or undefined when the source is
// neither a list of G-XXXXXX:line-N nor `hyp spec, guard NAME`.
const REF = "G-[A-Z0-9]{6}:line-\\d+";
export function sourceLines(source: string): { goal: string; line: string }[] | undefined {
  if (/^hyp spec, guard \S.*$/.test(source)) return [];
  if (!new RegExp(`^${REF}(\\s*,\\s*${REF})*$`).test(source)) return undefined;
  return source.split(",").map((r) => {
    const [goal, line] = r.trim().split(":");
    return { goal, line };
  });
}

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

type Plan =
  | { kind: "shortest"; walk: Walk }
  | { kind: "unknown"; commands: string[] }
  | { kind: "cannot"; steps: CommandStep[] }
  // the model has more than one start case and the Probe writes files before
  // its run section, which take no step: the tool cannot see the start case
  | { kind: "start"; starts: number }
  | { kind: "commands"; steps: CommandStep[]; edges: Edge[]; outside: boolean };

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

// goalSpecDir is the Goal spec folder a check writes into: the spec's own
// folder when it is a Goal's spec folder, goals/G/specs/; else, for a model
// kept outside any Goal, such as the kept model in goals/model/, the Goal
// spec folder that holds the output folder, when one does.
const goalSpecDir = (specFile: string, outDir: string) => {
  const own = path.dirname(path.resolve(specFile));
  if (path.basename(own) === "specs") return own;
  const parts = path.resolve(outDir).split(path.sep);
  for (let i = parts.length - 1; i >= 2; i--)
    if (parts[i] === "specs" && parts[i - 2] === "goals") return parts.slice(0, i + 1).join(path.sep);
  return own;
};

// goalReport runs the Goal part when the spec has variants or is in a Goal's
// spec folder, or its output folder is in one. It writes the variant configs
// into that folder's variants/ and returns
// the report, or undefined when there is nothing to report. In a Goal's spec
// folder, refused lists each rule of the check config that no variant config
// it wrote names: hyp select refuses a commit whose check config checks such
// a rule.
export async function goalReport(
  specFile: string,
  configFile: string,
  outDir: string,
  view: CheckedView,
  base: CheckResult,
  scratch: string,
): Promise<{ text: string; refused: string[]; badSource: string[] } | undefined> {
  if (!isGoalConfig(configFile)) return undefined;
  const cfgText = fs.readFileSync(configFile, "utf8");
  const cfg = readConfig({ file: configFile, text: cfgText });
  const variants = variantsOf(cfg.sections);
  const specDir = goalSpecDir(specFile, outDir);
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

  // Walks of the check config and predictions. When the module tags its
  // actions with the commands of the system, each example is predicted from
  // the walk of the commands its Probe and its run section call; a model with
  // no command tag falls back to the shortest walk whose prints meet it.
  const baseModel = base.paths ? undefined : modelOf(base.graph, view);
  const commands = commandsOf(fs.readFileSync(specFile, "utf8"));
  const plans = new Map<string, Plan>();
  for (const ex of goal.examples) {
    if (!baseModel) continue;
    if (!commands.size) {
      plans.set(ex.key, { kind: "shortest", walk: findWalk(baseModel, ex) });
      continue;
    }
    const cw = commandWalk(ex, commands);
    if ("unknown" in cw) {
      plans.set(ex.key, { kind: "unknown", commands: cw.unknown });
      continue;
    }
    if (baseModel.starts.length > 1 && writesBeforeRun(ex)) {
      plans.set(ex.key, { kind: "start", starts: baseModel.starts.length });
      continue;
    }
    const edges = pathOf(baseModel, cw.steps);
    if (!edges) {
      plans.set(ex.key, { kind: "cannot", steps: cw.steps });
      continue;
    }
    const floorsOnly = { ...ex, ceilings: [] };
    const outside = ![baseModel, ...variants.map((v) => models.get(v))].some(
      (m) => m && follow(m, cw.steps, floorsOnly),
    );
    plans.set(ex.key, { kind: "commands", steps: cw.steps, edges, outside });
  }
  const grid = new Map<string, Map<string, Cell>>();
  const separatedBy = new Map<string, string[]>();
  for (const v of variants) {
    const row = new Map<string, Cell>();
    const sep: string[] = [];
    const vm = models.get(v);
    for (const ex of goal.examples) {
      const plan = plans.get(ex.key);
      let predicted: Cell["predicted"] = "not predicted";
      if (plan?.kind === "shortest") {
        const w = plan.walk;
        if ("none" in w) predicted = "no walk";
        else if ("edges" in w && vm)
          predicted = replay(
            vm,
            w.edges.map((e) => e.action),
            ex,
          )
            ? "met"
            : "fail";
      } else if (plan?.kind === "commands" && !plan.outside && vm && pathOf(vm, plan.steps))
        predicted = follow(vm, plan.steps, ex) ? "met" : "fail";
      // A tried verdict stands over the prediction: a tried fail separates
      // the variant, and a tried met does not, whatever the model predicts.
      const tried = readTried(specDir, v, ex.key);
      if ((tried ?? predicted) === "fail") sep.push(ex.key);
      row.set(ex.key, { predicted, tried });
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
    for (const s of sources.values())
      for (const ref of sourceLines(s) ?? []) if (ref.goal === name) named.add(ref.line);
    const unnamed = goal.lines.filter((l) => !named.has(l));
    out.push(`- Lines of the Goal that no rule names: ${unnamed.length ? unnamed.join(", ") : "none"}.`);
  }
  const badSource = rules.filter((r) => sources.has(r) && !sourceLines(sources.get(r)!));
  for (const r of badSource)
    out.push(
      `- ${r} has the source "${sources.get(r)}", which is neither a list of G-XXXXXX:line-N nor "hyp spec, guard NAME".`,
    );
  if (written.length) out.push(`- Variant configs written: ${written.join(", ")}.`);
  const named = new Set(variants.flatMap((v) => broken.get(v) ?? []));
  const refused = goalDir ? rules.filter((r) => !named.has(r)) : [];
  if (refused.length)
    out.push(
      `- hyp select refuses this check config: no variant config names ${refused.join(", ")}. Add a variant that breaks each such rule, or take the rule out of the check config.`,
    );

  // Walks
  const toTry: string[] = [];
  if (goal.examples.length) {
    out.push("", "## Walks of the examples", "");
    if (baseModel && !commands.size)
      out.push(
        "The module tags no action with a `\\* command:` comment, so each walk is the shortest walk whose prints meet the example, and it can set up another case than the example's Probe.",
        "",
      );
    for (const ex of goal.examples) {
      const plan = plans.get(ex.key);
      if (!plan) out.push(`- ${ex.key}: not searched.`);
      else if (plan.kind === "shortest") {
        const w = plan.walk;
        if ("none" in w) out.push(`- ${ex.key}: no walk of the model gives an output that meets this example.`);
        else if ("stopped" in w) out.push(`- ${ex.key}: the search stopped at ${SEARCH_LIMIT} states.`);
        else {
          const acts = w.edges.map((e) => e.action);
          out.push(`- ${ex.key}: walk ${acts.length ? acts.join(", ") : "(no step)"}`);
          if (w.edges.length) out.push("", fence(w.edges.map((e) => e.print).join("\n")), "");
        }
      } else if (plan.kind === "unknown") {
        const names = plan.commands.join(", ");
        out.push(
          `- ${ex.key}: its Probe or run section calls ${names}, which the model has no step for, so the tool predicts no verdict.`,
        );
        toTry.push(`- ${ex.key}: the model has no step for ${names}.`);
      } else if (plan.kind === "start") {
        out.push(
          `- ${ex.key}: the model has ${plan.starts} start cases, and the example's Probe writes files before its run section, which takes no step of the model, so the tool cannot see the example's start case and predicts no verdict.`,
        );
        toTry.push(`- ${ex.key}: the tool cannot see its start case.`);
      } else if (plan.kind === "cannot") {
        const acts = plan.steps.map((s) => s.action).join(", ");
        out.push(
          `- ${ex.key}: the model cannot take the walk of its commands, ${acts}, so the tool predicts no verdict.`,
        );
        toTry.push(`- ${ex.key}: the model cannot take the walk ${acts}.`);
      } else {
        const part = (run: boolean) =>
          plan.steps
            .filter((s) => s.run === run)
            .map((s) => s.action)
            .join(", ") || "none";
        out.push(
          `- ${ex.key}: walk ${plan.steps.map((s) => s.action).join(", ") || "(no step)"} (setup: ${part(false)}; run section: ${part(true)})`,
        );
        const prints = plan.edges.filter((_, i) => plan.steps[i].run).map((e) => e.print);
        out.push("", "The run section prints, on one path of the walk:", "", fence(prints.join("\n")), "");
        if (plan.outside) {
          out.push(
            `${ex.key} is outside the model's prints: no print of its walk gives its floors, under the check config or any variant, so the tool predicts no verdict.`,
            "",
          );
          toTry.push(`- ${ex.key}: outside the model's prints.`);
        }
      }
    }
  }
  if (!baseModel) for (const ex of goal.examples) toTry.push(`- ${ex.key}: the model has more states than the chart budget.`);
  if (toTry.length && variants.length)
    out.push(
      "",
      "## Examples to try on every variant",
      "",
      "The tool predicts no verdict for these examples. Try each one on the patch of every variant with `hyp try --worktree`.",
      "",
      ...toTry,
    );

  // Grid
  if (variants.length && goal.examples.length) {
    out.push("", "## Examples by variants", "");
    out.push(
      "Each cell is predicted from the model alone: the example's walk, followed under the variant. A tried cell, from specs/tried/<variant>/<example>.txt, is the verdict `hyp try --worktree` printed on the variant's patch. An example separates a variant when its tried verdict is a fail, or when it was not tried and the tool predicts it to fail.",
      "",
    );
    out.push(`| Example | ${variants.join(" | ")} |`);
    out.push(`|---|${variants.map(() => "---|").join("")}`);
    for (const ex of goal.examples)
      out.push(`| ${ex.key} | ${variants.map((v) => cellText(grid.get(v)!.get(ex.key)!)).join(" | ")} |`);
    out.push(
      `| Separated by | ${variants.map((v) => (separatedBy.get(v)!.length ? separatedBy.get(v)!.join(" ") : baseModel ? "none" : "not predicted")).join(" | ")} |`,
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
  return { text: out.join("\n"), refused, badSource };
}
