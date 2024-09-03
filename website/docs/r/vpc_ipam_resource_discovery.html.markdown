---
subcategory: "VPC IPAM (IP Address Manager)"
layout: "aws"
page_title: "AWS: aws_vpc_ipam_resource_discovery"
description: |-
  Provides an IPAM Resource Discovery resource.
---

# Resource: aws_vpc_ipam_resource_discovery

Provides an IPAM Resource Discovery resource. IPAM Resource Discoveries are resources meant for multi-organization customers. If you wish to use a single IPAM across multiple orgs, a resource discovery can be created and shared from a subordinate organization to the management organizations IPAM delegated admin account. For a full deployment example, see `aws_vpc_ipam_resource_discovery_association` resource.

## Example Usage

Basic usage:

```terraform
data "aws_region" "current" {}

resource "aws_vpc_ipam_resource_discovery" "main" {
  description = "My IPAM Resource Discovery"
  operating_regions {
    region_name = data.aws_region.current.name
  }

  tags = {
    Test = "Main"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `description` - (Optional) A description for the IPAM Resource Discovery.
* `operating_regions` - (Required) Determines which regions the Resource Discovery will enable IPAM features for usage and monitoring. Locale is the Region where you want to make an IPAM pool available for allocations. You can only create pools with locales that match the operating Regions of the IPAM Resource Discovery. You can only create VPCs from a pool whose locale matches the VPC's Region. You specify a region using the [region_name](#operating_regions) parameter. **You must set your provider block region as an operating_region.**
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### operating_regions

* `region_name` - (Required) The name of the Region you want to add to the IPAM.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of IPAM Resource Discovery
* `id` - The ID of the IPAM Resource Discovery
* `is_default` - A boolean to identify if the Resource Discovery is the accounts default resource discovery
* `owner_id` - The account ID for the account that manages the Resource Discovery
* `ipam_resource_discovery_region` - The home region of the Resource Discovery
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IPAMs using the IPAM resource discovery `id`. For example:

```terraform
import {
  to = aws_vpc_ipam_resource_discovery.example
  id = "ipam-res-disco-0178368ad2146a492"
}
```

Using `terraform import`, import IPAMs using the IPAM resource discovery `id`. For example:

```console
% terraform import aws_vpc_ipam_resource_discovery.example ipam-res-disco-0178368ad2146a492
```
