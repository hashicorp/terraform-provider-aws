# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

split("\n")
| map(select(length > 0) | split("\t") | {sha: .[0], ref: .[1]})
# Annotated tags have separate object IDs; prefer their peeled ^{} commit IDs.
| group_by(.ref | rtrimstr("^{}"))
| map(sort_by(.ref | endswith("^{}")) | last | .ref |= rtrimstr("^{}"))
