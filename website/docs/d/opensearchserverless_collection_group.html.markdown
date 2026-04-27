---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_collection_group"
description: |-
  Terraform data source for managing an AWS OpenSearch Serverless Collection Group.
---

# Data Source: aws_opensearchserverless_collection_group

Terraform data source for managing an AWS OpenSearch Serverless Collection Group.

## Example Usage

### Basic Usage

```terraform
data "aws_opensearchserverless_collection_group" "example" {
  name = "example-group"
}
```

## Argument Reference

The following arguments are optional:

* `id` - (Optional) ID of the collection group.
* `name` - (Optional) Name of the collection group.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

~> Specify exactly one of `id` or `name`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the collection group.
* `capacity_limits` - Capacity limits configured for the collection group. See [`capacity_limits`](#capacity_limits) below for details.
* `created_date` - Date the collection group was created.
* `description` - Description of the collection group.
* `standby_replicas` - Indicates whether standby replicas should be used for collections in this group.
* `tags` - A map of tags assigned to the collection group.

### `capacity_limits`

* `max_indexing_capacity_in_ocu` - Maximum indexing capacity, in OpenSearch Compute Units (OCUs), for the collection group.
* `max_search_capacity_in_ocu` - Maximum search capacity, in OpenSearch Compute Units (OCUs), for the collection group.
* `min_indexing_capacity_in_ocu` - Minimum indexing capacity, in OpenSearch Compute Units (OCUs), for the collection group.
* `min_search_capacity_in_ocu` - Minimum search capacity, in OpenSearch Compute Units (OCUs), for the collection group.
