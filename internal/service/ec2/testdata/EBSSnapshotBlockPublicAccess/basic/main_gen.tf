# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_ebs_snapshot_block_public_access" "test" {
  state = "block-all-sharing"
}

