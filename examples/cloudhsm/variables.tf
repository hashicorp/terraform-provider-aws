# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

variable "aws_region" {
  description = "AWS region to launch cloudHSM cluster."
  default     = "eu-west-1"
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}
