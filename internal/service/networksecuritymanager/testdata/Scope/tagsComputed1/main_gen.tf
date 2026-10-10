# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_networksecuritymanager_scope" "test" {
  name = var.rName

  scope_configuration {
    resource_scope {
      resource_type = "AWS::CloudFront::Distribution"

      include {
        expression {
          criteria {
            tags = {
              nsm-acctest = var.rName
            }
          }
        }
      }
    }
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
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
