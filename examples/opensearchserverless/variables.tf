# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

variable "aws_region" {
  description = "The AWS region to create things in."
  default     = "us-east-1"
}

variable "collection_name" {
  description = "Name of the OpenSearch Serverless collection."
  default     = "example-collection"
}
