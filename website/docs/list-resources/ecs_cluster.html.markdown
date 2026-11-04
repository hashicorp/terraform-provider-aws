---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_cluster"
description: |-
  Lists ECS (Elastic Container) Cluster resources.
---

# List Resource: aws_ecs_cluster

Lists ECS (Elastic Container) Cluster resources.

## Example Usage

```terraform
list "aws_ecs_cluster" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
