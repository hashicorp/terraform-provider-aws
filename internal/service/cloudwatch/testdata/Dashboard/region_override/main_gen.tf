# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_dashboard" "test" {
  region = var.region

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
        "markdown": "Hi there from Terraform: CloudWatch"
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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
