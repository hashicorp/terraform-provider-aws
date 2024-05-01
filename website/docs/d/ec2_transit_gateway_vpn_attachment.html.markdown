---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_vpn_attachment"
description: |-
  Get information on an EC2 Transit Gateway VPN Attachment
---

# Data Source: aws_ec2_transit_gateway_vpn_attachment

Get information on an EC2 Transit Gateway VPN Attachment.

-> EC2 Transit Gateway VPN Attachments are implicitly created by VPN Connections referencing an EC2 Transit Gateway so there is no managed resource. For ease, the [`aws_vpn_connection` resource](/docs/providers/aws/r/vpn_connection.html) includes a `transit_gateway_attachment_id` attribute which can replace some usage of this data source. For tagging the attachment, see the [`aws_ec2_tag` resource](/docs/providers/aws/r/ec2_tag.html).

## Example Usage

### By Transit Gateway and VPN Connection Identifiers

```terraform
data "aws_ec2_transit_gateway_vpn_attachment" "example" {
  transit_gateway_id = aws_ec2_transit_gateway.example.id
  vpn_connection_id  = aws_vpn_connection.example.id
}
```

### Filter

```terraform
data "aws_ec2_transit_gateway_vpn_attachment" "test" {
  filter {
    name   = "resource-id"
    values = ["some-resource"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `transit_gateway_id` - (Optional) Identifier of the EC2 Transit Gateway.
* `vpn_connection_id` - (Optional) Identifier of the EC2 VPN Connection.
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.
* `tags` - (Optional) Map of tags, each pair of which must exactly match a pair on the desired Transit Gateway VPN Attachment.

### filter Configuration Block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter field. Valid values can be found in the [EC2 DescribeTransitGatewayAttachments API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeTransitGatewayAttachments.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - EC2 Transit Gateway VPN Attachment identifier
* `tags` - Key-value tags for the EC2 Transit Gateway VPN Attachment

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
