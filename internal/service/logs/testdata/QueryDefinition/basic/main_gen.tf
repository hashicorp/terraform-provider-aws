# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_query_definition" "test" {
  name = var.rName

  query_string = <<EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 20
EOF
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
