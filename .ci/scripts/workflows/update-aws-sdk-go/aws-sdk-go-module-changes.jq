# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

($before[0].Require | INDEX(.Path)) as $old
| ($after[0].Require | INDEX(.Path)) as $new
| "| Module | Old version | New version |",
  "| --- | --- | --- |",
  (($old + $new | keys[]) as $path
    | select($old[$path].Version != $new[$path].Version)
    | "| `\($path)` | `\($old[$path].Version // "-")` | `\($new[$path].Version // "-")` |")
