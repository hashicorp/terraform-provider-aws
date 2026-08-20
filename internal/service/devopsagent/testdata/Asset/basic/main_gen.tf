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

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
