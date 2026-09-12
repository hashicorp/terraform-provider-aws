# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_taxsettings_tax_inheritance" "test" {
  heritage_status = "OptOut"
}
