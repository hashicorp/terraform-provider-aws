# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_network_insights_access_scope" "test" {
  count = var.resource_count

  match_paths {
    source {
      resource_statement {
        resource_types = ["AWS::EC2::NetworkInterface"]
      }
    }
  }

  tags = var.resource_tags
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource"
  type        = map(string)
  nullable    = false
}
