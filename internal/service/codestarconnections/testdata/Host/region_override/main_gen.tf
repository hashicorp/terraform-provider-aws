# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_codestarconnections_host" "test" {
  region = var.region

  name              = var.rName
  provider_endpoint = "https://example.com"
  provider_type     = "GitHubEnterpriseServer"
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
