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
* `availability_zone_rebalancing` - Whether Availability Zone rebalancing is enabled
* `capacity_provider_strategy` - Capacity provider strategy for the service. See [`capacity_provider_strategy` Block](#capacity_provider_strategy-block) for details.
* `created_at` - Time when the service was created (RFC3339 format)
* `created_by` - Principal that created the service
* `deployment_configuration` - Deployment configuration for the service. See [`deployment_configuration` Block](#deployment_configuration-block) for details.
* `deployment_controller` - Deployment controller configuration. See [`deployment_controller` Block](#deployment_controller-block) for details.
* `deployments` - Current deployments for the service. See [`deployments` Block](#deployments-block) for details.
* `desired_count` - Number of tasks for the ECS Service
* `enable_ecs_managed_tags` - Whether ECS managed tags are enabled
* `enable_execute_command` - Whether execute command functionality is enabled
* `events` - Recent service events. See [`events` Block](#events-block) for details.
* `health_check_grace_period_seconds` - Grace period for health checks
* `iam_role` - ARN of the IAM role associated with the service
* `launch_type` - Launch type for the ECS Service
* `load_balancer` - Load balancers for the ECS Service. See [`load_balancer` Block](#load_balancer-block) for details.
* `network_configuration` - Network configuration for the service. See [`network_configuration` Block](#network_configuration-block) for details.
* `ordered_placement_strategy` - Placement strategy for tasks. See [`ordered_placement_strategy` Block](#ordered_placement_strategy-block) for details.
* `pending_count` - Number of tasks in PENDING state
* `placement_constraints` - Placement constraints for tasks. See [`placement_constraints` Block](#placement_constraints-block) for details.
* `platform_family` - Platform family for Fargate tasks
* `platform_version` - Platform version for Fargate tasks
* `propagate_tags` - Whether tags are propagated from task definition or service
* `running_count` - Number of tasks in RUNNING state
* `scheduling_strategy` - Scheduling strategy for the ECS Service
* `service_registries` - Service discovery registries. See [`service_registries` Block](#service_registries-block) for details.
* `status` - Status of the service
* `task_definition` - Family for the latest ACTIVE revision or full ARN of the task definition
* `task_sets` - Task sets for the service. See [`task_sets` Block](#task_sets-block) for details.
* `tags` - Resource tags.

### `capacity_provider_strategy` Block

The `capacity_provider_strategy` block exports the following attributes:

* `base` - Number of tasks using the specified capacity provider
* `capacity_provider` - Name of the capacity provider
* `weight` - Relative percentage of total tasks to launch

### `deployment_configuration` Block

The `deployment_configuration` block exports the following attributes:

* `alarms` - CloudWatch alarms configuration. See [`alarms` Block](#alarms-block) for details.
* `bake_time_in_minutes` - Time to wait after deployment before terminating old tasks
* `canary_configuration` - Canary deployment configuration. See [`canary_configuration` Block](#canary_configuration-block) for details.
* `deployment_circuit_breaker` - Circuit breaker configuration. See [`deployment_circuit_breaker` Block](#deployment_circuit_breaker-block) for details.
* `linear_configuration` - Linear deployment configuration. See [`linear_configuration` Block](#linear_configuration-block) for details.
* `lifecycle_hook` - Lifecycle hooks for deployments. See [`lifecycle_hook` Block](#lifecycle_hook-block) for details.
* `maximum_percent` - Upper limit on tasks during deployment
* `minimum_healthy_percent` - Lower limit on healthy tasks during deployment
* `strategy` - Deployment strategy (ROLLING, BLUE_GREEN, LINEAR, or CANARY)

### `alarms` Block

The `alarms` block exports the following attributes:

* `alarm_names` - List of CloudWatch alarm names
* `enable` - Whether alarms are enabled
* `rollback` - Whether to rollback on alarm

### `canary_configuration` Block

The `canary_configuration` block exports the following attributes:

* `canary_bake_time_in_minutes` - Time to wait before shifting remaining traffic
* `canary_percent` - Percentage of traffic to route to canary deployment

### `deployment_circuit_breaker` Block

The `deployment_circuit_breaker` block exports the following attributes:

* `enable` - Whether circuit breaker is enabled
* `rollback` - Whether to rollback on failure

### `linear_configuration` Block

The `linear_configuration` block exports the following attributes:

* `step_bake_time_in_minutes` - Time to wait between deployment steps
* `step_percent` - Percentage of traffic to shift in each step

### `lifecycle_hook` Block

The `lifecycle_hook` block exports the following attributes:

* `hook_details` - Additional details for the hook
* `hook_target_arn` - ARN of the Lambda function to invoke
* `lifecycle_stages` - Deployment stages when hook is invoked
* `role_arn` - ARN of the IAM role for invoking the hook

### `deployment_controller` Block

The `deployment_controller` block exports the following attributes:

* `type` - Deployment controller type (ECS, CODE_DEPLOY, or EXTERNAL)

### `deployments` Block

The `deployments` block exports the following attributes:

* `created_at` - Time when deployment was created (RFC3339 format)
* `desired_count` - Desired number of tasks
* `id` - Deployment ID
* `pending_count` - Number of pending tasks
* `running_count` - Number of running tasks
* `status` - Deployment status
* `task_definition` - Task definition ARN
* `updated_at` - Time when deployment was last updated (RFC3339 format)

### `events` Block

The `events` block exports the following attributes:

* `created_at` - Time when event occurred (RFC3339 format)
* `id` - Event ID
* `message` - Event message

### `network_configuration` Block

The `network_configuration` block exports the following attributes:

* `assign_public_ip` - Whether tasks receive public IP addresses
* `security_groups` - Security groups associated with tasks
* `subnets` - Subnets associated with tasks

### `ordered_placement_strategy` Block

The `ordered_placement_strategy` block exports the following attributes:

* `field` - Field to apply placement strategy against
* `type` - Placement strategy type

### `placement_constraints` Block

The `placement_constraints` block exports the following attributes:

* `expression` - Cluster query language expression
* `type` - Constraint type

### `service_registries` Block

The `service_registries` block exports the following attributes:

* `container_name` - Container name for service discovery
* `container_port` - Container port for service discovery
* `port` - Port value for service discovery
* `registry_arn` - ARN of the service registry

### `task_sets` Block

The `task_sets` block exports the following attributes:

* `arn` - ARN of the task set
* `created_at` - Time when task set was created (RFC3339 format)
* `id` - Task set ID
* `pending_count` - Number of pending tasks
* `running_count` - Number of running tasks
* `stability_status` - Stability status of the task set
* `status` - Task set status
* `task_definition` - Task definition ARN
* `updated_at` - Time when task set was last updated (RFC3339 format)

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
