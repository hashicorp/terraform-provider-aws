# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

variable "cidr_blocks" {
  default     = "0.0.0.0/0"
  description = "CIDR for sg"
}

variable "sg_name" {
  default     = "rds_sg"
  description = "Tag Name for sg"
}
