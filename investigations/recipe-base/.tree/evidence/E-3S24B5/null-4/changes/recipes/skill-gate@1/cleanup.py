#!/usr/bin/env python3
# Drive kills the runner's process group before it returns, so nothing is
# left to stop. Succeeds whether or not Launch ran, and when run twice.
print("cleanup: nothing to stop")
