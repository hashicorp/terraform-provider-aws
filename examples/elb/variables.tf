# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

variable "key_name" {
  description = "Name of the SSH keypair to use in AWS."
}

variable "aws_region" {
  description = "AWS region to launch servers."
  default     = "us-east-1"
}


# ubuntu-jammy-22.04 (amd64)
variable "aws_amis" {
  default = {
    "us-east-1" = "ami-005fc0f236362e99f"
    "us-west-2" = "ami-0075013580f6322a1"
  }
}
