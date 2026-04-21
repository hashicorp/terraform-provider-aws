---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_daemon"
description: |-
    Provides details about an ECS daemon
---

# Data Source: aws_ecs_daemon

The ECS daemon data source allows access to details of a specific AWS ECS daemon.

## Example Usage

```terraform
data "aws_ecs_daemon" "example" {
  arn = "arn:aws:ecs:us-west-2:123456789012:daemon/my-cluster/my-daemon"
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Required) ARN of the ECS daemon.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `capacity_provider_arns` - List of capacity provider ARNs associated with the daemon.
* `cluster` - ARN of the ECS cluster.
* `daemon_task_definition` - ARN of the daemon task definition.
* `id` - ARN of the daemon.
* `name` - Name of the daemon.
* `status` - Status of the daemon.
* `tags` - Key-value map of resource tags.
