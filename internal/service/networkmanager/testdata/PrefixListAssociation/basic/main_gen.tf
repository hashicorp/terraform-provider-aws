# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networkmanager_prefix_list_association" "test" {
  core_network_id   = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  prefix_list_arn   = aws_ec2_managed_prefix_list.test.arn
  prefix_list_alias = "testprefixlist"
}

resource "aws_ec2_managed_prefix_list" "test" {
  name           = var.rName
  address_family = "IPv4"
  max_entries    = 5

  entry {
    cidr        = "10.0.0.0/8"
    description = "Test CIDR"
  }
}

resource "aws_networkmanager_global_network" "test" {}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = data.aws_region.current.region
    }
  }

  segments {
    name                          = "segment"
    require_attachment_acceptance = false
  }
}

data "aws_region" "current" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
