---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_vpc_attachments"
description: |-
  Get information on EC2 Transit Gateway VPC Attachments
---

# Data Source: aws_ec2_transit_gateway_vpc_attachments

Get information on EC2 Transit Gateway VPC Attachments.

## Example Usage

### By Filter

```hcl
data "aws_ec2_transit_gateway_vpc_attachments" "example" {
  filter {
    name   = "state"
    values = ["pendingAcceptance"]
  }
}
```
## Argument Reference

The following arguments are supported:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.

### filter Argument Reference

* `name` - (Required) Name of the filter @see https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-transit-gateway-attachments.html.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `transit_gateway_vpc_attachments` A list of all attachments matching filter

### transit_gateway_vpc_attachments Attribute Reference

Each member of this list have following attributes:

* `dns_support` - Whether DNS support is enabled.
* `id` - EC2 Transit Gateway VPC Attachment identifier
* `ipv6_support` - Whether IPv6 support is enabled.
* `subnet_ids` - Identifiers of EC2 Subnets.
* `transit_gateway_id` - EC2 Transit Gateway identifier
* `tags` - Key-value tags for the EC2 Transit Gateway VPC Attachment
* `vpc_id` - Identifier of EC2 VPC.
* `vpc_owner_id` - Identifier of the AWS account that owns the EC2 VPC.
