#!/usr/bin/env python3
"""Draw the body of an issue from a rethink page's data file.

Usage: draw_issue.py PAGE.json
Checks the data with check_items.py of the hyp-rethinking skill, beside this
skill, and draws nothing when it finds a defect. Then prints a Markdown body:
the page's sections that hold items, in page order, each item with its id, tag,
date, text, source and findings (an item with no section goes under "Other
items"), and a holder table, | Item | Tag | Holder |, with one row for each
item and its holder empty. The grill fills each holder; check_holders.py checks
them.

Exit 0 when the body is printed, 1 when the data has defects (each printed as
CODE PATH: message), 2 when the file or check_items.py cannot be read.
Python 3 standard library only.
"""
import os, sys

sys.dont_write_bytecode = True  # keep the skill folder free of caches
HERE = os.path.dirname(os.path.abspath(__file__))
RETHINK = os.path.join(HERE, '..', '..', 'hyp-rethinking', 'scripts')

# The section names of a rethink page, as hyp-rethinking's page shape gives them.
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


def cell(text):
    """Text safe inside one cell of a Markdown table."""
    return ' '.join(str(text).split()).replace('|', '\\|')


def item_line(it, check_items):
    parts = [f"- **{it['id']}** ({it['tag']}, {it['date']}): {' '.join(it['text'].split())}"]
    if it.get('source'):
        parts.append(f"Source: `{it['source']}`.")
    fs = check_items.findings_of(it)
    if fs:
        parts.append('Findings: ' + ', '.join(f'`{f}`' for f in fs) + '.')
    return ' '.join(parts)


def draw(page, check_items):
    """Return the issue body of a page whose data has no defect."""
    items = page.get('items', [])
    out = [f"Drawn from the rethink page `{page['id']}`"
           + (f", {page['title']}" if page.get('title') else '') + '.',
           'Every item keeps its id and tag. Each item is held by the holder in the '
           'last table.', '']
    known = {n for n, _ in SECTIONS}
    for n, name in SECTIONS:
        here = [it for it in items if it.get('section') == n]
        if here:
            out += [f'## {n}. {name}', ''] + [item_line(it, check_items) for it in here] + ['']
    rest = [it for it in items if it.get('section') not in known]
    if rest:
        out += ['## Other items', ''] + [item_line(it, check_items) for it in rest] + ['']
    out += ['## Holders', '', '| Item | Tag | Holder |', '|---|---|---|']
    out += [f"| {cell(it['id'])} | {cell(it['tag'])} |  |" for it in items]
    return '\n'.join(out) + '\n'


def main(argv):
    if len(argv) != 1:
        print('usage: draw_issue.py PAGE.json', file=sys.stderr)
        return 2
    sys.path.insert(0, RETHINK)
    try:
        import check_items
    except ImportError:
        print(f'cannot read check_items.py: the hyp-rethinking skill is not at {RETHINK}',
              file=sys.stderr)
        return 2
    try:
        page = check_items.load(argv[0])
    except (OSError, ValueError) as e:
        print(f'cannot read {argv[0]}: {e}', file=sys.stderr)
        return 2
    defects = check_items.check(page)
    if defects:
        for code, path, msg in defects:
            print(f'{code} {path}: {msg}', file=sys.stderr)
        print('no body drawn: fix the data file first', file=sys.stderr)
        return 1
    sys.stdout.write(draw(page, check_items))
    return 0


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))
