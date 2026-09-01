# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_secondary_network" "test" {
  region = var.region

  ipv4_cidr_block = "10.0.0.0/16"
  network_type    = "rdma"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
