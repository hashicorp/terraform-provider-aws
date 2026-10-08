# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_network_insights_access_scope" "test" {
  region = var.region


  match_paths {
    source {
      resource_statement {
        resource_types = ["AWS::EC2::NetworkInterface"]
      }
    }
  }
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
