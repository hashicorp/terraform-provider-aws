---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_tag"
description: |-
  Manages an individual ECS resource tag
---

# Resource: aws_ecs_tag

Manages an individual ECS resource tag. This resource should only be used in cases where ECS resources are created outside Terraform (e.g., ECS Clusters implicitly created by Batch Compute Environments).

~> **NOTE:** This tagging resource should not be combined with the Terraform resource for managing the parent resource. For example, using `aws_ecs_cluster` and `aws_ecs_tag` to manage tags of the same ECS Cluster will cause a perpetual difference where the `aws_ecs_cluster` resource will try to remove the tag being added by the `aws_ecs_tag` resource.

~> **NOTE:** This tagging resource does not use the [provider `ignore_tags` configuration](/docs/providers/aws/index.html#ignore_tags).

## Example Usage

```terraform
resource "aws_batch_compute_environment" "example" {
  compute_environment_name = "example"
  service_role             = aws_iam_role.example.arn
  type                     = "UNMANAGED"
}

resource "aws_ecs_tag" "example" {
  resource_arn = aws_batch_compute_environment.example.ecs_cluster_arn
  key          = "Name"
  value        = "Hello World"
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_arn` - (Required) Amazon Resource Name (ARN) of the ECS resource to tag.
* `key` - (Required) Tag name.
* `value` - (Required) Tag value.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ECS resource identifier and key, separated by a comma (`,`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ecs_tag` using the ECS resource identifier and key, separated by a comma (`,`). For example:

```terraform
import {
  to = aws_ecs_tag.example
  id = "arn:aws:ecs:us-east-1:123456789012:cluster/example,Name"
}
```

Using `terraform import`, import `aws_ecs_tag` using the ECS resource identifier and key, separated by a comma (`,`). For example:

```console
% terraform import aws_ecs_tag.example arn:aws:ecs:us-east-1:123456789012:cluster/example,Name
```
