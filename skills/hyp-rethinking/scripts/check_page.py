#!/usr/bin/env python3
"""Check a rethink page for the defects a long design talk leaves behind.

Usage: check_page.py PAGE.html [--max-words 25]
Reports headings out of order, "section N" references to a missing section,
duplicate id values, descriptive sentences longer than the limit, and a section 6
with fewer than three diagrams or a section 7 with none. A diagram is an <svg> or a
Mermaid block. A quotation in double quotes counts as one word in the sentence
length: the page keeps the owner's words exactly, so their length is not the
page's to fix.
Exit 0 when clean, 1 with findings, 2 on bad arguments.
"""
import argparse, html, re, sys

def text_of(fragment):
    fragment = re.sub(r'<(svg|script|style|code|pre)\b.*?</\1>', ' ', fragment, flags=re.S)
    fragment = re.sub(r'<span class="tag[^"]*">.*?</span>', ' ', fragment, flags=re.S)
    fragment = re.sub(r'</(td|th|li|p|h\d|div|figcaption)>', '. ', fragment)
    return html.unescape(re.sub(r'\s+', ' ', re.sub(r'<[^>]+>', ' ', fragment)))

QUOTE = re.compile(r'"[^"<>]{0,800}?"|\u201c[^\u201d<>]{0,800}?\u201d')

def section(body, n):
    m = re.search(r'<h2[^>]*>\s*%d\.' % n, body)
    if not m:
        return None
    nxt = re.search(r'<h2[^>]*>', body[m.end():])
    return body[m.end():m.end() + nxt.start()] if nxt else body[m.end():]

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('page')
    ap.add_argument('--max-words', type=int, default=25)
    a = ap.parse_args()
    try:
        page = open(a.page, encoding='utf-8').read()
    except OSError as e:
        print(f'cannot read {a.page}: {e}', file=sys.stderr); return 2
    body = page[page.find('<body'):] if '<body' in page else page
    out = []
    nums = [int(n) for n in re.findall(r'<h2[^>]*>\s*(\d+)\.', body)]
    for prev, cur in zip(nums, nums[1:]):
        if cur != prev + 1:
            out.append(f'heading order: section {cur} follows section {prev}')
    have = set(nums)
    for n in sorted({int(n) for n in re.findall(r'\bsections? (\d+)\b', text_of(body))}):
        if have and n not in have:
            out.append(f'reference to missing section {n}')
    ids = re.findall(r'\sid="([^"]+)"', body)
    for i in sorted({i for i in ids if ids.count(i) > 1}):
        out.append(f'duplicate id "{i}" ({ids.count(i)} times)')
    for n, need in ((6, 3), (7, 1)):
        part = section(body, n)
        if part is not None:
            got = len(re.findall(r'<svg\b|class="mermaid"|```mermaid', part))
            if got < need:
                out.append(f'section {n}: {got} diagrams, needs {need}')
    for s in re.split(r'(?<=[.!?])\s+', QUOTE.sub('"Q"', text_of(body))):
        w = [t for t in s.split() if re.search(r'\w', t)]
        if len(w) > a.max_words:
            out.append(f'{len(w)} words: {s[:110]}')
    for line in out:
        print(line)
    return 1 if out else 0

if __name__ == '__main__':
    sys.exit(main())
