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

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `service_name` - (Required) Name of the ECS Service
* `cluster_arn` - (Required) ARN of the ECS Cluster

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the ECS Service
* `desired_count` - Number of tasks for the ECS Service
* `launch_type` - Launch type for the ECS Service
* `load_balancer` - Load balancers for the ECS Service. See [`load_balancer` Block](#load_balancer-block) for details.
* `scheduling_strategy` - Scheduling strategy for the ECS Service
* `task_definition` - Family for the latest ACTIVE revision or full ARN of the task definition.
* `tags` - Resource tags.

### `load_balancer` Block

The `load_balancer` block exports the following attributes:

* `advanced_configuration` - Settings for Blue/Green deployment. See [`advanced_configuration` Block](#advanced_configuration-block) for details.
* `container_name` - Name of the container to associate with the load balancer.
* `container_port` - Port on the container to associate with the load balancer.
* `elb_name` - Name of the load balancer.
* `target_group_arn` - ARN of the target group to associate with the load balancer.

### `advanced_configuration` Block

The `advanced_configuration` block exports the following attributes:

* `alternate_target_group_arn` - ARN of the alternate target group to use for Blue/Green deployments.
* `production_listener_rule` - ARN of the listener rule that routes production traffic.
* `role_arn` - ARN of the IAM role that allows ECS to manage the target groups.
* `test_listener_rule` - ARN of the listener rule that routes test traffic.
