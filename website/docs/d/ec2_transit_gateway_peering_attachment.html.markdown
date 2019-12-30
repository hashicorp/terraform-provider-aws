---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_peering_attachment"
description: |-
  Get information on an EC2 Transit Gateway Peering Attachment
---

# Data Source: aws_ec2_transit_gateway_peering_attachment

Get information on an EC2 Transit Gateway Peering Attachment.

## Example Usage

### By Filter

```hcl
data "aws_ec2_transit_gateway_peering_attachment" "example" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = ["tgw-attach-12345678"]
  }
}
```

### By Identifier

```hcl
 data "aws_ec2_transit_gateway_peering_attachment" "attachment" {
   id = "tgw-attach-12345678"
 }

```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.
* `id` - (Optional) Identifier of the EC2 Transit Gateway Peering Attachment.

### filter Argument Reference

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:


* `id` - EC2 Transit Gateway Peering Attachment identifier
* `peer_account_id` - Identifier of the AWS account that owns the accepter Transit Gateway.
* `peer_region` - Identifier of the AWS region that owns the accepter Transit Gateway.
* `transit_gateway_id` - Requester EC2 Transit Gateway identifier
* `peer_transit_gateway_id` - Accepter EC2 Transit Gateway identifier
* `tags` - Key-value tags for the EC2 Transit Gateway VPC Attachment