---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_connect"
description: |-
  Manages a Connect attachment from a specified transit gateway attachment
---

# Resource: aws_ec2_transit_gateway_connect

Manages a Connect attachment from a specified transit gateway attachment.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway_connect" "example" {
  transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.example.id
}
```

## Argument Reference

The following arguments are supported:

* `transport_transit_gateway_attachment_id` - (Required) The ID of the transit gateway attachment. You can specify a VPC attachment or Amazon Web Services Direct Connect attachment.
* `protocol` - (Optional) The tunnel protocol. Valid and Default value: `gre`.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Connect Attachment. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `transit_gateway_default_route_table_association` - (Optional) Boolean whether the Connect Attachment should be associated with the EC2 Transit Gateway association default route table. This cannot be configured or perform drift detection with Resource Access Manager shared EC2 Transit Gateways. Default value: `true`.
* `transit_gateway_default_route_table_propagation` - (Optional) Boolean whether the Connect Attachment should propagate routes with the EC2 Transit Gateway propagation default route table. This cannot be configured or perform drift detection with Resource Access Manager shared EC2 Transit Gateways. Default value: `true`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway Attachment identifier
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).
* `transit_gateway_id` - Identifier of EC2 Transit Gateway.

## Import

`aws_ec2_transit_gateway_connect` can be imported by using the EC2 Transit Gateway Attachment identifier, e.g.

```
$ terraform import aws_ec2_transit_gateway_connect.example tgw-attach-12345678
```
