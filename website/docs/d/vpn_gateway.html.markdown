---
subcategory: "VPN (Site-to-Site)"
layout: "aws"
page_title: "AWS: aws_vpn_gateway"
description: |-
    Provides details about a specific VPN gateway.
---

# Data Source: aws_vpn_gateway

The VPN Gateway data source provides details about
a specific VPN gateway.

## Example Usage

```terraform
data "aws_vpn_gateway" "selected" {
  filter {
    name   = "tag:Name"
    values = ["vpn-gw"]
  }
}

output "vpn_gateway_id" {
  value = data.aws_vpn_gateway.selected.id
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `id` - (Optional) ID of the specific VPN Gateway to retrieve.
* `state` - (Optional) State of the specific VPN Gateway to retrieve.
* `availability_zone` - (Optional) Availability Zone of the specific VPN Gateway to retrieve.
* `attached_vpc_id` - (Optional) ID of a VPC attached to the specific VPN Gateway to retrieve.
* `filter` - (Optional) Custom filter block as described below.
* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired VPN Gateway.
* `amazon_side_asn` - (Optional) Autonomous System Number (ASN) for the Amazon side of the specific VPN Gateway to retrieve.

The arguments of this data source act as filters for querying the available VPN gateways.
The given filters must match exactly one VPN gateway whose data will be exported as attributes.

### `filter`

More complex filters can be expressed using one or more `filter` sub-blocks, which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeVpnGateways.html).
* `values` - (Required) Set of values that are accepted for the given field.
  A VPN Gateway will be selected if any one of the given values matches.

## Attribute Reference

This data source exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
