---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_connect"
description: |-
  Get information on an EC2 Transit Gateway Connect
---

# Data Source: aws_ec2_transit_gateway_connect

Get information on an EC2 Transit Gateway Connect.

## Example Usage

### By Filter

```terraform
data "aws_ec2_transit_gateway_connect" "example" {
  filter {
    name   = "transport-transit-gateway-attachment-id"
    values = ["tgw-attach-12345678"]
  }
}
```

### By Identifier

```terraform
data "aws_ec2_transit_gateway_connect" "example" {
  transit_gateway_connect_id = "tgw-attach-12345678"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.
* `transit_gateway_connect_id` - (Optional) Identifier of the EC2 Transit Gateway Connect.

### filter Argument Reference

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `protocol` - Tunnel protocol
* `tags` - Key-value tags for the EC2 Transit Gateway Connect
* `transit_gateway_id` - EC2 Transit Gateway identifier
* `transport_attachment_id` - The underlaying VPC attachment

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
