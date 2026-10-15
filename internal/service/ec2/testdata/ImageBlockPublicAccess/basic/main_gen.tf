# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_image_block_public_access" "test" {
  state = "block-new-sharing"
}

