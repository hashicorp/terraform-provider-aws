# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_ec2_network_insights_access_scope" "test" {

  match_paths {
    source {
      resource_statement {
        resource_types = ["AWS::EC2::NetworkInterface"]
      }
    }
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "null_resource" "test" {}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
