import json, os, subprocess, sys, tempfile, unittest

HERE = os.path.dirname(os.path.abspath(__file__))
DRAW = os.path.join(HERE, '..', 'draw_issue.py')
CHECK = os.path.join(HERE, '..', 'check_holders.py')
sys.dont_write_bytecode = True
sys.path.insert(0, os.path.join(HERE, '..'))
import check_holders  # noqa: E402

PAGE = {'id': 'demo', 'title': 'Demo page', 'items': [
    {'id': 'demo-I1', 'tag': 'Proposed', 'date': '2026-11-28', 'section': 1,
     'text': 'A runner meets the refusal.', 'findings': ['F-1']},
    {'id': 'demo-I2', 'tag': 'Hypothetical', 'date': '2026-11-28', 'section': 3,
     'text': 'Two Goals | one premise.'},
    {'id': 'demo-I3', 'tag': 'Fact', 'date': '2026-11-28', 'text': 'No section.'},
], 'diagrams': [{'id': 'problem', 'section': 0, 'nodes': [{'id': 'a', 'label': 'A'}]}]}


def run(script, data, suffix):
    with tempfile.NamedTemporaryFile('w', suffix=suffix, delete=False) as f:
        f.write(data)
    try:
        return subprocess.run([sys.executable, script, f.name], capture_output=True, text=True)
    finally:
        os.unlink(f.name)


class FakeStore:
    def __init__(self, goals):
        self.goals = goals

    def status(self, goal):
        return self.goals.get(goal)


OPEN = 'G-OPEN  open  runs: 0 of 3 used\nstop: none\n\n## line-1 (gap): x\n'
CLOSED = 'G-SHUT  ended\nstop: none\nclosed: by owner\n\n## line-1 (gap): x\n'
LANDED = 'G-DONE  ended\nstop: selected, by 01X\nlanded: by owner\n\n## line-1 (gap): x\n'
STOPPED = 'G-STOP  ended\nstop: budget used\nlanded: by owner\n\n## line-1 (gap): x\n'


class DrawIssue(unittest.TestCase):
    def test_body(self):
        p = run(DRAW, json.dumps(PAGE), '.json')
        self.assertEqual(p.returncode, 0, p.stderr)
        out = p.stdout
        self.assertLess(out.index('## 1. The event'), out.index('## 3. Where it breaks'))
        self.assertLess(out.index('## 3. Where it breaks'), out.index('demo-I2'))
        self.assertIn('## Other items', out)
        self.assertIn('Findings: `F-1`.', out)
        for row in ('| demo-I1 | Proposed |  |', '| demo-I2 | Hypothetical |  |',
                    '| demo-I3 | Fact |  |'):
            self.assertIn(row, out)
        self.assertEqual(check_holders.check(out, FakeStore({}))[0][0], 'holder-missing')

    def test_defect_draws_nothing(self):
        page = dict(PAGE, items=PAGE['items'] + [PAGE['items'][0]])
        p = run(DRAW, json.dumps(page), '.json')
        self.assertEqual(p.returncode, 1)
        self.assertEqual(p.stdout, '')
        self.assertIn('id-duplicate', p.stderr)

    def test_unreadable(self):
        self.assertEqual(run(DRAW, '{', '.json').returncode, 2)


class CheckHolders(unittest.TestCase):
    def codes(self, rows, goals=None):
        body = '| Item | Tag | Holder |\n|---|---|---|\n' + ''.join(
            f'| {i} | {t} | {h} |\n' for i, t, h in rows)
        return [(c, i) for c, i, _ in check_holders.check(body, FakeStore(goals or {}))]

    def test_kinds(self):
        ok = [('a', 'Owner said', 'G-OPEN line-1'), ('b', 'Proposed', 'close: "no"'),
              ('c', 'Fact', 'F-ABC123'), ('d', 'Fact', 'G-OPEN line-1'),
              ('e', 'Open', 'owner: "yes"'), ('f', 'Hypothetical', 'G-OPEN line-1/a'),
              ('g', 'Hypothetical', 'draft d1 line-1/a'), ('h', 'Proposed', 'G-DONE line-1')]
        self.assertEqual(self.codes(ok, {'G-OPEN': OPEN, 'G-DONE': LANDED}), [])

    def test_defects(self):
        rows = [('a', 'Open', 'F-ABC123'), ('b', 'Proposed', ''), ('c', 'Odd', 'F-1'),
                ('d', 'Proposed', 'G-NONE line-1'), ('e', 'Proposed', 'G-OPEN line-9'),
                ('f', 'Proposed', 'G-SHUT line-1'), ('g', 'Proposed', 'G-STOP line-1'),
                ('h', 'Proposed', 'see above'), ('h', 'Open', 'owner: "x"')]
        goals = {'G-OPEN': OPEN, 'G-SHUT': CLOSED, 'G-STOP': STOPPED}
        self.assertEqual(self.codes(rows, goals), [
            ('holder-kind', 'a'), ('holder-missing', 'b'), ('tag-unknown', 'c'),
            ('holder-unknown', 'd'), ('holder-unknown', 'e'), ('holder-ended', 'f'),
            ('holder-ended', 'g'), ('holder-form', 'h'), ('id-duplicate', 'h')])

    def test_no_table(self):
        p = run(CHECK, 'no table here\n', '.md')
        self.assertEqual((p.returncode, p.stdout.split()[0]), (1, 'table-missing'))

    def test_escaped_pipe(self):
        self.assertEqual(check_holders.cells('| a | Open | owner: "x \\| y" |'),
                         ['a', 'Open', 'owner: "x | y"'])


if __name__ == '__main__':
    unittest.main()
