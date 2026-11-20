# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_devopsagent_asset" "test" {
  count = 2

  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "agents_md"
  metadata       = jsonencode({ agent_type = ["GENERIC", "CHAT"][count.index] })

  content {
    file {
      path = "AGENTS.md"
      body {
        text = "Follow the operational runbook."
      }
    }
  }
}

resource "aws_devopsagent_agent_space" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
