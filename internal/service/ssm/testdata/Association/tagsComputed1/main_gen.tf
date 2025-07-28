# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_ssm_association" "test" {
  name                = aws_ssm_document.test.name
  schedule_expression = "cron(0 16 ? * WED *)"

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_ssm_document" "test" {
  name          = var.rName
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
