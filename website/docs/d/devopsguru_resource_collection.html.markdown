---
subcategory: "DevOps Guru"
layout: "aws"
page_title: "AWS: aws_devopsguru_resource_collection"
description: |-
  Terraform data source for managing an AWS DevOps Guru Resource Collection.
---

# Data Source: aws_devopsguru_resource_collection

Terraform data source for managing an AWS DevOps Guru Resource Collection.

## Example Usage

### Basic Usage

```terraform
data "aws_devopsguru_resource_collection" "example" {
  type = "AWS_SERVICE"
}
```

## Argument Reference

The following arguments are required:

* `type` - (Required) Type of AWS resource collection to create. Valid values are `AWS_CLOUD_FORMATION`, `AWS_SERVICE`, and `AWS_TAGS`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `cloudformation` - A collection of AWS CloudFormation stacks. See [`cloudformation`](#cloudformation-attribute-reference) below for additional details.
* `id` - Type of AWS resource collection to create (same value as `type`).
* `tags` - AWS tags used to filter the resources in the resource collection. See [`tags`](#tags-attribute-reference) below for additional details.

### `cloudformation` Attribute Reference

* `stack_names` - Array of the names of the AWS CloudFormation stacks.

### `tags` Attribute Reference

* `app_boundary_key` - An AWS tag key that is used to identify the AWS resources that DevOps Guru analyzes.
* `tag_values` - Array of tag values.
