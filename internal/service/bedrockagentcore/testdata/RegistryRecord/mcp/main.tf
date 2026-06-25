# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrockagentcore_registry_record" "test" {
  name        = "${var.rName}-record"
  registry_id = aws_bedrockagentcore_registry.test.registry_id

  descriptor_type = "MCP"

  descriptors {
    mcp {
      server {
        inline_content = <<EOF
{
  "name": "io.example/my-server",
  "description": "Brief description of server functionality",
  "version": "1.0.0"
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
