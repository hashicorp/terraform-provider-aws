# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

"github.com/aws/aws-sdk-go-v2" as $prefix
| [
    .Require[]
    | select(.Path == $prefix or (.Path | startswith($prefix + "/")))
    | if .Path == $prefix then .Version
      else (.Path | ltrimstr($prefix + "/") | sub("/v([2-9]|[1-9][0-9]+)$"; "")) + "/" + .Version
      end
  ]
