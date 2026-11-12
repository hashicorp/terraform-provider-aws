# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_devopsagent_asset_file" "test" {
  provider = aws

  include_resource = true

  config {
    agent_space_id = aws_devopsagent_asset.test.agent_space_id
    asset_id       = aws_devopsagent_asset.test.asset_id
  }
}
