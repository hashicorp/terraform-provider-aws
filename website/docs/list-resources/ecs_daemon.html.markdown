---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_daemon"
description: |-
  Lists ECS (Elastic Container) Daemon resources.
---

# List Resource: aws_ecs_daemon

Lists ECS (Elastic Container) Daemon resources.

## Example Usage

```terraform
list "aws_ecs_daemon" "example" {
  provider = aws

  config {
    cluster_arn = aws_ecs_cluster.example.arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `cluster_arn` - (Required) ARN of the ECS cluster to list daemons from.
* `region` - (Optional) Region to query. Defaults to provider region.
