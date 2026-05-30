---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_collection_groups"
description: |-
  Terraform data source for listing AWS OpenSearch Serverless Collection Groups.
---

# Data Source: aws_opensearchserverless_collection_groups

Terraform data source for listing AWS OpenSearch Serverless Collection Groups.

## Example Usage

### Basic Usage

```terraform
data "aws_opensearchserverless_collection_groups" "example" {}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `collection_group_summaries` - List of collection group summary objects. See [`collection_group_summaries`](#collection_group_summaries) below for details.

### `collection_group_summaries`

* `arn` - Amazon Resource Name (ARN) of the collection group.
* `capacity_limits` - Capacity limits configured for the collection group. See [`capacity_limits`](#capacity_limits) below for details.
* `created_date` - Epoch time, in milliseconds, when the collection group was created.
* `id` - Unique identifier for the collection group.
* `name` - Name of the collection group.
* `number_of_collections` - Number of collections currently associated with the collection group.
* `standby_replicas` - Indicates whether standby replicas are used for collections in the group.

### `capacity_limits`

* `max_indexing_capacity_in_ocu` - Maximum indexing capacity, in OpenSearch Compute Units (OCUs), for the collection group.
* `max_search_capacity_in_ocu` - Maximum search capacity, in OpenSearch Compute Units (OCUs), for the collection group.
* `min_indexing_capacity_in_ocu` - Minimum indexing capacity, in OpenSearch Compute Units (OCUs), for the collection group.
* `min_search_capacity_in_ocu` - Minimum search capacity, in OpenSearch Compute Units (OCUs), for the collection group.
