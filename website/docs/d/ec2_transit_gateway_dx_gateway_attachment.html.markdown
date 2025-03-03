---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_dx_gateway_attachment"
description: |-
  Get information on an EC2 Transit Gateway's attachment to a Direct Connect Gateway
---

# Data Source: aws_ec2_transit_gateway_dx_gateway_attachment

Get information on an EC2 Transit Gateway's attachment to a Direct Connect Gateway.

## Example Usage

### By Transit Gateway and Direct Connect Gateway Identifiers

```terraform
data "aws_ec2_transit_gateway_dx_gateway_attachment" "example" {
  transit_gateway_id = aws_ec2_transit_gateway.example.id
  dx_gateway_id      = aws_dx_gateway.example.id
}
```

## Argument Reference

This data source supports the following arguments:

* `transit_gateway_id` - (Optional) Identifier of the EC2 Transit Gateway.
* `dx_gateway_id` - (Optional) Identifier of the Direct Connect Gateway.
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.
* `tags` - (Optional) Map of tags, each pair of which must exactly match a pair on the desired Transit Gateway Direct Connect Gateway Attachment.

### filter Configuration Block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter field. Valid values can be found in the [EC2 DescribeTransitGatewayAttachments API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeTransitGatewayAttachments.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the attachment.
* `id` - EC2 Transit Gateway Attachment identifier,
* `tags` - Key-value tags for the EC2 Transit Gateway Attachment.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
