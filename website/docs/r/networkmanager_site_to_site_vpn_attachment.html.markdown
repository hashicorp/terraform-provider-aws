---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_site_to_site_vpn_attachment"
description: |-
  Manages a Network Manager site-to-site VPN attachment.
---

# Resource: aws_networkmanager_site_to_site_vpn_attachment

Manages a Network Manager site-to-site VPN attachment.

## Example Usage

### Basic Usage

```terraform
resource "aws_networkmanager_site_to_site_vpn_attachment" "example" {
  core_network_id    = awscc_networkmanager_core_network.example.id
  vpn_connection_arn = aws_vpn_connection.example.arn
}
```

### Full Usage

```terraform
resource "aws_customer_gateway" "test" {
  bgp_asn    = 65000
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}
resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  tags = {
    Name = "test"
  }
}
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = "test"
  }
}
resource "awscc_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  policy_document   = jsonencode(jsondecode(data.aws_networkmanager_core_network_policy_document.test.json))
}
data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges       = ["64512-64555"]
    edge_locations {
      location = data.aws_region.current.name
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

resource "aws_networkmanager_site_to_site_vpn_attachment" "test" {
  core_network_id    = awscc_networkmanager_core_network.test.id
  vpn_connection_arn = aws_vpn_connection.test.arn
  tags = {
    segment = "shared"
  }
}
resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_site_to_site_vpn_attachment.test.id
  attachment_type = aws_networkmanager_site_to_site_vpn_attachment.test.attachment_type
}
```

## Argument Reference

The following arguments are required:

* `core_network_id` - (Required) ID of a core network for the VPN attachment.
* `vpn_connection_arn` - (Required) ARN of the site-to-site VPN connection.

The following arguments are optional:

* `tags` - (Optional) Key-value tags for the attachment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the attachment.
* `attachment_policy_rule_number` - Policy rule number associated with the attachment.
* `attachment_type` - Type of attachment.
* `core_network_arn` - ARN of a core network.
* `edge_location` - Region where the edge is located.
* `id` - ID of the attachment.
* `owner_account_id` - ID of the attachment account owner.
* `resource_arn` - Attachment resource ARN.
* `segment_name` - Name of the segment attachment.
* `state` - State of the attachment.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)
* `update` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_site_to_site_vpn_attachment` using the attachment ID. For example:

```terraform
import {
  to = aws_networkmanager_site_to_site_vpn_attachment.example
  id = "attachment-0f8fa60d2238d1bd8"
}
```

Using `terraform import`, import `aws_networkmanager_site_to_site_vpn_attachment` using the attachment ID. For example:

```console
% terraform import aws_networkmanager_site_to_site_vpn_attachment.example attachment-0f8fa60d2238d1bd8
```
