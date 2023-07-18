---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_container_definition"
description: |-
    Provides details about a single container within an ecs task definition
---

# Data Source: aws_ecs_container_definition

The ECS container definition data source allows access to details of
a specific container within an AWS ECS service.

## Example Usage

```terraform
data "aws_ecs_container_definition" "ecs-mongo" {
  task_definition = aws_ecs_task_definition.mongo.id
  container_name  = "mongodb"
}
```

## Argument Reference

This data source supports the following arguments:

* `task_definition` - (Required) ARN of the task definition which contains the container
* `container_name` - (Required) Name of the container definition

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `image` - Docker image in use, including the digest
* `image_digest` - Digest of the docker image in use
* `cpu` - CPU limit for this container definition
* `memory` - Memory limit for this container definition
* `memory_reservation` - Soft limit (in MiB) of memory to reserve for the container. When system memory is under contention, Docker attempts to keep the container memory to this soft limit
* `environment` - Environment in use
* `disable_networking` - Indicator if networking is disabled
* `docker_labels` - Set docker labels
