# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_version = ">= 0.12"
}

# First account owns the transit gateway and accepts the Network Firewall attachment.
provider "aws" {
  alias = "first"

  region     = var.aws_region
  access_key = var.aws_first_access_key
  secret_key = var.aws_first_secret_key
}

# Second account owns the Network Firewall and creates the VPC attachment.
provider "aws" {
  alias = "second"

  region     = var.aws_region
  access_key = var.aws_second_access_key
  secret_key = var.aws_second_secret_key
}

data "aws_availability_zones" "available" {
  provider = aws.first

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}


data "aws_caller_identity" "second" {
  provider = aws.second
}

resource "aws_ec2_transit_gateway" "example" {
  provider = aws.first

  tags = {
    Name = "terraform-example"
  }
}

resource "aws_ram_resource_share" "example" {
  provider = aws.first

  name = "terraform-example"

  tags = {
    Name = "terraform-example"
  }
}

# Share the transit gateway...
resource "aws_ram_resource_association" "example" {
  provider = aws.first

  resource_arn       = aws_ec2_transit_gateway.example.arn
  resource_share_arn = aws_ram_resource_share.example.id
}

# ...with the second account.
resource "aws_ram_principal_association" "example" {
  provider = aws.first

  principal          = data.aws_caller_identity.second.account_id
  resource_share_arn = aws_ram_resource_share.example.id
}


resource "aws_networkfirewall_firewall_policy" "example" {
  provider = aws.second

  name = "terraform-example"

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}

#Create Network Firewall in the second account attached to the shared transit gateway
resource "aws_networkfirewall_firewall" "example" {
  provider = aws.second

  depends_on = [
    aws_ram_resource_association.example,
    aws_ram_principal_association.example,
  ]

  name                = "terraform-example"
  firewall_policy_arn = aws_networkfirewall_firewall_policy.example.arn
  transit_gateway_id  = aws_ec2_transit_gateway.example.id

  availability_zone_mapping {
    availability_zone_id = data.aws_availability_zones.available.zone_ids[0]
  }

}

# ...and accept it in the first account.
resource "aws_networkfirewall_firewall_transit_gateway_attachment_accepter" "example" {
  provider = aws.first

  transit_gateway_attachment_id = aws_networkfirewall_firewall.example.firewall_status[0].transit_gateway_attachment_sync_states[0].attachment_id
}
