import os, subprocess, sys, tempfile, unittest

SCRIPT = os.path.join(os.path.dirname(__file__), '..', 'check_page.py')

def run(page):
    with tempfile.NamedTemporaryFile('w', suffix='.html', delete=False) as f:
        f.write(page)
    r = subprocess.run([sys.executable, SCRIPT, f.name], capture_output=True, text=True)
    os.unlink(f.name)
    return r.returncode, r.stdout

CLEAN = ('<body><h2>1. One</h2><p>See section 2.</p>'
         '<h2>2. Two</h2><p><span class="tag t-owner">Owner said</span> Short text.</p>'
         '<svg id="a"><text>' + 'word ' * 40 + '</text></svg></body>')

class CheckPage(unittest.TestCase):
    def test_clean_page_exits_0(self):
        self.assertEqual(run(CLEAN), (0, ''))

    def test_each_defect_is_reported(self):
        page = ('<body><h2>1. One</h2><h2>3. Three</h2><p>See section 9.</p>'
                '<svg id="x"></svg><svg id="x"></svg><p>' + 'word ' * 30 + '.</p></body>')
        code, out = run(page)
        self.assertEqual(code, 1)
        self.assertIn('section 3 follows section 1', out)
        self.assertIn('missing section 9', out)
        self.assertIn('duplicate id "x"', out)
        self.assertIn('30 words', out)

    def test_quotation_counts_as_one_word(self):
        quote = '"' + 'word ' * 30 + 'end. More words here."'
        self.assertEqual(run('<body><p>The owner said ' + quote + ' on the event.</p></body>'), (0, ''))
        curly = '\u201c' + 'word ' * 30 + '\u201d'
        self.assertEqual(run('<body><p>Owner: ' + curly + '.</p></body>'), (0, ''))
        own = 'word ' * 24 + '"short" and more.'
        code, out = run('<body><p>' + own + '</p></body>')
        self.assertEqual(code, 1)
        self.assertIn('27 words', out)

    def test_approaches_need_diagrams(self):
        bare = '<body><h2>7. Approaches</h2><svg></svg><p>A.</p><h2>8. Side</h2><p>T.</p></body>'
        code, out = run(bare)
        self.assertEqual(code, 1)
        self.assertIn('section 7: 1 diagrams, needs 3', out)
        self.assertIn('section 8: 0 diagrams, needs 1', out)
        full = ('<body><h2>7. Approaches</h2><svg></svg><svg></svg><pre class="mermaid">a</pre>'
                '<h2>8. Side</h2><svg></svg></body>')
        self.assertEqual(run(full), (0, ''))

    def test_missing_file_exits_2(self):
        r = subprocess.run([sys.executable, SCRIPT, '/no/such/page.html'], capture_output=True)
        self.assertEqual(r.returncode, 2)

if __name__ == '__main__':
    unittest.main()
