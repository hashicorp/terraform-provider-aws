# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_devopsagent_asset_file" "test" {
  count  = 2
  region = var.region

  agent_space_id = aws_devopsagent_asset.test.agent_space_id
  asset_id       = aws_devopsagent_asset.test.asset_id
  path           = "references/runbook-${count.index}.md"

  content {
    text = "Check the runbook first."
  }
}

resource "aws_devopsagent_asset" "test" {
  region = var.region

  agent_space_id = aws_devopsagent_agent_space.test.agent_space_id
  asset_type     = "skill"
  metadata = jsonencode({
    name        = var.rName
    description = "Terraform acceptance test skill."
    agent_types = ["GENERIC"]
  })

  content {
    file {
      path = "SKILL.md"
      body {
        text = "Follow the operational runbook."
      }
    }
  }
}

resource "aws_devopsagent_agent_space" "test" {
  region = var.region

  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
