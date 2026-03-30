---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_service"
description: |-
  Lists ECS (Elastic Container) Service resources.
---

# List Resource: aws_ecs_service

Lists ECS (Elastic Container) Service resources.

## Example Usage

```terraform
list "aws_ecs_service" "example" {
  provider = aws
  config {
    cluster = "my-cluster"
  }
}
```

### Filter by Launch Type

```terraform
list "aws_ecs_service" "example" {
  provider = aws
  config {
    cluster     = "my-cluster"
    launch_type = "FARGATE"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `cluster` - (Required) Name or ARN of the ECS cluster to list services in.
* `launch_type` - (Optional) Launch type to filter results by. Valid values: `EC2`, `FARGATE`, `EXTERNAL`.
* `region` - (Optional) Region to query. Defaults to provider region.
