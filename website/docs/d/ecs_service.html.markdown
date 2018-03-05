---
layout: "aws"
page_title: "AWS: aws_ecs_service"
sidebar_current: "docs-aws-datasource-ecs-service"
description: |-
    Provides details about an ecs service
---

# Data Source: aws_ecs_service

The ECS Service data source allows access to details of a specific
Service within a AWS ECS Cluster.

## Example Usage

```hcl
data "aws_ecs_service" "ecs-mongo" {
  service_name = "${aws_ecs_service.mongo.name}"
  cluster_arn  = "${aws_ecs_cluster.foo.arn}"
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
  cluster       = "${aws_ecs_cluster.foo.id}"
  desired_count = 2
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
}
```

## Argument Reference

The following arguments are supported:

* `service_name` - (Required) The name of the ECS Service
* `cluster_arn` - (Required) The arn of the ECS Cluster

## Attributes Reference

The following attributes are exported:

* `arn` - The ARN of the ECS Service
* `desired_count` - The number of tasks for the ECS Service
* `launch_type` - The launch type for the ECS Service
* `task_definition` - The family for the latest ACTIVE revision
