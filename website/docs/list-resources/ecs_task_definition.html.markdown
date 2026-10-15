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

* `region` - (Optional) Region to query. Defaults to provider region.
