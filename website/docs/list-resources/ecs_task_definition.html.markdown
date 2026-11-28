---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_task_definition"
description: |-
  Lists ECS (Elastic Container) Task Definition resources.
---

# List Resource: aws_ecs_task_definition

Lists ECS (Elastic Container) Task Definition resources.

## Example Usage

```terraform
list "aws_ecs_task_definition" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
