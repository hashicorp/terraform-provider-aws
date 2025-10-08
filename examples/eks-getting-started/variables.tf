# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

variable "aws_region" {
  default = "us-west-2"
}

variable "cluster_name" {
  default = "terraform-eks-demo"
  type    = string
}
