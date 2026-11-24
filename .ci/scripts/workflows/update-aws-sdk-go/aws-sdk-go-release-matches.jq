# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Compare revisions numerically: .10 follows .2, and an unsuffixed date is revision 0.
def release_key:
  ltrimstr("refs/tags/release-") | split(".") | [.[0], (.[1] // "0" | tonumber)];

. as $tags
| ($tags | INDEX(.ref)) as $index
| ($tags
    | map(select(.ref | test("^refs/tags/release-[0-9]{4}-[0-9]{2}-[0-9]{2}(\\.[0-9]+)?$")))
    | sort_by(.ref | release_key) | reverse) as $releases
| [
    $modules[0][]
    | $index["refs/tags/" + .] as $module
    | if $module == null then null
      else first($releases[] | select(.sha == $module.sha)) // null
      end
  ] as $matches
| {
    unmatched: ($matches | map(select(. == null)) | length),
    releases: ([$matches[] | select(. != null)] | unique_by(.ref)
      | sort_by(.ref | release_key) | reverse | map(.ref | ltrimstr("refs/tags/")))
  }
