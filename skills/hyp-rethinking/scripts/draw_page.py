#!/usr/bin/env python3
"""Draw a whole rethink page from its data file.

Usage: draw_page.py PAGE.json PAGE.html
Checks the data with check_items.py first and draws nothing when it finds a
defect. Then writes one self-contained HTML page: the banner, sections 1 to 10
of the page shape with each item in its section (an item with no section goes
under "Other items"), each diagram of the data file in its section, the impact
items of section 1 under "Who meets it", a diagram for each approach in section 7,
and the same diagrams small in section 8 above the side-by-side table. An item
shows its source and the findings it names.

A diagram of the data file, such as a commit graph, is drawn left to right: each
node one column right of the nodes its edges come from, with an arrow per edge.

Every diagram is drawn on one canvas: the parties the page declares, in the order
it declares them, at the same places. The concept is the one highlighted shape,
above the party where the approach places it. The steps are numbered; a step with
a from and a to party is also drawn as a numbered arrow.

Exit 0 when the page is written, 1 when the data has defects, 2 when a file cannot
be read or written. Python 3 standard library only.
"""
import html, os, sys

sys.dont_write_bytecode = True  # keep the skill folder free of caches
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import check_items  # noqa: E402

SECTIONS = (
    (1, 'The event'),
    (2, 'What it is for'),
    (3, 'Where it breaks'),
    (4, 'What gets cleared'),
    (5, 'What any answer must keep'),
    (6, 'A worked case'),
    (7, 'Three approaches'),
    (8, 'Side by side'),
    (9, 'The lean'),
    (10, 'Open questions'),
)
APPROACHES, SIDE = 7, 8
TAG_CLASS = {'Owner said': 'owner', 'Proposed': 'proposed', 'Open': 'open',
             'Fact': 'fact', 'Hypothetical': 'hypo'}
MARK = {'caught': '✓', 'missed': '✗'}
BOX_W, GAP, BOX_H = 140, 40, 44

CSS = """
:root{--bg:#fbfaf7;--fg:#1d1d1b;--muted:#6b6a66;--line:#d8d5cc;--box:#ffffff;
--hl:#c2410c;--hl-bg:#fff1e6;--owner:#1f6f43;--proposed:#2856a3;--open:#9a6700;
--fact:#4b5563;--hypo:#7c3aed}
@media (prefers-color-scheme: dark){:root:not([data-theme="light"]){--bg:#171716;
--fg:#ecebe7;--muted:#a3a19a;--line:#3a3935;--box:#22221f;--hl:#fb923c;
--hl-bg:#3a2416;--owner:#4ade80;--proposed:#7aa7ff;--open:#facc15;--fact:#cbd5e1;
--hypo:#c4b5fd}}
:root[data-theme="dark"]{--bg:#171716;--fg:#ecebe7;--muted:#a3a19a;--line:#3a3935;
--box:#22221f;--hl:#fb923c;--hl-bg:#3a2416;--owner:#4ade80;--proposed:#7aa7ff;
--open:#facc15;--fact:#cbd5e1;--hypo:#c4b5fd}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--fg);
font:16px/1.55 system-ui,-apple-system,"Segoe UI",sans-serif}
main{max-width:960px;margin:0 auto;padding:24px 16px 64px}
h1{font-size:1.6rem;margin:0 0 8px}h2{margin-top:2.2rem;font-size:1.25rem}
.banner{border:1px solid var(--line);border-radius:8px;padding:12px 14px;
background:var(--box)}
.banner p{margin:.2rem 0}
ul.items{list-style:none;padding:0;margin:0}
ul.items li{padding:8px 0;border-bottom:1px solid var(--line)}
.tag{display:inline-block;font-size:.75rem;font-weight:600;padding:1px 8px;
border-radius:999px;border:1px solid currentColor;margin-right:6px}
.t-owner{color:var(--owner)}.t-proposed{color:var(--proposed)}.t-open{color:var(--open)}
.t-fact{color:var(--fact)}.t-hypo{color:var(--hypo)}
.iid{font-family:ui-monospace,Menlo,monospace;font-size:.8rem;color:var(--muted);
margin-right:6px}
.date{font-size:.8rem;color:var(--muted);margin-right:6px}
.src,.fnd{font-size:.8rem;color:var(--muted);margin-left:6px}
.src code,.fnd code{font-family:ui-monospace,Menlo,monospace;white-space:nowrap}
.empty{color:var(--muted)}
figure{margin:12px 0}
figure svg{width:100%;height:auto;display:block}
figcaption{color:var(--muted);font-size:.9rem}
.row{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:12px}
table{border-collapse:collapse;width:100%;font-size:.92rem}
th,td{border:1px solid var(--line);padding:6px 8px;text-align:left;vertical-align:top}
.scroll{overflow-x:auto}
svg .party{fill:var(--box);stroke:var(--line);stroke-width:1.5}
svg .party.at{stroke:var(--hl)}
svg .label{fill:var(--fg);font-size:14px}
svg .concept{fill:var(--hl-bg);stroke:var(--hl);stroke-width:2}
svg .clabel{fill:var(--hl);font-size:13px;font-weight:600}
svg .step{fill:var(--fg);font-size:12px}
svg .arrow{stroke:var(--muted);stroke-width:1.5;fill:none}
svg .num{fill:var(--muted);font-size:12px;font-weight:600}
svg .head{fill:var(--muted)}
svg .node{fill:var(--box);stroke:var(--fg);stroke-width:1.5}
"""


def esc(s):
    return html.escape(str(s), quote=True)


def wrap(text, width):
    lines, cur = [], ''
    for w in text.split():
        if cur and len(cur) + 1 + len(w) > width:
            lines.append(cur)
            cur = w
        else:
            cur = f'{cur} {w}' if cur else w
    return lines + ([cur] if cur else [])


def step_parts(st):
    if isinstance(st, str):
        return st, None, None, None
    return st['text'], st.get('from'), st.get('to'), st.get('mark')


def diagram(page, ap, n, key, concept):
    """One approach on the shared canvas, as an inline SVG."""
    parties = page.get('parties', [])
    xs = {p['id']: 20 + i * (BOX_W + GAP) for i, p in enumerate(parties)}
    width = max(40 + len(parties) * BOX_W + max(len(parties) - 1, 0) * GAP, 360)
    box_y = 70
    out = []
    marker = f'arrow-{key}-{n}'
    out.append(f'<defs><marker id="{marker}" viewBox="0 0 10 10" refX="9" refY="5" '
               'markerWidth="7" markerHeight="7" orient="auto-start-reverse">'
               '<path class="head" d="M0,0 L10,5 L0,10 z"/></marker></defs>')
    for p in parties:
        x = xs[p['id']]
        cls = 'party at' if p['id'] == ap['concept_at'] else 'party'
        out.append(f'<rect class="{cls}" x="{x}" y="{box_y}" width="{BOX_W}" '
                   f'height="{BOX_H}" rx="6"/>')
        out.append(f'<text class="label" x="{x + BOX_W / 2}" y="{box_y + 27}" '
                   f'text-anchor="middle">{esc(p.get("label", p["id"]))}</text>')
    cx = xs.get(ap['concept_at'], 20)
    out.append(f'<rect class="concept" x="{cx + 10}" y="14" width="{BOX_W - 20}" '
               'height="34" rx="17"/>')
    out.append(f'<text class="clabel" x="{cx + BOX_W / 2}" y="36" '
               f'text-anchor="middle">{esc(concept)}</text>')
    out.append(f'<line class="arrow" x1="{cx + BOX_W / 2}" y1="48" '
               f'x2="{cx + BOX_W / 2}" y2="{box_y}"/>')
    y = box_y + BOX_H + 26
    for i, st in enumerate(ap.get('steps', []), 1):
        text, a, b, mark = step_parts(st)
        if a in xs and b in xs and a != b:
            x1, x2 = xs[a] + BOX_W / 2, xs[b] + BOX_W / 2
            out.append(f'<path class="arrow" d="M{x1},{box_y + BOX_H} '
                       f'C{x1},{y} {x2},{y} {x2},{box_y + BOX_H}" '
                       f'marker-end="url(#{marker})"/>')
            apex = box_y + BOX_H + 0.75 * (y - box_y - BOX_H)
            out.append(f'<text class="num" x="{(x1 + x2) / 2}" y="{apex + 15}" '
                       f'text-anchor="middle">{i}</text>')
            y += 30
    y += 8
    for i, st in enumerate(ap.get('steps', []), 1):
        text, _, _, mark = step_parts(st)
        line = f'{i}. {text}' + (f' {MARK[mark]}' if mark else '')
        for k, part in enumerate(wrap(line, max(width // 8, 30))):
            out.append(f'<text class="step" x="20" y="{y}">{esc(part)}</text>')
            y += 17
        y += 4
    label = f'Approach {ap.get("name", n)}: {concept} at ' + next(
        (p.get('label', p['id']) for p in parties if p['id'] == ap['concept_at']), '')
    return (f'<svg viewBox="0 0 {width} {y + 6}" role="img" aria-label="{esc(label)}" '
            f'xmlns="http://www.w3.org/2000/svg">' + ''.join(out) + '</svg>')


def graph(dg, n):
    """One diagram of the data file, such as a commit graph, as an inline SVG."""
    nodes = dg.get('nodes', [])
    ids = [nd['id'] for nd in nodes]
    edges = [e for e in dg.get('edges', []) if e['from'] in ids and e['to'] in ids]
    col = {i: 0 for i in ids}
    for _ in ids:  # longest path from a node with no edge in; a cycle stops at len(ids)
        for e in edges:
            if e['from'] != e['to']:
                col[e['to']] = max(col[e['to']], min(col[e['from']] + 1, len(ids) - 1))
    row, rows = {}, {}
    for i in ids:
        row[i] = rows.get(col[i], 0)
        rows[col[i]] = row[i] + 1
    # an edge that skips a column, or goes back, arcs above the boxes it passes
    arcs = [e for e in edges if col[e['to']] - col[e['from']] != 1]
    lift = max((26 + 12 * abs(col[e['to']] - col[e['from']]) for e in arcs), default=0)
    top = 20 + lift
    pos = {i: (20 + col[i] * (BOX_W + GAP), top + row[i] * (BOX_H + 30)) for i in ids}
    width = 40 + (max(col.values(), default=0) + 1) * (BOX_W + GAP) - GAP
    height = top + 20 + max(rows.values(), default=1) * (BOX_H + 30) - 30
    marker = f'arrow-graph-{n}'
    out = [f'<defs><marker id="{marker}" viewBox="0 0 10 10" refX="9" refY="5" '
           'markerWidth="7" markerHeight="7" orient="auto-start-reverse">'
           '<path class="head" d="M0,0 L10,5 L0,10 z"/></marker></defs>']
    for e in edges:
        (x1, y1), (x2, y2) = pos[e['from']], pos[e['to']]
        if e in arcs:
            high = top - 26 - 12 * abs(col[e['to']] - col[e['from']])
            a, b = x1 + BOX_W / 2, x2 + BOX_W / 2
            if a == b:
                b += BOX_W / 4
            d = f'M{a},{y1} C{a},{high} {b},{high} {b},{y2}'
        else:
            d = f'M{x1 + BOX_W},{y1 + BOX_H / 2} L{x2},{y2 + BOX_H / 2}'
        out.append(f'<path class="arrow" d="{d}" marker-end="url(#{marker})"/>')
    for nd in nodes:
        x, y = pos[nd['id']]
        out.append(f'<rect class="node" x="{x}" y="{y}" width="{BOX_W}" height="{BOX_H}" rx="6"/>')
        lines = wrap(str(nd.get('label', nd['id'])), 18)
        for k, part in enumerate(lines):
            ty = y + 27 + (k - (len(lines) - 1) / 2) * 16
            out.append(f'<text class="label" x="{x + BOX_W / 2}" y="{ty}" '
                       f'text-anchor="middle">{esc(part)}</text>')
    label = dg.get('caption') or dg['id']
    return (f'<svg viewBox="0 0 {max(width, 200)} {height}" role="img" '
            f'aria-label="{esc(label)}" xmlns="http://www.w3.org/2000/svg">'
            + ''.join(out) + '</svg>')


def graphs_html(page, n):
    out = []
    for i, dg in enumerate(page.get('diagrams', [])):
        if dg.get('section') == n:
            cap = (f'<figcaption>{esc(dg["caption"])}</figcaption>'
                   if dg.get('caption') else '')
            out.append(f'<figure id="d-{esc(dg["id"])}">{graph(dg, i)}{cap}</figure>')
    return ''.join(out)


def caption(page, ap, concept):
    if ap.get('caption'):
        return ap['caption']
    where = next((p.get('label', p['id']) for p in page.get('parties', [])
                  if p['id'] == ap['concept_at']), ap['concept_at'])
    return f'Here, {concept} lives in {where}.'


def item_li(it):
    t = it['tag']
    return (f'<li id="{esc(it["id"])}"><span class="tag t-{TAG_CLASS[t]}">{esc(t)}</span>'
            f'<span class="iid">{esc(it["id"])}</span>'
            f'<span class="date">{esc(it["date"])}</span> {esc(it["text"])}'
            + (f'<span class="src">source: <code>{esc(it["source"])}</code></span>'
               if it.get('source') else '')
            + (f'<span class="fnd">findings: ' + ', '.join(
                f'<code>{esc(f)}</code>' for f in check_items.findings_of(it)) + '</span>'
               if check_items.findings_of(it) else '')
            + '</li>')


def items_html(items):
    if not items:
        return '<p class="empty">No items yet.</p>'
    return '<ul class="items">' + ''.join(item_li(it) for it in items) + '</ul>'


def draw(page):
    concept = page.get('concept') or 'the concept'
    title = page.get('title') or f'Rethink: {page["id"]}'
    items = page.get('items', [])
    aps = page.get('approaches', [])
    by = {n: [it for it in items if it.get('section') == n] for n, _ in SECTIONS}
    other = [it for it in items if 'section' not in it]
    body = [f'<h1>{esc(title)}</h1>',
            '<div class="banner"><p><strong>Proposal, not decided.</strong></p>'
            '<p>Each item carries its source tag and the date of its words. '
            '<span class="tag t-owner">Owner said</span> '
            '<span class="tag t-proposed">Proposed</span> '
            '<span class="tag t-open">Open</span> '
            '<span class="tag t-fact">Fact</span> '
            '<span class="tag t-hypo">Hypothetical</span></p></div>']
    for n, name in SECTIONS:
        body.append(f'<h2 id="s{n}">{n}. {esc(name)}</h2>')
        figs = graphs_html(page, n)
        if n == check_items.EVENT:
            impact = [it for it in by[n] if check_items.findings_of(it)]
            rest = [it for it in by[n] if not check_items.findings_of(it)]
            if rest or not (impact or figs):
                body.append(items_html(rest))
            body.append(figs)
            if impact:
                body.append('<h3>Who meets it</h3>')
                body.append(items_html(impact))
        elif n == APPROACHES:
            for i, ap in enumerate(aps):
                body.append(f'<h3>Approach {esc(ap.get("name", i + 1))}</h3>')
                body.append(f'<figure>{diagram(page, ap, i, "big", concept)}'
                            f'<figcaption>{esc(caption(page, ap, concept))}</figcaption>'
                            '</figure>')
                table = ap.get('table') or {}
                if table:
                    body.append('<div class="scroll"><table>' + ''.join(
                        f'<tr><th>{esc(k)}</th><td>{esc(v)}</td></tr>'
                        for k, v in table.items()) + '</table></div>')
            body.append(figs)
            if by[n] or not (aps or figs):
                body.append(items_html(by[n]))
        elif n == SIDE:
            body.append('<div class="row">' + ''.join(
                f'<figure>{diagram(page, ap, i, "small", concept)}'
                f'<figcaption>Approach {esc(ap.get("name", i + 1))}</figcaption></figure>'
                for i, ap in enumerate(aps)) + '</div>')
            if aps:
                body.append(side_by_side(page, aps))
            body.append(figs)
            if by[n]:
                body.append(items_html(by[n]))
        else:
            body.append(figs)
            if by[n] or not figs:
                body.append(items_html(by[n]))
    if other:
        body.append('<h2 id="other">Other items</h2>')
        body.append(items_html(other))
    return ('<!doctype html><html lang="en"><head><meta charset="utf-8">'
            '<meta name="viewport" content="width=device-width,initial-scale=1">'
            f'<title>{esc(title)}</title><style>{CSS}</style></head>'
            f'<body><main>{"".join(body)}</main></body></html>\n')


def side_by_side(page, aps):
    labels = {p['id']: p.get('label', p['id']) for p in page.get('parties', [])}
    rows = [('Concept sits at', [labels.get(ap['concept_at'], '') for ap in aps])]
    names = []
    for ap in aps:
        for k in (ap.get('table') or {}):
            if k not in names:
                names.append(k)
    rows += [(k, [(ap.get('table') or {}).get(k, 'nothing') for ap in aps]) for k in names]
    depth = max((len(ap.get('steps', [])) for ap in aps), default=0)
    for j in range(depth):
        cells = []
        for ap in aps:
            st = ap.get('steps', [])
            if j < len(st):
                text, _, _, mark = step_parts(st[j])
                cells.append(f'{mark}: {text}' if mark else text)
            else:
                cells.append('nothing')
        rows.append((f'Step {j + 1}', cells))
    head = ''.join(f'<th>Approach {esc(ap.get("name", i + 1))}</th>' for i, ap in enumerate(aps))
    return ('<div class="scroll"><table><tr><th></th>' + head + '</tr>' + ''.join(
        f'<tr><th>{esc(k)}</th>' + ''.join(f'<td>{esc(c)}</td>' for c in cells) + '</tr>'
        for k, cells in rows) + '</table></div>')


def main(argv):
    if len(argv) != 2:
        print('usage: draw_page.py PAGE.json PAGE.html', file=sys.stderr)
        return 2
    try:
        page = check_items.load(argv[0])
    except (OSError, ValueError) as e:
        print(f'cannot read {argv[0]}: {e}', file=sys.stderr)
        return 2
    defects = check_items.check(page)
    if defects:
        for code, path, msg in defects:
            print(f'{code} {path}: {msg}')
        return 1
    try:
        with open(argv[1], 'w', encoding='utf-8') as f:
            f.write(draw(page))
    except OSError as e:
        print(f'cannot write {argv[1]}: {e}', file=sys.stderr)
        return 2
    return 0


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))
