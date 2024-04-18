---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_cluster"
description: |-
    Provides details about an ecs cluster
---

# Data Source: aws_ecs_cluster

The ECS Cluster data source allows access to details of a specific
cluster within an AWS ECS service.

## Example Usage

```terraform
data "aws_ecs_cluster" "ecs-mongo" {
  cluster_name = "ecs-mongo-production"
}
```

## Argument Reference

This data source supports the following arguments:

* `cluster_name` - (Required) Name of the ECS Cluster

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the ECS Cluster
* `status` - Status of the ECS Cluster
* `pending_tasks_count` - Number of pending tasks for the ECS Cluster
* `running_tasks_count` - Number of running tasks for the ECS Cluster
* `registered_container_instances_count` - The number of registered container instances for the ECS Cluster
* `service_connect_defaults` - The default Service Connect namespace
* `setting` - Settings associated with the ECS Cluster
* `tags` - Key-value map of resource tags
