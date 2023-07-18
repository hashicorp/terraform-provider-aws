---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_service"
description: |-
    Provides details about an ecs service
---

# Data Source: aws_ecs_service

The ECS Service data source allows access to details of a specific
Service within a AWS ECS Cluster.

## Example Usage

```terraform
data "aws_ecs_service" "example" {
  service_name = "example"
  cluster_arn  = data.aws_ecs_cluster.example.arn
}
```

## Argument Reference

This data source supports the following arguments:

* `service_name` - (Required) Name of the ECS Service
* `cluster_arn` - (Required) ARN of the ECS Cluster

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the ECS Service
* `desired_count` - Number of tasks for the ECS Service
* `launch_type` - Launch type for the ECS Service
* `scheduling_strategy` - Scheduling strategy for the ECS Service
* `task_definition` - Family for the latest ACTIVE revision or full ARN of the task definition.
* `tags` - Resource tags.
