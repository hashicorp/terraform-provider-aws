# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_medialive_input_security_group" "test" {
  whitelist_rules {
    cidr = "10.2.0.0/16"
  }
}

