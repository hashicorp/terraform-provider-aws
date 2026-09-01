# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_dashboard" "test" {
  dashboard_name = var.rName

  dashboard_body = <<EOF
{
  "widgets": [
    {
      "type": "text",
      "x": 0,
      "y": 0,
      "width": 6,
      "height": 6,
      "properties": {
        "markdown": "${var.markdown}"
      }
    }
  ]
}
EOF
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "markdown" {
  type     = string
  nullable = false
}