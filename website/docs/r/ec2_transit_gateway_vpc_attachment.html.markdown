---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_vpc_attachment"
description: |-
  Manages an EC2 Transit Gateway VPC Attachment
---

# Resource: aws_ec2_transit_gateway_vpc_attachment

Manages an EC2 Transit Gateway VPC Attachment. For examples of custom route table association and propagation, see the EC2 Transit Gateway Networking Examples Guide.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway_vpc_attachment" "example" {
  subnet_ids         = [aws_subnet.example.id]
  transit_gateway_id = aws_ec2_transit_gateway.example.id
  vpc_id             = aws_vpc.example.id
}
```

A full example of how to create a Transit Gateway in one AWS account, share it with a second AWS account, and attach a VPC in the second account to the Transit Gateway via the `aws_ec2_transit_gateway_vpc_attachment` and `aws_ec2_transit_gateway_vpc_attachment_accepter` resources can be found in [the `./examples/transit-gateway-cross-account-vpc-attachment` directory within the Github Repository](https://github.com/hashicorp/terraform-provider-aws/tree/main/examples/transit-gateway-cross-account-vpc-attachment).

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `subnet_ids` - (Required) Identifiers of EC2 Subnets.
* `transit_gateway_id` - (Required) Identifier of EC2 Transit Gateway.
* `vpc_id` - (Required) Identifier of EC2 VPC.
* `appliance_mode_support` - (Optional) Whether Appliance Mode support is enabled. If enabled, a traffic flow between a source and destination uses the same Availability Zone for the VPC attachment for the lifetime of that flow. Valid values: `disable`, `enable`. Default value: `disable`.
* `dns_support` - (Optional) Whether DNS support is enabled. Valid values: `disable`, `enable`. Default value: `enable`.
* `ipv6_support` - (Optional) Whether IPv6 support is enabled. Valid values: `disable`, `enable`. Default value: `disable`.
* `security_group_referencing_support` - (Optional) Whether Security Group Referencing Support is enabled. Valid values: `disable`, `enable`.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway VPC Attachment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `transit_gateway_default_route_table_association` - (Optional) Boolean whether the VPC Attachment should be associated with the EC2 Transit Gateway association default route table. This cannot be configured or perform drift detection with Resource Access Manager shared EC2 Transit Gateways. Default value: `true`. **Setting this argument to** `false` will cause Terraform to remove the default association if it exists. This argument does not simply “ignore” the association; it actively manages (adds or removes) the association in AWS to match the value in your configuration. Do not attempt to manage the same association with both this resource and `aws_ec2_transit_gateway_route_table_association`, as this will cause perpetual diffs and resource churn. Use conditional logic to ensure only one resource manages the association at a time.
* `transit_gateway_default_route_table_propagation` - (Optional) Boolean whether the VPC Attachment should propagate routes with the EC2 Transit Gateway propagation default route table. This cannot be configured or perform drift detection with Resource Access Manager shared EC2 Transit Gateways. Default value: `true`. **Setting this argument to** `false` will cause Terraform to remove the default propagation if it exists. This argument does not simply “ignore” the propagation; it actively manages (adds or removes) the propagation in AWS to match the value in your configuration. Do not attempt to manage the same propagation with both this resource and `aws_ec2_transit_gateway_route_table_propagation`, as this will cause perpetual diffs and resource churn. Use conditional logic to ensure only one resource manages the propagation at a time.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the attachment.
* `id` - EC2 Transit Gateway Attachment identifier.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `vpc_owner_id` - Identifier of the AWS account that owns the EC2 VPC.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_transit_gateway_vpc_attachment` using the EC2 Transit Gateway Attachment identifier. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_vpc_attachment.example
  id = "tgw-attach-12345678"
}
```

Using `terraform import`, import `aws_ec2_transit_gateway_vpc_attachment` using the EC2 Transit Gateway Attachment identifier. For example:

```console
% terraform import aws_ec2_transit_gateway_vpc_attachment.example tgw-attach-12345678
```
