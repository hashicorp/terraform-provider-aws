# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

variable "aws_region" {
  description = "The AWS region to create things in."
  default     = "us-west-2"
}

variable "instance_type" {
  description = "Instance type. Must be 8th gen Intel (c8i, m8i, r8i) for nested virtualization."
  default     = "c8i.large"
}
