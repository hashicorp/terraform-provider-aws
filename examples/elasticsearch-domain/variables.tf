# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

variable "aws_region" {
  default = "us-west-2"
}

variable "domain" {
  description = "The name of the Elasticsearch Domain"
  default     = "elasticsearch-domain-test"
}
