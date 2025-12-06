# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_ebs_snapshot_block_public_access" "test" {
  state = "block-all-sharing"
}

