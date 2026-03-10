---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_metering_policy"
description: |-
  Manages an EC2 Transit Gateway Metering Policy
---

# Resource: aws_ec2_transit_gateway_metering_policy

Manages an EC2 Transit Gateway Metering Policy for Flexible Cost Allocation (FCA). A metering policy defines how traffic is metered for cost allocation purposes on a Transit Gateway.

## Example Usage

### Basic Usage

```terraform
resource "aws_ec2_transit_gateway" "example" {
  tags = {
    Name = "example"
  }
}

resource "aws_ec2_transit_gateway_metering_policy" "example" {
  transit_gateway_id = aws_ec2_transit_gateway.example.id

  tags = {
    Name = "example"
  }
}
```

### With Middlebox Attachments

```terraform
resource "aws_ec2_transit_gateway_metering_policy" "example" {
  transit_gateway_id       = aws_ec2_transit_gateway.example.id
  middlebox_attachment_ids = [aws_ec2_transit_gateway_vpc_attachment.example.id]

  tags = {
    Name = "example"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `transit_gateway_id` - (Required, Forces new resource) EC2 Transit Gateway identifier.
* `middlebox_attachment_ids` - (Optional) Set of Transit Gateway attachment IDs to designate as middlebox attachments for this metering policy.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Metering Policy. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - EC2 Transit Gateway Metering Policy ARN.
* `transit_gateway_metering_policy_id` - EC2 Transit Gateway Metering Policy identifier.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_metering_policy.example
  identity = {
    id = "tgw-mp-12345678"
  }
}

resource "aws_ec2_transit_gateway_metering_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) ID of the EC2 Transit Gateway Metering Policy.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Transit Gateway Metering Policies using the `transit_gateway_metering_policy_id`. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_metering_policy.example
  id = "tgw-mp-12345678"
}
```

Using `terraform import`, import EC2 Transit Gateway Metering Policies using the `transit_gateway_metering_policy_id`. For example:

```console
% terraform import aws_ec2_transit_gateway_metering_policy.example tgw-mp-12345678
```
