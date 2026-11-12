---
subcategory: "VPC IPAM (IP Address Manager)"
layout: "aws"
page_title: "AWS: aws_vpc_ipam_scope"
description: |-
  Creates a scope for AWS IPAM.
---

# Resource: aws_vpc_ipam_scope

Creates a scope for AWS IPAM.

## Example Usage

Basic usage:

```terraform
data "aws_region" "current" {}

resource "aws_vpc_ipam" "example" {
  operating_regions {
    region_name = data.aws_region.current.region
  }
}

resource "aws_vpc_ipam_scope" "example" {
  ipam_id     = aws_vpc_ipam.example.id
  description = "Another Scope"
}
```

## Argument Reference

This resource supports the following arguments:

* `description` - (Optional) Description for the scope you're creating.
* `external_authority_configuration` - (Optional) Configuration for linking the scope to an external IPAM authority. See [`external_authority_configuration` Block](#external_authority_configuration-block) below.
* `ipam_id` - (Required) ID of the IPAM for which you're creating this scope.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `external_authority_configuration` Block

The `external_authority_configuration` block supports:

* `external_resource_identifier` - (Required) Identifier of the external resource managing this scope. For Infoblox, use the Infoblox resource identifier in the format `<version>.identity.account.<entity_realm>.<entity_id>`. See [Integrate VPC IPAM with Infoblox infrastructure](https://docs.aws.amazon.com/vpc/latest/ipam/integrate-infoblox-ipam.html) for details.
* `type` - (Required) Type of external authority managing this scope. Valid value is `infoblox`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the scope.
* `id` - ID of the IPAM Scope.
* `ipam_arn` - ARN of the IPAM for which you're creating this scope.
* `ipam_scope_type` - Type of the scope.
* `is_default` - Whether the scope is the default scope.
* `pool_count` - Number of pools in the scope.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `3m`)
* `update` - (Default `3m`)
* `delete` - (Default `3m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IPAMs using the `scope_id`. For example:

```terraform
import {
  to = aws_vpc_ipam_scope.example
  id = "ipam-scope-0513c69f283d11dfb"
}
```

Using `terraform import`, import IPAMs using the `scope_id`. For example:

```console
% terraform import aws_vpc_ipam_scope.example ipam-scope-0513c69f283d11dfb
```
