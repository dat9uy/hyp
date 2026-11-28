# G-YFAERD spec

The change moves no state of hyp's machine, so it holds no model and no chart. It
changes the two scripts of the hyp-rethinking skill and the page-shape reference.

- `check_items.py` refuses a data file with no diagram of section 0 (`picture-missing`),
  with more than six items in section 1 besides one impact item (`event-too-long`), and
  with a bar or a limit that has no label or no number (`bar-number-missing`). A
  diagram's section may be 0 to 10; an item's section stays 1 to 10.
- `draw_page.py` draws the diagrams of section 0 in a block above "1. The event". A
  diagram with `bars` and a `limit` is drawn as bars to scale, each with its number, the
  bars above the limit marked "over", and the limit as a line with its label and number.
- The tests of both scripts are in `skills/hyp-rethinking/scripts/tests/`.
