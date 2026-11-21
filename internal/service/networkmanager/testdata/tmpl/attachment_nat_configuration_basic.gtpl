resource "aws_networkmanager_attachment_nat_configuration" "test" {
  attachment_id       = aws_networkmanager_vpc_attachment.test.id
  core_network_id     = aws_networkmanager_vpc_attachment.test.core_network_id
  nat_gateway_support = "ENABLE"

  nat_gateway_rule {
    rule_number       = 100
    match_source_cidr = "10.0.0.0/16"
    nat_cidr          = "100.64.0.0/16"
  }
}

resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = aws_subnet.test[*].arn
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpc_arn         = aws_vpc.test.arn
}

# testAccVPCAttachmentConfig_base

resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = var.rName
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = var.rName
  }
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
    require_attachment_acceptance = false
  }

  segment_actions {
    action     = "share"
    mode       = "attachment-route"
    segment    = "shared"
    share_with = ["*"]
  }

  attachment_policies {
    rule_number = 1

    conditions {
      type = "any"
    }

    action {
      association_method = "constant"
      segment            = "shared"
    }
  }
}

{{ template "acctest.ConfigVPCWithSubnetsIPv6" 2 }}
