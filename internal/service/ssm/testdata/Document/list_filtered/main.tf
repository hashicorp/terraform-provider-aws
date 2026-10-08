# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssm_document" "expected" {
  count = 2

  document_type = "Command"
  name          = "${var.rName}-expected-${count.index}"

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

resource "aws_ssm_document" "not_expected" {
  count = 2

  document_type = "Command"
  name          = "${var.rName}-other-${count.index}"

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

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
