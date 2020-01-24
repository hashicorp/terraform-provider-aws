---
subcategory: "ECS"
layout: "aws"
page_title: "AWS: aws_ecs_task_definition"
description: |-
    Provides details about an ecs task definition
---

# Data Source: aws_ecs_task_definition

The ECS task definition data source allows access to details of
a specific AWS ECS task definition.


## Example Usage

```hcl
# Simply specify the family to find the latest ACTIVE revision in that family.
data "aws_ecs_task_definition" "mongo" {
  task_definition = "${aws_ecs_task_definition.test.family}"
}

resource "aws_ecs_cluster" "test" {
  name = "test"
}

resource "aws_ecs_task_definition" "test" {
  family = "test"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "environment": [{
      "name": "SECRET",
      "value": "KEY"
    }],
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "memoryReservation": 64,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name          = "test"
  cluster       = "${aws_ecs_cluster.test.id}"
  desired_count = 2

  # Track the latest ACTIVE revision
  task_definition = "${aws_ecs_task_definition.test.family}:${max("${aws_ecs_task_definition.test.revision}", "${data.aws_ecs_task_definition.test.revision}")}"
}
```

## Argument Reference

The following arguments are supported:

* `task_definition` - (Required) The family for the latest ACTIVE revision, family and revision (family:revision) for a specific revision in the family, the ARN of the task definition to access to.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `family` - The family of this task definition
* `network_mode` - The Docker networking mode to use for the containers in this task.
* `revision` - The revision of this task definition
* `status` - The status of this task definition
* `task_role_arn` - The ARN of the IAM role that containers in this task can assume
* `execution_role_arn` - The Amazon Resource Name (ARN) of the task execution role that the Amazon ECS container agent and the Docker daemon can assume.
* `volume` - A set of volume blocks that containers in your task may use.
* `placement_constraints` - A set of placement constraints rules that are taken into consideration during task placement.
* `cpu` - The number of cpu units used by the task.
* `memory` - The amount (in MiB) of memory used by the task.
* `requires_compatibilities` - A set of launch types required by the task. The valid values are `EC2` and `FARGATE`.
* `proxy_configuration` - The proxy configuration details for the App Mesh proxy.
* `tags` - Key-value mapping of resource tags
