# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_network_insights_access_scope" "test" {

  match_paths {
    source {
      resource_statement {
        resource_types = ["AWS::EC2::NetworkInterface"]
      }
    }
  }

  tags = var.resource_tags
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
