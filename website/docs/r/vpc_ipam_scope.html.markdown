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
    region_name = data.aws_region.current.name
  }
}

resource "aws_vpc_ipam_scope" "example" {
  ipam_id     = aws_vpc_ipam.example.id
  description = "Another Scope"
}
```

## Argument Reference

This resource supports the following arguments:

* `ipam_id` - The ID of the IPAM for which you're creating this scope.
* `description` - (Optional) A description for the scope you're creating.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the scope.
* `id` - The ID of the IPAM Scope.
* `ipam_arn` - The ARN of the IPAM for which you're creating this scope.
* `is_default` - Defines if the scope is the default scope or not.
* `pool_count` - The number of pools in the scope.
* `type` - The type of the scope.

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
