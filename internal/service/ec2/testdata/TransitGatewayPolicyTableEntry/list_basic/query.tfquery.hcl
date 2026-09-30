# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_ec2_transit_gateway_policy_table_entry" "test" {
  provider = aws

  config {
    transit_gateway_policy_table_id = aws_ec2_transit_gateway_policy_table.test.id
  }
}
