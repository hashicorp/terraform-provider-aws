# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_xray_encryption_config" "test" {
  type = "NONE"
}

