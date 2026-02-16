---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_encryption_control"
description: |-
  Manages a VPC Encryption Control.
---

# Resource: aws_vpc_encryption_control

Manages a VPC Encryption Control.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpc_encryption_control" "example" {
  vpc_id = aws_vpc.example.id
  mode   = "monitor"
}

resource "aws_vpc" "example" {
  cidr_block = "10.1.0.0/16"
}
```

## Argument Reference

The following arguments are required:

* `mode` - (Required) Mode to enable for VPC Encryption Control.
  Valid values are `monitor` or `enforce`.
* `vpc_id` - (Required) The ID of the VPC the VPC Encryption Control is linked to.

The following arguments are optional:

* `egress_only_internet_gateway_exclusion` - (Optional) Whether to exclude Egress-Only Internet Gateways from encryption enforcement.
  Valid values are `disable` or `enable`.
  Default is `disable`.
  Only valid when `mode` is `enforce`.
* `elastic_file_system_exclusion` - (Optional) Whether to exclude Elastic File System (EFS) from encryption enforcement.
  Valid values are `disable` or `enable`.
  Default is `disable`.
  Only valid when `mode` is `enforce`.
* `internet_gateway_exclusion` - (Optional) Whether to exclude Internet Gateways from encryption enforcement.
  Valid values are `disable` or `enable`.
  Default is `disable`.
  Only valid when `mode` is `enforce`.
* `lambda_exclusion` - (Optional) Whether to exclude Lambda Functions from encryption enforcement.
  Valid values are `disable` or `enable`.
  Default is `disable`.
  Only valid when `mode` is `enforce`.
* `nat_gateway_exclusion` - (Optional) Whether to exclude NAT Gateways from encryption enforcement.
  Valid values are `disable` or `enable`.
  Default is `disable`.
  Only valid when `mode` is `enforce`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `virtual_private_gateway_exclusion` - (Optional) Whether to exclude Virtual Private Gateways from encryption enforcement.
  Valid values are `disable` or `enable`.
  Default is `disable`.
  Only valid when `mode` is `enforce`.
* `vpc_lattice_exclusion` - (Optional) Whether to exclude VPC Lattice from encryption enforcement.
  Valid values are `disable` or `enable`.
  Default is `disable`.
  Only valid when `mode` is `enforce`.
* `vpc_peering_exclusion` - (Optional) Whether to exclude peered VPCs from encryption enforcement.
  Valid values are `disable` or `enable`.
  Default is `disable`.
  Only valid when `mode` is `enforce`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the VPC Encryption Control.
* `resource_exclusions` - State of exclusions from encryption enforcement.
  Will be `nil` if `mode` is `monitor`.
  See [`resource_exclusions`](#resource_exclusions) below
* `state` - The current state of the VPC Encryption Control.
* `state_message` - A message providing additional information about the state of the VPC Encryption Control.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `resource_exclusions`

* `egress_only_internet_gateway` - `state` and `state_message` describing encryption enforcement state for Egress-Only Internet Gateways.
* `elastic_file_system` - `state` and `state_message` describing encryption enforcement state for Elastic File System (EFS).
* `internet_gateway` - `state` and `state_message` describing encryption enforcement state for Internet Gateways.
* `lambda` - `state` and `state_message` describing encryption enforcement state for Lambda Functions.
* `nat_gateway` - `state` and `state_message` describing encryption enforcement state for NAT Gateways.
* `virtual_private_gateway` - `state` and `state_message` describing encryption enforcement state for Virtual Private Gateways.
* `vpc_lattice` - `state` and `state_message` describing encryption enforcement state for VPC Lattice.
* `vpc_peering` - `state` and `state_message` describing encryption enforcement state for peered VPCs.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Encryption Control using the `id`. For example:

```terraform
import {
  to = aws_vpc_encryption_control.example
  id = "vpcec-12345678901234567"
}
```

Using `terraform import`, import VPC Encryption Control using the `id`. For example:

```console
% terraform import aws_vpc_encryption_control.example vpcec-12345678901234567
```
