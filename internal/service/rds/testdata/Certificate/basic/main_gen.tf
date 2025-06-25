# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_rds_certificate" "test" {
  certificate_identifier = "rds-ca-rsa4096-g1"
}

