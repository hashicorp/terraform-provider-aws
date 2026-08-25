# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_kinesis_account_settings" "test" {
  minimum_throughput_billing_commitment {
    status = "DISABLED"
  }
}

