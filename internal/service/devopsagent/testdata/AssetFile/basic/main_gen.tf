# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_devopsagent_agent_space" "test" {
  name = "tf-acc-test-devopsagent"
}

resource "aws_devopsagent_asset" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "skill"
  content_path   = "SKILL.md"
  content_body   = "# Test Skill\n\nThis is a test skill."

  metadata = jsonencode({
    name        = var.rName
    description = "A test skill"
    agent_types = ["GENERIC"]
  })
}

resource "aws_devopsagent_asset_file" "test" {
  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_id       = aws_devopsagent_asset.test.asset_id
  path           = "README.md"
  content_body   = "# Hello\n\nThis is a test file."
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
