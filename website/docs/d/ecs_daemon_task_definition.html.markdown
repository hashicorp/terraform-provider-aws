---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_daemon_task_definition"
description: |-
    Provides details about an ECS daemon task definition
---

# Data Source: aws_ecs_daemon_task_definition

The ECS daemon task definition data source allows access to details of a specific AWS ECS daemon task definition.

## Example Usage

```terraform
data "aws_ecs_daemon_task_definition" "example" {
  daemon_task_definition = "arn:aws:ecs:us-west-2:123456789012:daemon-task-definition/my-daemon-family:1"
}
```

## Argument Reference

This data source supports the following arguments:

* `daemon_task_definition` - (Required) ARN of the daemon task definition to access.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the daemon task definition.
* `container_definition` - The container definitions for the daemon task definition.
* `cpu` - Number of cpu units used by the task.
* `delete_requested_at` - Timestamp when deletion was requested (if applicable).
* `execution_role_arn` - ARN of the task execution role that the Amazon ECS container agent and the Docker daemon can assume.
* `family` - Family name of the daemon task definition.
* `memory` - Amount (in MiB) of memory used by the task.
* `registered_at` - Timestamp when the daemon task definition was registered.
* `registered_by` - Principal that registered the daemon task definition.
* `revision` - Revision of the task in a particular family.
* `status` - Status of the daemon task definition.
* `task_role_arn` - ARN of IAM role that allows your Amazon ECS container task to make calls to other AWS services.
* `volume` - Configuration block for [volumes](#volume) that containers in your task may use. Detailed below.

### volume

* `host_path` - Path on the host container instance that is presented to the container.
* `name` - Name of the volume.
