# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

variable "public_key_path" {
  description = <<DESCRIPTION
Path to the SSH public key to be used for authentication.
Ensure this keypair is added to your local SSH agent so provisioners can
connect.

Example: ~/.ssh/terraform.pub
DESCRIPTION
}

variable "key_name" {
  description = "Desired name of AWS key pair"
}

variable "aws_region" {
  description = "AWS region to launch servers."
  default     = "us-west-2"
}

# Ubuntu Bionic 18.04 LTS (x64)
variable "aws_amis" {
  default = {
    eu-west-1 = "ami-0c259a97cbf621daf"
    us-east-1 = "ami-04751c628226b9b59"
    us-west-1 = "ami-0558dde970ca91ee5"
    us-west-2 = "ami-0bdef2eb518663879"
  }
}
