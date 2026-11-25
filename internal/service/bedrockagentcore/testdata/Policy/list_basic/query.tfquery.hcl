# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_bedrockagentcore_policy" "test" {
  provider = aws

  config {
    policy_engine_id = aws_bedrockagentcore_policy_engine.test.policy_engine_id
  }
}
