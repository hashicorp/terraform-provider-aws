---
subcategory: "VPC IPAM (IP Address Manager)"
layout: "aws"
page_title: "AWS: aws_vpc_ipams"
description: |-
  Terraform data source for managing VPC IPAMs.
---

# Data Source: aws_vpc_ipams

Terraform data source for managing VPC IPAMs.

## Example Usage

### Basic Usage

```terraform
data "aws_vpc_ipams" "example" {
  ipam_ids = ["ipam-abcd1234"]
}
```

### Filter by `tags`

```terraform
data "aws_vpc_ipams" "example" {
  filter {
    name   = "tags.Some"
    values = ["Value"]
  }
}
```

### Filter by `tier`

```terraform
data "aws_vpc_ipams" "example" {
  filter {
    name   = "tier"
    values = ["free"]
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available IPAMs.

* `ipam_ids` - (Optional) IDs of the IPAM resources to query for.
* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeIpams.html).

* `values` - (Required) Set of values that are accepted for the given field.
  An IPAM resource will be selected if any one of the given values matches.

## Attribute Reference

All of the argument attributes except `filter` are also exported as result attributes.

* `ipams` - List of IPAM resources matching the provided arguments.

### ipams

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
