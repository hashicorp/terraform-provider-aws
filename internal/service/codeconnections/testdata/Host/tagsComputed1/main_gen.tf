# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_codeconnections_host" "test" {
  name              = var.rName
  provider_endpoint = "https://example.com"
  provider_type     = "GitHubEnterpriseServer"

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
