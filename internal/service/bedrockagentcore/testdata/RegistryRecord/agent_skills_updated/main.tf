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
description: Make a cup of tea.
---
# My Skill

1. Fill kettle.
2. Boil water.
3. Put teabag in cup.
4. Pour water into cup.
5. Enjoy your tea!
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
