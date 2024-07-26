---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_managed_prefix_list"
description: |-
    Provides details about a specific managed prefix list
---

# Data Source: aws_ec2_managed_prefix_list

`aws_ec2_managed_prefix_list` provides details about a specific AWS prefix list or
customer-managed prefix list in the current region.

## Example Usage

### Find the regional DynamoDB prefix list

```terraform
data "aws_region" "current" {}

data "aws_ec2_managed_prefix_list" "example" {
  name = "com.amazonaws.${data.aws_region.current.name}.dynamodb"
}
```

### Find a managed prefix list using filters

```terraform
data "aws_ec2_managed_prefix_list" "example" {
  filter {
    name   = "prefix-list-name"
    values = ["my-prefix-list"]
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
prefix lists. The given filters must match exactly one prefix list
whose data will be exported as attributes.

* `id` - (Optional) ID of the prefix list to select.
* `name` - (Optional) Name of the prefix list to select.
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.

### filter Configuration Block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter field. Valid values can be found in the EC2 [DescribeManagedPrefixLists](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeManagedPrefixLists.html) API Reference.
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the selected prefix list.
* `arn` - ARN of the selected prefix list.
* `name` - Name of the selected prefix list.
* `entries` - Set of entries in this prefix list. Each entry is an object with `cidr` and `description`.
* `owner_id` - Account ID of the owner of a customer-managed prefix list, or `AWS` otherwise.
* `address_family` - Address family of the prefix list. Valid values are `IPv4` and `IPv6`.
* `max_entries` - When then prefix list is managed, the maximum number of entries it supports, or null otherwise.
* `tags` - Map of tags assigned to the resource.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
