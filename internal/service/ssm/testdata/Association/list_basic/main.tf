# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssm_association" "test" {
  count            = var.resource_count
  name             = aws_ssm_document.test[count.index].name
  association_name = "${var.rName}-${count.index}"

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}

resource "aws_ssm_document" "test" {
  count         = var.resource_count
  name          = "${var.rName}-${count.index}"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Test doc",
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

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
