#!/bin/sh
echo '{}' >> .tree/events/stub.jsonl
echo "warn: stub $1" >&2
