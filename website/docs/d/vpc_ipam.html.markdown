---
subcategory: "VPC IPAM (IP Address Manager)"
layout: "aws"
page_title: "AWS: aws_vpc_ipam"
description: |-
  Terraform data source for managing a VPC IPAM.
---

# Data Source: aws_vpc_ipam

Terraform data source for managing a VPC IPAM.

## Example Usage

### Basic Usage

```terraform
data "aws_vpc_ipam" "example" {
  id = "ipam-abcd1234"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) ID of the IPAM.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the IPAM.
* `default_resource_discovery_association_id` - The default resource discovery association ID.
* `default_resource_discovery_id` - The default resource discovery ID.
* `description` - Description for the IPAM.
* `enable_private_gua` - If private GUA is enabled.
* `id` - ID of the IPAM resource.
* `ipam_region` - Region that the IPAM exists in.
* `operating_regions` - Regions that the IPAM is configured to operate in.
* `owner_id` - ID of the account that owns this IPAM.
* `private_default_scope_id` - ID of the default private scope.
* `public_default_scope_id` - ID of the default public scope.
* `resource_discovery_association_count` - Number of resource discovery associations.
* `scope_count` - Number of scopes on this IPAM.
* `state` - Current state of the IPAM.
* `state_message` - State message of the IPAM.
* `tier` - IPAM Tier.
* `tags` - Tags of the IPAM resource.
