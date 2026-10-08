# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_bedrockagentcore_gateway_rule" "test" {
  provider = aws

  include_resource = true

  config {
    gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  }
}
