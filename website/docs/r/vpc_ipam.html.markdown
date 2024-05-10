---
subcategory: "VPC IPAM (IP Address Manager)"
layout: "aws"
page_title: "AWS: aws_vpc_ipam"
description: |-
  Provides an IPAM resource.
---

# Resource: aws_vpc_ipam

Provides an IPAM resource.

## Example Usage

Basic usage:

```terraform
data "aws_region" "current" {}

resource "aws_vpc_ipam" "main" {
  description = "My IPAM"
  operating_regions {
    region_name = data.aws_region.current.name
  }

  tags = {
    Test = "Main"
  }
}
```

Shared with multiple operating_regions:

```terraform
resource "aws_vpc_ipam" "main" {
  description = "multi region ipam"
  dynamic operating_regions {
    for_each = local.all_ipam_regions
    content {
      region_name = operating_regions.value
    }
  }
}

data "aws_region" "current" {}

variable "ipam_regions" {
  type    = list
  default = ["us-east-1", "us-west-2"]
}

locals {
  # ensure current provider region is an operating_regions entry
  all_ipam_regions = distinct(concat([data.aws_region.current.name], var.ipam_regions))
}
```

## Argument Reference

This resource supports the following arguments:

* `cascade` - (Optional) Enables you to quickly delete an IPAM, private scopes, pools in private scopes, and any allocations in the pools in private scopes.
* `description` - (Optional) A description for the IPAM.
* `operating_regions` - (Required) Determines which locales can be chosen when you create pools. Locale is the Region where you want to make an IPAM pool available for allocations. You can only create pools with locales that match the operating Regions of the IPAM. You can only create VPCs from a pool whose locale matches the VPC's Region. You specify a region using the [region_name](#operating_regions) parameter. You **must** set your provider block region as an operating_region.
* `tier` - (Optional) specifies the IPAM tier. Valid options include `free` and `advanced`. Default is `advanced`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### operating_regions

* `region_name` - (Required) The name of the Region you want to add to the IPAM.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of IPAM
* `id` - The ID of the IPAM
* `default_resource_discovery_id` - The IPAM's default resource discovery ID.
* `default_resource_discovery_association_id` - The IPAM's default resource discovery association ID.
* `private_default_scope_id` - The ID of the IPAM's private scope. A scope is a top-level container in IPAM. Each scope represents an IP-independent network. Scopes enable you to represent networks where you have overlapping IP space. When you create an IPAM, IPAM automatically creates two scopes: public and private. The private scope is intended for private IP space. The public scope is intended for all internet-routable IP space.
* `public_default_scope_id` - The ID of the IPAM's public scope. A scope is a top-level container in IPAM. Each scope represents an IP-independent network. Scopes enable you to represent networks where you have overlapping IP space. When you create an IPAM, IPAM automatically creates two scopes: public and private. The private scope is intended for private
IP space. The public scope is intended for all internet-routable IP space.
* `scope_count` - The number of scopes in the IPAM.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IPAMs using the IPAM `id`. For example:

```terraform
import {
  to = aws_vpc_ipam.example
  id = "ipam-0178368ad2146a492"
}
```

Using `terraform import`, import IPAMs using the IPAM `id`. For example:

```console
% terraform import aws_vpc_ipam.example ipam-0178368ad2146a492
```
