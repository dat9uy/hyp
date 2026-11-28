#!/usr/bin/env python3
"""Check that every item of an issue body has a holder of the kind its tag needs.

Usage: check_holders.py BODY.md
Reads the holder table of an issue body that draw_issue.py drew,
| Item | Tag | Holder |, and prints one line per defect, as CODE ITEM: message.
A holder is one of:

  G-XXXXXX line-N        a line of a Goal
  G-XXXXXX line-N/NAME   an example of a Goal's line
  draft NAME line-N/EX   an example in the named draft NAME, before its Goal is added
  F-XXXXXX               a finding
  owner: "WORDS"         the owner's answer, in the owner's words
  close: "WORDS"         the owner's close of the item, in the owner's words

Each tag takes these kinds of holder:

  Owner said, Proposed   a Goal line, or a close
  Fact                   a finding, a Goal line (evidence in its why), or a close
  Open                   the owner's answer, or a close
  Hypothetical           an example drafted for a line, or a close

The codes:

  table-missing   the body holds no holder table
  tag-unknown     a row's tag is not one of the tags above
  id-duplicate    a row repeats the item of an earlier row
  holder-missing  a row has no holder
  holder-form     a holder is none of the forms above
  holder-kind     a holder's kind does not fit the item's tag
  holder-unknown  a holder names a Goal, or a line of it, that the store does not hold
  holder-ended    a holder names a Goal that ended with no select: closed, or landed
                  with no selected commit

A Goal holder is read with `hyp status G`, run in the current directory, as
$HYP_ACTOR or else as the actor check-holders. Run the check at every split and
before any report of done.

Exit 0 when every item is held, 1 with defects, 2 when the body cannot be read
or hyp cannot run. Python 3 standard library only.
"""
import re, os, subprocess, sys

HEADER = ('Item', 'Tag', 'Holder')
KINDS = {
    'Owner said': ('line', 'close'),
    'Proposed': ('line', 'close'),
    'Fact': ('finding', 'line', 'close'),
    'Open': ('answer', 'close'),
    'Hypothetical': ('example', 'close'),
}
FORMS = (
    ('line', re.compile(r'^(G-[A-Z0-9]+) (line-\d+)$')),
    ('example', re.compile(r'^(G-[A-Z0-9]+) (line-\d+)/[A-Za-z0-9._-]+$')),
    ('example', re.compile(r'^draft [A-Za-z0-9._-]+ line-\d+/[A-Za-z0-9._-]+$')),
    ('finding', re.compile(r'^F-[A-Z0-9]+$')),
    ('answer', re.compile(r'^owner: ".+"$')),
    ('close', re.compile(r'^close: ".+"$')),
)
NAME = {'line': 'a Goal line', 'example': 'an example', 'finding': 'a finding',
        'answer': "the owner's answer", 'close': 'a close'}


class HypError(Exception):
    pass


def cells(line):
    """The cells of one Markdown table row, or None when the line is no row."""
    s = line.strip()
    if not (s.startswith('|') and s.endswith('|') and len(s) > 1):
        return None
    parts = re.split(r'(?<!\\)\|', s[1:-1])
    return [p.strip().replace('\\|', '|') for p in parts]


def rows(body):
    """The (item, tag, holder) rows of the holder table, or None when it has none."""
    lines = body.splitlines()
    for i, line in enumerate(lines):
        if tuple(cells(line) or ()) == HEADER:
            out = []
            for row in lines[i + 2:]:
                c = cells(row)
                if c is None:
                    break
                out.append((c + ['', '', ''])[:3])
            return out
    return None


def kind_of(holder):
    """The kind of a holder and its match, or (None, None)."""
    for kind, rx in FORMS:
        m = rx.match(holder)
        if m:
            return kind, m
    return None, None


class Store:
    """Reads Goals with hyp status, once each."""

    def __init__(self):
        self.seen = {}

    def status(self, goal):
        if goal not in self.seen:
            actor = os.environ.get('HYP_ACTOR') or 'check-holders'
            try:
                p = subprocess.run(['hyp', 'status', goal, '--actor', actor],
                                   capture_output=True, text=True)
            except OSError as e:
                raise HypError(f'cannot run hyp: {e}')
            if p.returncode == 0:
                self.seen[goal] = p.stdout
            elif f'no Goal {goal}' in p.stderr:
                self.seen[goal] = None
            else:
                raise HypError(f'hyp status {goal} exit {p.returncode}: {p.stderr.strip()}')
        return self.seen[goal]


def ended_with_no_select(st):
    lines = st.splitlines()
    if not lines or lines[0].split()[1:2] != ['ended']:
        return False
    if any(l.startswith('closed: ') for l in lines):
        return True
    return not any(l.startswith('stop: selected') for l in lines)


def check(body, store):
    """Return the defects of a body's holder table as (code, item, message) tuples."""
    table = rows(body)
    if table is None:
        return [('table-missing', '-', 'the body holds no table | Item | Tag | Holder |')]
    out, seen = [], set()
    for item, tag, holder in table:
        if item in seen:
            out.append(('id-duplicate', item, 'the item has a row above'))
        seen.add(item)
        if tag not in KINDS:
            out.append(('tag-unknown', item, f'tag "{tag}" is not one of: ' + ', '.join(KINDS)))
            continue
        if not holder:
            out.append(('holder-missing', item, 'the item has no holder'))
            continue
        kind, m = kind_of(holder)
        if kind is None:
            out.append(('holder-form', item, f'"{holder}" is not a holder of a known form'))
            continue
        if kind not in KINDS[tag]:
            out.append(('holder-kind', item, f'a {tag} item is held by '
                        + ' or '.join(NAME[k] for k in KINDS[tag]) + f', not {NAME[kind]}'))
            continue
        if kind in ('line', 'example') and holder.startswith('G-'):
            goal, line = m.group(1), m.group(2)
            st = store.status(goal)
            if st is None:
                out.append(('holder-unknown', item, f'the store holds no Goal {goal}'))
            elif not re.search(rf'^## {re.escape(line)} \(', st, re.M):
                out.append(('holder-unknown', item, f'{goal} has no {line}'))
            elif ended_with_no_select(st):
                out.append(('holder-ended', item, f'{goal} ended with no select, so it '
                            'holds no item'))
    return out


def main(argv):
    if len(argv) != 1:
        print('usage: check_holders.py BODY.md', file=sys.stderr)
        return 2
    try:
        with open(argv[0], encoding='utf-8') as f:
            body = f.read()
    except OSError as e:
        print(f'cannot read {argv[0]}: {e}', file=sys.stderr)
        return 2
    try:
        out = check(body, Store())
    except HypError as e:
        print(str(e), file=sys.stderr)
        return 2
    for code, item, msg in out:
        print(f'{code} {item}: {msg}')
    return 1 if out else 0


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))
