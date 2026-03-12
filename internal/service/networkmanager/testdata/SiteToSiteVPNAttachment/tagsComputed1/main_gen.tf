# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_networkmanager_site_to_site_vpn_attachment" "test" {
  core_network_id    = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpn_connection_arn = aws_vpn_connection.test.arn

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_site_to_site_vpn_attachment.test.id
  attachment_type = aws_networkmanager_site_to_site_vpn_attachment.test.attachment_type
}

# testAccSiteToSiteVPNAttachmentConfig_base

resource "aws_customer_gateway" "test" {
  bgp_asn     = var.rBgpAsn
  ip_address  = var.rIPv4Address
  type        = "ipsec.1"
  device_name = var.rName
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
}

resource "aws_networkmanager_global_network" "test" {}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}

data "aws_region" "current" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges       = ["64512-64555"]
    edge_locations {
      location = data.aws_region.current.region
      asn      = 64512
    }
  }

  segments {
    name                          = "shared"
    description                   = "SegmentForSharedServices"
    require_attachment_acceptance = true
  }

  segment_actions {
    action     = "share"
    mode       = "attachment-route"
    segment    = "shared"
    share_with = ["*"]
  }

  attachment_policies {
    rule_number     = 1
    condition_logic = "or"

    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "segment"
      value    = "shared"
    }

    action {
      association_method = "constant"
      segment            = "shared"
    }
  }
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "rBgpAsn" {
  type     = string
  nullable = false
}

variable "rIPv4Address" {
  type     = string
  nullable = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
