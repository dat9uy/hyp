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
  section-unknown an item names a section that is not 1 to 9
  party-unknown   an approach places its concept at a party the page does not declare
  field-unknown   a field this format does not hold, such as a "replaced" marker
  shape           a value of the wrong type, such as items that is not a list

Exit 0 when the file is whole, 1 with defects, 2 when the file cannot be read or
is not JSON. Python 3 standard library only.
"""
import json, sys

TAGS = ('Owner said', 'Proposed', 'Open', 'Fact', 'Hypothetical')
PAGE_FIELDS = ('id', 'title', 'concept', 'items', 'parties', 'approaches')
ITEM_FIELDS = ('id', 'tag', 'date', 'text', 'section')
PARTY_FIELDS = ('id', 'label')
APPROACH_FIELDS = ('name', 'concept_at', 'steps', 'caption', 'table')
STEP_FIELDS = ('text', 'from', 'to', 'mark')


def _text(v):
    return isinstance(v, str) and v.strip() != ''


def _unknown(obj, allowed, path, out):
    for k in obj:
        if k not in allowed:
            out.append(('field-unknown', f'{path}.{k}' if path else k,
                        f'"{k}" is not a field of this format'))


def _list(page, key, out):
    v = page.get(key, [])
    if not isinstance(v, list):
        out.append(('shape', key, f'{key} must be a list'))
        return []
    return v


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
        if 'section' in it and it['section'] not in range(1, 10):
            out.append(('section-unknown', p, f'section {json.dumps(it["section"])} is not 1 to 9'))
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
