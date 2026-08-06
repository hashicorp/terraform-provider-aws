---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_services"
description: |-
  Terraform data source for managing AWS ECS (Elastic Container) Services.
---

# Data Source: aws_ecs_services

Terraform data source for listing AWS ECS (Elastic Container) Services within a cluster.

## Example Usage

### Basic Usage

```terraform
data "aws_ecs_services" "example" {
  cluster_arn = aws_ecs_cluster.example.arn
}
```

### Filter by Launch Type

```terraform
data "aws_ecs_services" "example" {
  cluster_arn = aws_ecs_cluster.example.arn
  launch_type = "FARGATE"
}
```

## Argument Reference

This data source supports the following arguments:

* `cluster_arn` - (Required) ARN or short name of the ECS cluster to list services for.
* `launch_type` - (Optional) Launch type to filter services by. Valid values are `EC2`, `FARGATE`, and `EXTERNAL`.
* `scheduling_strategy` - (Optional) Scheduling strategy to filter services by. Valid values are `REPLICA` and `DAEMON`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `service_arns` - List of ECS service ARNs associated with the specified cluster.
