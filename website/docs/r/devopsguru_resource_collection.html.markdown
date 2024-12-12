---
subcategory: "DevOps Guru"
layout: "aws"
page_title: "AWS: aws_devopsguru_resource_collection"
description: |-
  Terraform resource for managing an AWS DevOps Guru Resource Collection.
---
# Resource: aws_devopsguru_resource_collection

Terraform resource for managing an AWS DevOps Guru Resource Collection.

~> Only one type of resource collection (All Account Resources, CloudFormation, or Tags) can be enabled in an account at a time. To avoid persistent differences, this resource should be defined only once.

## Example Usage

### All Account Resources

```terraform
resource "aws_devopsguru_resource_collection" "example" {
  type = "AWS_SERVICE"
  cloudformation {
    stack_names = ["*"]
  }
}
```

### CloudFormation Stacks

```terraform
resource "aws_devopsguru_resource_collection" "example" {
  type = "AWS_CLOUD_FORMATION"
  cloudformation {
    stack_names = ["ExampleStack"]
  }
}
```

### Tags

```terraform
resource "aws_devopsguru_resource_collection" "example" {
  type = "AWS_TAGS"
  tags {
    app_boundary_key = "DevOps-Guru-Example"
    tag_values       = ["Example-Value"]
  }
}
```

### Tags All Resources

To analyze all resources with the `app_boundary_key` regardless of the corresponding tag value, set `tag_values` to `["*"]`.

```terraform
resource "aws_devopsguru_resource_collection" "example" {
  type = "AWS_TAGS"
  tags {
    app_boundary_key = "DevOps-Guru-Example"
    tag_values       = ["*"]
  }
}
```

## Argument Reference

The following arguments are required:

* `type` - (Required) Type of AWS resource collection to create. Valid values are `AWS_CLOUD_FORMATION`, `AWS_SERVICE`, and `AWS_TAGS`.

The following arguments are optional:

* `cloudformation` - (Optional) A collection of AWS CloudFormation stacks. See [`cloudformation`](#cloudformation-argument-reference) below for additional details.
* `tags` - (Optional) AWS tags used to filter the resources in the resource collection. See [`tags`](#tags-argument-reference) below for additional details.

### `cloudformation` Argument Reference

* `stack_names` - (Required) Array of the names of the AWS CloudFormation stacks. If `type` is `AWS_SERVICE` (all acccount resources) this array should be a single item containing a wildcard (`"*"`).

### `tags` Argument Reference

* `app_boundary_key` - (Required) An AWS tag key that is used to identify the AWS resources that DevOps Guru analyzes. All AWS resources in your account and Region tagged with this key make up your DevOps Guru application and analysis boundary. The key must begin with the prefix `DevOps-Guru-`. Any casing can be used for the prefix, but the associated tags __must use the same casing__ in their tag key.
* `tag_values` - (Required) Array of tag values. These can be used to further filter for specific resources within the application boundary. To analyze all resources tagged with the `app_boundary_key` regardless of the corresponding tag value, this array should be a single item containing a wildcard (`"*"`).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Type of AWS resource collection to create (same value as `type`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DevOps Guru Resource Collection using the `id`. For example:

```terraform
import {
  to = aws_devopsguru_resource_collection.example
  id = "AWS_CLOUD_FORMATION"
}
```

Using `terraform import`, import DevOps Guru Resource Collection using the `id`. For example:

```console
% terraform import aws_devopsguru_resource_collection.example AWS_CLOUD_FORMATION
```
