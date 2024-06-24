---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_task_definition"
description: |-
    Provides details about an ecs task definition
---

# Data Source: aws_ecs_task_definition

The ECS task definition data source allows access to details of
a specific AWS ECS task definition.

## Example Usage

```terraform
# Simply specify the family to find the latest ACTIVE revision in that family.
data "aws_ecs_task_definition" "mongo" {
  task_definition = aws_ecs_task_definition.mongo.family
}

resource "aws_ecs_cluster" "foo" {
  name = "foo"
}

resource "aws_ecs_task_definition" "mongo" {
  family = "mongodb"

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

resource "aws_ecs_service" "mongo" {
  name          = "mongo"
  cluster       = aws_ecs_cluster.foo.id
  desired_count = 2

  # Track the latest ACTIVE revision
  task_definition = data.aws_ecs_task_definition.mongo.arn
}
```

## Argument Reference

This data source supports the following arguments:

* `task_definition` - (Required) Family for the latest ACTIVE revision, family and revision (family:revision) for a specific revision in the family, the ARN of the task definition to access to.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ARN of the task definition.
* `arn` - ARN of the task definition.
* `arn_without_revision` - ARN of the Task Definition with the trailing `revision` removed. This may be useful for situations where the latest task definition is always desired. If a revision isn't specified, the latest ACTIVE revision is used. See the [AWS documentation](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_StartTask.html#ECS-StartTask-request-taskDefinition) for details.
* `execution_role_arn` - ARN of the task execution role that the Amazon ECS container agent and the Docker.
* `family` - Family of this task definition.
* `network_mode` - Docker networking mode to use for the containers in this task.
* `revision` - Revision of this task definition.
* `status` - Status of this task definition.
* `task_role_arn` - ARN of the IAM role that containers in this task can assume.
