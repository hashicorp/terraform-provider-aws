# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_registry_record" "test" {
  name        = "${var.rName}-record"
  registry_id = aws_bedrockagentcore_registry.test.registry_id

  descriptor_type = "A2A"

  descriptors {
    a2a {
      agent_card {
        inline_content = <<EOF
{
  "name": "My Agent",
  "description": "Brief description of what this agent does",
  "url": "https://api.example.com/a2a",
  "version": "1.0.0",
  "protocolVersion": "0.3",
  "capabilities": {},
  "defaultInputModes": [
    "text/plain"
  ],
  "defaultOutputModes": [
    "text/plain"
  ],
  "skills": [
    {
      "id": "default-skill",
      "name": "Default Skill",
      "description": "Description of what this skill does",
      "tags": [
        "general"
      ]
    }
  ]
}
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
