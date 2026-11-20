# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

"github.com/aws/aws-sdk-go-v2" as $prefix
| ($before[0].Require | INDEX(.Path)) as $old
| [
    $after[0].Require[]
    | select(.Path == $prefix or (.Path | startswith($prefix + "/")))
    | select(.Version != $old[.Path].Version)
    | if .Path == $prefix then .Version
      # Go module tags omit the /vN major-version suffix from module paths.
      else (.Path | ltrimstr($prefix + "/") | sub("/v([2-9]|[1-9][0-9]+)$"; "")) + "/" + .Version
      end
  ]
