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

* `mode` - (Required) Mode to enable for VPC Encryption Control. Valid values are `monitor` or `enforce`.
* `vpc_id` - (Required) ID of the VPC the VPC Encryption Control is linked to.

The following arguments are optional:

* `egress_only_internet_gateway_exclusion` - (Optional) Whether to exclude Egress-Only Internet Gateways from encryption enforcement. Valid values are `disable` or `enable`. Default is `disable`. Only valid when `mode` is `enforce`.
* `elastic_file_system_exclusion` - (Optional) Whether to exclude Elastic File System (EFS) from encryption enforcement. Valid values are `disable` or `enable`. Default is `disable`. Only valid when `mode` is `enforce`.
* `internet_gateway_exclusion` - (Optional) Whether to exclude Internet Gateways from encryption enforcement. Valid values are `disable` or `enable`. Default is `disable`. Only valid when `mode` is `enforce`.
* `lambda_exclusion` - (Optional) Whether to exclude Lambda Functions from encryption enforcement. Valid values are `disable` or `enable`. Default is `disable`. Only valid when `mode` is `enforce`.
* `nat_gateway_exclusion` - (Optional) Whether to exclude NAT Gateways from encryption enforcement. Valid values are `disable` or `enable`. Default is `disable`. Only valid when `mode` is `enforce`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `virtual_private_gateway_exclusion` - (Optional) Whether to exclude Virtual Private Gateways from encryption enforcement. Valid values are `disable` or `enable`. Default is `disable`. Only valid when `mode` is `enforce`.
* `vpc_lattice_exclusion` - (Optional) Whether to exclude VPC Lattice from encryption enforcement. Valid values are `disable` or `enable`. Default is `disable`. Only valid when `mode` is `enforce`.
* `vpc_peering_exclusion` - (Optional) Whether to exclude peered VPCs from encryption enforcement. Valid values are `disable` or `enable`. Default is `disable`. Only valid when `mode` is `enforce`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the VPC Encryption Control.
* `resource_exclusions` - State of exclusions from encryption enforcement. Will be `nil` if `mode` is `monitor`. See [`resource_exclusions`](#resource_exclusions-block) below.
* `state` - Current state of the VPC Encryption Control.
* `state_message` - Message providing additional information about the state of the VPC Encryption Control.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `resource_exclusions` Block

* `egress_only_internet_gateway` - Encryption enforcement state for Egress-Only Internet Gateways. See [`resource_exclusions.egress_only_internet_gateway`](#resource_exclusionsegress_only_internet_gateway-block) below.
* `elastic_file_system` - Encryption enforcement state for Elastic File System (EFS). See [`resource_exclusions.elastic_file_system`](#resource_exclusionselastic_file_system-block) below.
* `internet_gateway` - Encryption enforcement state for Internet Gateways. See [`resource_exclusions.internet_gateway`](#resource_exclusionsinternet_gateway-block) below.
* `lambda` - Encryption enforcement state for Lambda Functions. See [`resource_exclusions.lambda`](#resource_exclusionslambda-block) below.
* `nat_gateway` - Encryption enforcement state for NAT Gateways. See [`resource_exclusions.nat_gateway`](#resource_exclusionsnat_gateway-block) below.
* `virtual_private_gateway` - Encryption enforcement state for Virtual Private Gateways. See [`resource_exclusions.virtual_private_gateway`](#resource_exclusionsvirtual_private_gateway-block) below.
* `vpc_lattice` - Encryption enforcement state for VPC Lattice. See [`resource_exclusions.vpc_lattice`](#resource_exclusionsvpc_lattice-block) below.
* `vpc_peering` - Encryption enforcement state for peered VPCs. See [`resource_exclusions.vpc_peering`](#resource_exclusionsvpc_peering-block) below.

#### `resource_exclusions.egress_only_internet_gateway` Block

* `state` - Encryption enforcement state for Egress-Only Internet Gateways.
* `state_message` - Message providing additional information about the encryption enforcement state.

#### `resource_exclusions.elastic_file_system` Block

* `state` - Encryption enforcement state for Elastic File System (EFS).
* `state_message` - Message providing additional information about the encryption enforcement state.

#### `resource_exclusions.internet_gateway` Block

* `state` - Encryption enforcement state for Internet Gateways.
* `state_message` - Message providing additional information about the encryption enforcement state.

#### `resource_exclusions.lambda` Block

* `state` - Encryption enforcement state for Lambda Functions.
* `state_message` - Message providing additional information about the encryption enforcement state.

#### `resource_exclusions.nat_gateway` Block

* `state` - Encryption enforcement state for NAT Gateways.
* `state_message` - Message providing additional information about the encryption enforcement state.

#### `resource_exclusions.virtual_private_gateway` Block

* `state` - Encryption enforcement state for Virtual Private Gateways.
* `state_message` - Message providing additional information about the encryption enforcement state.

#### `resource_exclusions.vpc_lattice` Block

* `state` - Encryption enforcement state for VPC Lattice.
* `state_message` - Message providing additional information about the encryption enforcement state.

#### `resource_exclusions.vpc_peering` Block

* `state` - Encryption enforcement state for peered VPCs.
* `state_message` - Message providing additional information about the encryption enforcement state.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_vpc_encryption_control.example
  identity = {
    id = "vpcec-12345678901234567"
  }
}

resource "aws_vpc_encryption_control" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) VPC Encryption Control ID.

#### Optional

* `account_id` (String) Account ID where this resource is managed.
* `region` (String) Region where this resource is managed.

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
