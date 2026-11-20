# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_devopsagent_asset" "test" {
  provider = aws

  config {
    agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  }
}
