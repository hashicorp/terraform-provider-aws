# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

variable "aws_region" {
  description = "The AWS region to create things in."
  default     = "us-east-1"
}

# ubuntu-jammy-22.04 (amd64)
variable "aws_amis" {
  default = {
    "us-east-1" = "ami-005fc0f236362e99f"
    "us-west-2" = "ami-0075013580f6322a1"
  }
}

variable "key_name" {
  description = "Name of the SSH keypair to use in AWS."
}
