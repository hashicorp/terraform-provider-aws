---
subcategory: "VPC IPAM (IP Address Manager)"
layout: "aws"
page_title: "AWS: aws_vpc_ipam_default_scope"
description: |-
  Manage a default scope for AWS IPAM.
---

# Resource: aws_vpc_ipam_default_scope

Provides a resource to manage [default AWS IPAM Scopes](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_IpamScope.html)

**This is an advanced resource** and has special caveats to be aware of when using it. Please read this document in its entirety before using this resource.

The `aws_vpc_ipam_default_scope` resource behaves differently from normal resources in that if a default IPAM scope exists, Terraform does not _create_ this resource, but instead "adopts" it into management.

Default IPAM scopes cannot be deleted, therefore `terraform destroy` does not delete the default IPAM Scope but does remove the resource from Terraform state. To truly delete you must delete the parent IPAM.

## Example Usage

Basic usage:

```terraform
data "aws_region" "current" {}

resource "aws_vpc_ipam" "example" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}

resource "aws_vpc_ipam_default_scope" "example" {
  default_scope_id = aws_vpc_ipam.example.private_default_scope_id

  tags = {
    terraform = true
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `default_scope_id` - The ID of the IPAM default scope for which you're managing.

* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the scope.
* `id` - The ID of the IPAM Scope.
* `ipam_arn` - The ARN of the IPAM scope being managed.
* `ipam_id` - The ID of the IPAM scope being managed.
* `description` - A description for the scope under management.
* `is_default` - Defines if the scope is the default scope or not.
* `pool_count` - The number of pools in the scope.
* `type` - The type of the scope.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IPAMs using the `scope_id`. For example:

```terraform
import {
  to = aws_vpc_ipam_default_scope.example
  id = "ipam-scope-0513c69f283d11dfb"
}
```

Using `terraform import`, import IPAMs using the `scope_id`. For example:

```console
% terraform import aws_vpc_ipam_default_scope.example ipam-scope-0513c69f283d11dfb
```
