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
variable "ipam_regions" {
  type    = list
  default = ["us-east-1", "us-west-2"]
}

resource "aws_vpc_ipam" "example" {
  description = "test4"
  dynamic operating_regions {
    for_each = var.ipam_regions
    content {
      region_name = operating_regions.value
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) A description for the IPAM.
* `operating_regions` - (Required) Determines which locales can be chosen when you create pools. Locale is the Region where you want to make an IPAM pool available for allocations. You can only create pools with locales that match the operating Regions of the IPAM. You can only create VPCs from a pool whose locale matches the VPC's Region. You specify a region using the [region_name](#operating_regions) parameter. You **must** set your provider block region as an operating_region.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `cascade` - (Optional) Enables you to quickly delete an IPAM, private scopes, pools in private scopes, and any allocations in the pools in private scopes.

### operating_regions

* `region_name` - (Required) The name of the Region you want to add to the IPAM.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of IPAM
* `id` - The ID of the IPAM
* `private_default_scope_id` - The ID of the IPAM's private scope. A scope is a top-level container in IPAM. Each scope represents an IP-independent network. Scopes enable you to represent networks where you have overlapping IP space. When you create an IPAM, IPAM automatically creates two scopes: public and private. The private scope is intended for private IP space. The public scope is intended for all internet-routable IP space.
* `public_default_scope_id` - The ID of the IPAM's public scope. A scope is a top-level container in IPAM. Each scope represents an IP-independent network. Scopes enable you to represent networks where you have overlapping IP space. When you create an IPAM, IPAM automatically creates two scopes: public and private. The private scope is intended for private
IP space. The public scope is intended for all internet-routable IP space.
* `scope_count` - The number of scopes in the IPAM.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).


## Import

IPAMs can be imported using the `ipam id`, e.g.

```
$ terraform import aws_vpc_ipam.example ipam-0178368ad2146a492
```
