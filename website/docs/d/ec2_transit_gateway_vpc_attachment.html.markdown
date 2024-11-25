---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_vpc_attachment"
description: |-
  Get information on an EC2 Transit Gateway VPC Attachment
---

# Data Source: aws_ec2_transit_gateway_vpc_attachment

Get information on an EC2 Transit Gateway VPC Attachment.

## Example Usage

### By Filter

```terraform
data "aws_ec2_transit_gateway_vpc_attachment" "example" {
  filter {
    name   = "vpc-id"
    values = ["vpc-12345678"]
  }
}
```

### By Identifier

```terraform
data "aws_ec2_transit_gateway_vpc_attachment" "example" {
  id = "tgw-attach-12345678"
}
```

## Argument Reference

This data source supports the following arguments:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.
* `id` - (Optional) Identifier of the EC2 Transit Gateway VPC Attachment.

### filter Argument Reference

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `appliance_mode_support` - Whether Appliance Mode support is enabled.
* `dns_support` - Whether DNS support is enabled.
* `security_group_referencing_support` - Whether Security Group Referencing Support is enabled.
* `id` - EC2 Transit Gateway VPC Attachment identifier
* `ipv6_support` - Whether IPv6 support is enabled.
* `subnet_ids` - Identifiers of EC2 Subnets.
* `transit_gateway_id` - EC2 Transit Gateway identifier
* `tags` - Key-value tags for the EC2 Transit Gateway VPC Attachment
* `vpc_id` - Identifier of EC2 VPC.
* `vpc_owner_id` - Identifier of the AWS account that owns the EC2 VPC.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
