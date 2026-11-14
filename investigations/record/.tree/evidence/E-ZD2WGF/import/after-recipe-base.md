# G-93VFSB A first author with only hyp and its skills builds a working runner-gate instrument with fewer stops given a base — complete

## Q-WXZRA3 Does an embedded base help a first author build a runner instrument, and if so, how is the base split? — answered

Selected: H-HYXCF9
Design: Read over the held-out recipe as the mean over trials per arm: neither base arm has pass count >= null's and mean stops <= null's mean stops - 1. Today's path (recipe-template.md + RECIPE-FORMAT.md) is the arm; the null hyp binary embeds no base.

Limits:

- R-H5H95P#1 accepted by owner-dat — Both base-arm end-state fails (coarse-3, fine-5) and their f_weak_cases stops come from one base defect: score-usage's empty input/, which remove and git drop (empty-input.txt). Fixed, coarse is 5/5 at 0.80 and fine 5/5 at 0.40: gaps 0.40 and 0.80, still under the 1-stop bar.
- R-H5H95P#2 accepted by owner-dat — The rule's bar counts stops, not writing: bases cut edited lines from 1068 to about 230 per trial, but null runners also passed 5/5 at 1.2 stops, so the measured benefit is less writing, which the rule does not reward.
- R-H5H95P#3 accepted by owner-dat — One method (the skill-text gate), five sonnet trials per arm, candidates derived from three recipes this repository wrote; two null runners made and discarded nested smoke sandboxes in the operator's ~/.hyp-sandbox.

- H-HYXCF9 Template and RECIPE-FORMAT are enough: a base gives no fewer stops, no better end state on held-out skill-text-arms — supported
  - E-3S24B5 Sonnet runners build the held-out skill-text gate instrument on null, coarse and fine arms; stops, end state, edits — run (recipe recipe-base-arms@1, run and rule)
    - R-H5H95P recorded H-HYXCF9=supports H-BFRAPB=contradicts H-Z3PBEA=contradicts: verdict (recipe-base-arms@1, 5 trials/arm after the pre-registered extension at 3): null mean stops 1.20, pass 5/5, edited 1068; coarse 1.00, 4/5, 217; fine 0.60, 4/5, 249. No base arm beats null (pass not worse and stops at least 1 lower), so h-null supports. Every base runner (10/10) started from its bases and wrote about a fifth of null's lines. Runners: sonnet, USD 11.88 total.
- H-BFRAPB One base:claude-runner-gate@1 (Launch, Doctor, Drive, Cleanup, score, verdict) gives the fewest stops and edits — contradicted
- H-Z3PBEA Bases claude-runner, transcript-score, gate-verdict and fixture-check give the fewest stops and edits — contradicted
