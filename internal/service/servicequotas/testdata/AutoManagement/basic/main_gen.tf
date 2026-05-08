# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_servicequotas_auto_management" "test" {

  opt_in_level = "ACCOUNT"
  opt_in_type  = "NotifyOnly"
}
