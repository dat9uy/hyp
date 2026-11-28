#!/usr/bin/env python3
"""Check the data file of a rethink page before the page is drawn from it.

Usage: check_items.py PAGE.json
Prints one line per defect, as CODE PATH: message, where PATH names the place in
the file, such as items[2] or approaches[0].concept_at. The codes:

  id-missing      the page or an item has no id
  id-duplicate    an item repeats the id of an earlier item
  id-prefix       an item id does not start with the page id and a dash
  tag-unknown     an item tag is not one of the tags in TAGS
  date-missing    an item has no date
  text-missing    an item has no text
  section-unknown an item names a section that is not 1 to 10, or a diagram one
                  that is not 0 to 10
  picture-missing the page has no diagram of section 0, the picture of the problem
                  that is drawn above section 1
  event-too-long  section 1 holds more than MAX_EVENT items besides one impact item
  bar-number-missing a bar, or the limit of a bar diagram, has no label or no number
  impact-missing  section 1 has items, and none of them names the findings that
                  show who meets the problem
  purpose-not-fact an item of section 2, "What it is for", is not tagged Fact
  source-missing  an item of section 2 has no source
  party-unknown   an approach places its concept at a party the page does not declare
  node-unknown    a diagram edge names a node the diagram does not declare
  field-unknown   a field this format does not hold, such as a "replaced" marker
  shape           a value of the wrong type, such as items that is not a list

Exit 0 when the file is whole, 1 with defects, 2 when the file cannot be read or
is not JSON. Python 3 standard library only.
"""
import json, sys

TAGS = ('Owner said', 'Proposed', 'Open', 'Fact', 'Hypothetical')
PAGE_FIELDS = ('id', 'title', 'concept', 'items', 'parties', 'approaches', 'diagrams')
ITEM_FIELDS = ('id', 'tag', 'date', 'text', 'section', 'source', 'findings')
SECTION_COUNT = 10
EVENT, PURPOSE = 1, 2
PICTURE = 0  # the section of the picture drawn above section 1
MAX_EVENT = 6
PARTY_FIELDS = ('id', 'label')
APPROACH_FIELDS = ('name', 'concept_at', 'steps', 'caption', 'table')
STEP_FIELDS = ('text', 'from', 'to', 'mark')
DIAGRAM_FIELDS = ('id', 'section', 'caption', 'nodes', 'edges', 'bars', 'limit')
BAR_FIELDS = ('label', 'value')
NODE_FIELDS = ('id', 'label')
EDGE_FIELDS = ('from', 'to')


def _text(v):
    return isinstance(v, str) and v.strip() != ''


def _unknown(obj, allowed, path, out):
    for k in obj:
        if k not in allowed:
            out.append(('field-unknown', f'{path}.{k}' if path else k,
                        f'"{k}" is not a field of this format'))


def _list(obj, key, out, path=''):
    v = obj.get(key, [])
    if not isinstance(v, list):
        where = f'{path}.{key}' if path else key
        out.append(('shape', where, f'{key} must be a list'))
        return []
    return v


def findings_of(it):
    """The finding ids an item names, or an empty list."""
    v = it.get('findings')
    return [f for f in v if _text(f)] if isinstance(v, list) else []


def _section(v, first=1):
    return v in range(first, SECTION_COUNT + 1) and not isinstance(v, bool)


def _number(v):
    return isinstance(v, (int, float)) and not isinstance(v, bool)


def check(page):
    """Return the defects of a page's data as (code, path, message) tuples."""
    out = []
    if not isinstance(page, dict):
        return [('shape', '$', 'the file must hold one JSON object')]
    _unknown(page, PAGE_FIELDS, '', out)
    pid = page.get('id')
    if not _text(pid):
        out.append(('id-missing', 'id', 'the page has no id'))
        pid = None
    seen = {}
    for i, it in enumerate(_list(page, 'items', out)):
        p = f'items[{i}]'
        if not isinstance(it, dict):
            out.append(('shape', p, 'an item must be an object'))
            continue
        _unknown(it, ITEM_FIELDS, p, out)
        iid = it.get('id')
        if not _text(iid):
            out.append(('id-missing', p, 'the item has no id'))
        else:
            if iid in seen:
                out.append(('id-duplicate', p, f'id "{iid}" is also the id of {seen[iid]}'))
            else:
                seen[iid] = p
            if pid and not iid.startswith(pid + '-'):
                out.append(('id-prefix', p, f'id "{iid}" does not start with "{pid}-"'))
        if it.get('tag') not in TAGS:
            out.append(('tag-unknown', p, f'tag {json.dumps(it.get("tag"))} is not one of: '
                        + ', '.join(TAGS)))
        if not _text(it.get('date')):
            out.append(('date-missing', p, 'the item has no date'))
        if not _text(it.get('text')):
            out.append(('text-missing', p, 'the item has no text'))
        if 'section' in it and not _section(it['section']):
            out.append(('section-unknown', p, f'section {json.dumps(it["section"])} '
                        f'is not 1 to {SECTION_COUNT}'))
        if 'source' in it and not _text(it['source']):
            out.append(('shape', f'{p}.source', 'source is a text, such as a path or an id'))
        if 'findings' in it and not (isinstance(it['findings'], list)
                                     and all(_text(f) for f in it['findings'])):
            out.append(('shape', f'{p}.findings', 'findings is a list of finding ids'))
        if it.get('section') == PURPOSE:
            if it.get('tag') != 'Fact':
                out.append(('purpose-not-fact', p, 'an item of section 2, "What it is for", '
                            'is not tagged Fact'))
            if not _text(it.get('source')):
                out.append(('source-missing', p, 'an item of section 2, "What it is for", '
                            'has no source'))
    items = page.get('items') if isinstance(page.get('items'), list) else []
    event = [it for it in items if isinstance(it, dict) and it.get('section') == EVENT]
    if event and not any(findings_of(it) for it in event):
        out.append(('impact-missing', 'items', 'section 1 holds no impact item: an item that '
                    'says who meets the problem, with the findings that show it'))
    rest = len(event) - (1 if any(findings_of(it) for it in event) else 0)
    if rest > MAX_EVENT:
        out.append(('event-too-long', 'items', f'section 1 holds {rest} items besides the '
                    f'impact item; the most is {MAX_EVENT}'))
    parties = set()
    for i, pa in enumerate(_list(page, 'parties', out)):
        p = f'parties[{i}]'
        if not isinstance(pa, dict):
            out.append(('shape', p, 'a party must be an object'))
            continue
        _unknown(pa, PARTY_FIELDS, p, out)
        if not _text(pa.get('id')):
            out.append(('id-missing', p, 'the party has no id'))
        else:
            parties.add(pa['id'])
    for i, ap in enumerate(_list(page, 'approaches', out)):
        p = f'approaches[{i}]'
        if not isinstance(ap, dict):
            out.append(('shape', p, 'an approach must be an object'))
            continue
        _unknown(ap, APPROACH_FIELDS, p, out)
        if ap.get('concept_at') not in parties:
            out.append(('party-unknown', f'{p}.concept_at',
                        f'{json.dumps(ap.get("concept_at"))} is not a declared party'))
        if 'table' in ap and not (isinstance(ap['table'], dict)
                                  and all(isinstance(v, str) for v in ap['table'].values())):
            out.append(('shape', f'{p}.table', 'table must map each row name to its text'))
        steps = ap.get('steps', [])
        if not isinstance(steps, list):
            out.append(('shape', f'{p}.steps', 'steps must be a list'))
            continue
        for j, st in enumerate(steps):
            sp = f'{p}.steps[{j}]'
            if isinstance(st, str):
                continue
            if not isinstance(st, dict) or not _text(st.get('text')):
                out.append(('shape', sp, 'a step is a text, or an object with a text'))
                continue
            _unknown(st, STEP_FIELDS, sp, out)
            for end in ('from', 'to'):
                if end in st and st[end] not in parties:
                    out.append(('party-unknown', f'{sp}.{end}',
                                f'{json.dumps(st[end])} is not a declared party'))
            if st.get('mark') not in (None, 'caught', 'missed'):
                out.append(('shape', f'{sp}.mark', 'mark is caught or missed'))
    for i, dg in enumerate(_list(page, 'diagrams', out)):
        p = f'diagrams[{i}]'
        if not isinstance(dg, dict):
            out.append(('shape', p, 'a diagram must be an object'))
            continue
        _unknown(dg, DIAGRAM_FIELDS, p, out)
        if not _text(dg.get('id')):
            out.append(('id-missing', p, 'the diagram has no id'))
        if not _section(dg.get('section'), PICTURE):
            out.append(('section-unknown', f'{p}.section', f'section '
                        f'{json.dumps(dg.get("section"))} is not 0 to {SECTION_COUNT}'))
        for j, br in enumerate(_list(dg, 'bars', out, p)):
            bp = f'{p}.bars[{j}]'
            if not isinstance(br, dict):
                out.append(('shape', bp, 'a bar is an object with a label and a value'))
                continue
            _unknown(br, BAR_FIELDS, bp, out)
            if not _text(br.get('label')) or not _number(br.get('value')):
                out.append(('bar-number-missing', bp, 'a bar has a label and a number'))
        lim = dg.get('limit')
        if 'limit' in dg:
            if not isinstance(lim, dict):
                out.append(('shape', f'{p}.limit', 'limit is an object with a label and a value'))
            else:
                _unknown(lim, BAR_FIELDS, f'{p}.limit', out)
                if not _text(lim.get('label')) or not _number(lim.get('value')):
                    out.append(('bar-number-missing', f'{p}.limit',
                                'the limit has a label and a number'))
        nodes = set()
        for j, nd in enumerate(_list(dg, 'nodes', out, p)):
            np = f'{p}.nodes[{j}]'
            if not isinstance(nd, dict) or not _text(nd.get('id')):
                out.append(('shape', np, 'a node is an object with an id'))
                continue
            _unknown(nd, NODE_FIELDS, np, out)
            if nd['id'] in nodes:
                out.append(('id-duplicate', np, f'node id "{nd["id"]}" is declared twice'))
            nodes.add(nd['id'])
        for j, ed in enumerate(_list(dg, 'edges', out, p)):
            ep = f'{p}.edges[{j}]'
            if not isinstance(ed, dict):
                out.append(('shape', ep, 'an edge is an object with a from and a to'))
                continue
            _unknown(ed, EDGE_FIELDS, ep, out)
            for end in ('from', 'to'):
                if ed.get(end) not in nodes:
                    out.append(('node-unknown', f'{ep}.{end}',
                                f'{json.dumps(ed.get(end))} is not a node of this diagram'))
    diagrams = page.get('diagrams') if isinstance(page.get('diagrams'), list) else []
    if not any(isinstance(d, dict) and d.get('section') == PICTURE
               and not isinstance(d.get('section'), bool) for d in diagrams):
        out.append(('picture-missing', 'diagrams', 'no diagram of section 0: the page opens '
                    'with a picture of the problem, drawn above section 1'))
    return out


def load(path):
    """Read a data file. Raise OSError or ValueError when it cannot be read."""
    with open(path, encoding='utf-8') as f:
        return json.load(f)


def main(argv):
    if len(argv) != 1:
        print('usage: check_items.py PAGE.json', file=sys.stderr)
        return 2
    try:
        page = load(argv[0])
    except (OSError, ValueError) as e:
        print(f'cannot read {argv[0]}: {e}', file=sys.stderr)
        return 2
    out = check(page)
    for code, path, msg in out:
        print(f'{code} {path}: {msg}')
    return 1 if out else 0


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))
