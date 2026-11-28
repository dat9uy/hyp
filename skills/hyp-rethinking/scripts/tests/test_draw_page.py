import json, os, subprocess, sys, tempfile, unittest

HERE = os.path.dirname(__file__)
CHECK = os.path.join(HERE, '..', 'check_items.py')
DRAW = os.path.join(HERE, '..', 'draw_page.py')
PAGE = os.path.join(HERE, '..', 'check_page.py')

WHOLE = {
    'id': 'demo', 'concept': 'the key',
    'items': [
        {'id': 'demo-I1', 'tag': 'Owner said', 'date': '2026-11-27',
         'text': 'Keep the words of the owner.', 'section': 1},
        {'id': 'demo-I2', 'tag': 'Fact', 'date': '2026-11-28', 'text': 'Items are data.'},
        {'id': 'demo-I3', 'tag': 'Fact', 'date': '2026-11-28', 'section': 1,
         'text': 'A runner meets the refusal.', 'source': 'F-1', 'findings': ['F-1']},
        {'id': 'demo-I4', 'tag': 'Fact', 'date': '2026-11-28', 'section': 2,
         'text': 'The key names a Goal.', 'source': 'docs/glossary.md'},
    ],
    'diagrams': [{'id': 'commits', 'section': 1, 'caption': 'The fix from main',
                  'nodes': [{'id': 'h0', 'label': 'H0'}, {'id': 'a', 'label': 'add A'},
                            {'id': 'fix', 'label': 'fix'}],
                  'edges': [{'from': 'h0', 'to': 'a'}, {'from': 'a', 'to': 'fix'},
                            {'from': 'h0', 'to': 'fix'}]}],
    'parties': [{'id': 'owner', 'label': 'Owner'}, {'id': 'tool', 'label': 'Tool'}],
    'approaches': [
        {'name': n, 'concept_at': at,
         'steps': ['The owner writes.', {'text': 'The tool reads.', 'from': 'owner',
                                         'to': 'tool', 'mark': 'caught'}]}
        for n, at in (('A', 'owner'), ('B', 'tool'), ('C', 'owner'))
    ],
}


def write(data):
    d = tempfile.mkdtemp()
    p = os.path.join(d, 'page.json')
    with open(p, 'w') as f:
        f.write(data if isinstance(data, str) else json.dumps(data))
    return d, p


def run(*args):
    r = subprocess.run([sys.executable, *args], capture_output=True, text=True)
    return r.returncode, r.stdout


class CheckItems(unittest.TestCase):
    def test_whole_file_exits_0(self):
        _, p = write(WHOLE)
        self.assertEqual(run(CHECK, p), (0, ''))

    def test_each_defect_has_its_code_and_path(self):
        page = json.loads(json.dumps(WHOLE))
        page['items'] = [
            {'tag': 'Proposed', 'date': 'd', 'text': 't'},
            {'id': 'demo-I1', 'tag': 'Proposed', 'date': 'd', 'text': 't'},
            {'id': 'demo-I1', 'tag': 'Proposed', 'date': 'd', 'text': 't'},
            {'id': 'other-I3', 'tag': 'Proposed', 'date': 'd', 'text': 't'},
            {'id': 'demo-I4', 'tag': 'Later', 'text': 't'},
            {'id': 'demo-I5', 'tag': 'Proposed', 'date': 'd', 'text': 't', 'replaced_by': 'x'},
        ]
        page['approaches'][0]['concept_at'] = 'nobody'
        _, p = write(page)
        code, out = run(CHECK, p)
        self.assertEqual(code, 1)
        for want in ('id-missing items[0]', 'id-duplicate items[2]', 'id-prefix items[3]',
                     'tag-unknown items[4]', 'date-missing items[4]',
                     'field-unknown items[5].replaced_by',
                     'party-unknown approaches[0].concept_at'):
            self.assertIn(want, out)

    def test_event_and_purpose_rules(self):
        page = json.loads(json.dumps(WHOLE))
        page['items'][2]['findings'] = []
        page['items'].append({'id': 'demo-I5', 'tag': 'Proposed', 'date': 'd', 'text': 't',
                              'section': 2})
        page['diagrams'][0]['edges'].append({'from': 'a', 'to': 'nowhere'})
        _, p = write(page)
        code, out = run(CHECK, p)
        self.assertEqual(code, 1)
        for want in ('impact-missing items:', 'purpose-not-fact items[4]',
                     'source-missing items[4]', 'node-unknown diagrams[0].edges[3].to'):
            self.assertIn(want, out)
        self.assertNotIn('items[3]', out)

    def test_a_page_with_no_event_items_needs_no_impact(self):
        page = json.loads(json.dumps(WHOLE))
        page['items'] = [it for it in page['items'] if it.get('section') != 1]
        _, p = write(page)
        self.assertEqual(run(CHECK, p), (0, ''))

    def test_unreadable_file_exits_2(self):
        self.assertEqual(run(CHECK, '/no/such/page.json')[0], 2)
        _, p = write('{not json')
        self.assertEqual(run(CHECK, p)[0], 2)


class DrawPage(unittest.TestCase):
    def test_drawn_page_passes_check_page(self):
        d, p = write(WHOLE)
        out = os.path.join(d, 'page.html')
        self.assertEqual(run(DRAW, p, out)[0], 0)
        self.assertEqual(run(PAGE, out), (0, ''))
        page = open(out, encoding='utf-8').read()
        self.assertEqual(page.count('<svg'), 7)
        for want in ('demo-I1', 'Owner said', '2026-11-27', 'Keep the words of the owner.',
                     '2. What it is for</h2>', '<h3>Who meets it</h3>', 'docs/glossary.md',
                     'findings: <code>F-1</code>', 'The fix from main'):
            self.assertIn(want, page)
        event = page[page.index('1. The event'):page.index('2. What it is for')]
        self.assertIn('add A', event)

    def test_defective_data_draws_nothing(self):
        page = json.loads(json.dumps(WHOLE))
        del page['items'][0]['date']
        d, p = write(page)
        out = os.path.join(d, 'page.html')
        code, text = run(DRAW, p, out)
        self.assertEqual(code, 1)
        self.assertIn('date-missing items[0]', text)
        self.assertFalse(os.path.exists(out))

    def test_unwritable_page_exits_2(self):
        _, p = write(WHOLE)
        self.assertEqual(run(DRAW, p, '/no/such/dir/page.html')[0], 2)


if __name__ == '__main__':
    unittest.main()
