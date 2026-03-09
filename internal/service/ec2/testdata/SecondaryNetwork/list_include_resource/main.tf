# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_secondary_network" "test" {
  count = var.resource_count

  ipv4_cidr_block = "10.0.${count.index}.0/24"
  network_type    = "rdma"
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
