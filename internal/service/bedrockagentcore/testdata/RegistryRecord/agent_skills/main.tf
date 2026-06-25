# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_registry_record" "test" {
  name        = "${var.rName}-record"
  registry_id = aws_bedrockagentcore_registry.test.registry_id

  descriptor_type = "AGENT_SKILLS"

  descriptors {
    agent_skills {
      skill_md {
        inline_content = <<EOF
---
name: my-skill
description: Brief description of what this skill does.
---
# My Skill

Describe your skill's purpose, usage, and capabilities here.
EOF
      }
    }
  }
}

resource "aws_bedrockagentcore_registry" "test" {
  name = "${var.rName}-registry"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
