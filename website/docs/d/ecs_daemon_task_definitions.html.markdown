---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_daemon_task_definitions"
description: |-
    Provides a list of ECS daemon task definitions
---

# Data Source: aws_ecs_daemon_task_definitions

The ECS daemon task definitions data source allows access to a list of AWS ECS daemon task definitions in a specific region.

## Example Usage

### List All Daemon Task Definitions

```terraform
data "aws_ecs_daemon_task_definitions" "all" {}
```

### Filter by Family Prefix

```terraform
data "aws_ecs_daemon_task_definitions" "monitoring" {
  family_prefix = "monitoring-"
}
```

### Filter by Status

```terraform
data "aws_ecs_daemon_task_definitions" "active" {
  status = "ACTIVE"
}
```

### Combine Filters

```terraform
data "aws_ecs_daemon_task_definitions" "active_monitoring" {
  family_prefix = "monitoring-"
  status        = "ACTIVE"
}
```

## Argument Reference

This data source supports the following arguments:

* `family` - (Optional) Exact family name to filter daemon task definitions.
* `family_prefix` - (Optional) Prefix to filter daemon task definitions by family name.
* `revision` - (Optional) Revision filter. Valid values are `ALL` or `LAST_REGISTERED`.
* `sort` - (Optional) Sort order. Valid values are `ASC` or `DESC`.
* `status` - (Optional) Status to filter daemon task definitions. Valid values are `ACTIVE`, `DELETE_IN_PROGRESS`, or `ALL`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `daemon_task_definitions` - List of daemon task definition summaries. Each element contains:
    * `arn` - ARN of the daemon task definition.
    * `delete_requested_at` - Timestamp when deletion was requested (if applicable).
    * `registered_at` - Timestamp when the daemon task definition was registered.
    * `registered_by` - Principal that registered the daemon task definition.
    * `status` - Status of the daemon task definition.
