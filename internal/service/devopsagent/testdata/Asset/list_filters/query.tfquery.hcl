# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_devopsagent_asset" "included" {
  provider = aws

  config {
    agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
    asset_type     = "agents_md"
    updated_after  = timeadd(aws_devopsagent_asset.test.updated_at, "-1h")
    updated_before = timeadd(aws_devopsagent_asset.test.updated_at, "1h")
  }
}

list "aws_devopsagent_asset" "excluded_type" {
  provider = aws

  config {
    agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
    asset_type     = "skill"
  }
}

list "aws_devopsagent_asset" "excluded_after" {
  provider = aws

  config {
    agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
    updated_after  = timeadd(aws_devopsagent_asset.test.updated_at, "1h")
  }
}

list "aws_devopsagent_asset" "excluded_before" {
  provider = aws

  config {
    agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
    updated_before = timeadd(aws_devopsagent_asset.test.updated_at, "-1h")
  }
}
