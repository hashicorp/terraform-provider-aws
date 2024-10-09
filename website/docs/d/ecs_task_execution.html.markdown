---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_task_execution"
description: |-
  Terraform data source for managing an AWS ECS (Elastic Container) Task Execution.
---

# Data Source: aws_ecs_task_execution

Terraform data source for managing an AWS ECS (Elastic Container) Task Execution. This data source calls the [RunTask](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_RunTask.html) API, allowing execution of one-time tasks that don't fit a standard resource lifecycle. See the [feature request issue](https://github.com/hashicorp/terraform-provider-aws/issues/1703) for additional context.

~> **NOTE on plan operations:** This data source calls the `RunTask` API on every read operation, which means new task(s) may be created from a `terraform plan` command if all attributes are known. Placing this functionality behind a data source is an intentional trade off to enable use cases requiring a one-time task execution without relying on [provisioners](https://developer.hashicorp.com/terraform/language/resources/provisioners/syntax). Caution should be taken to ensure the data source is only executed once, or that the resulting tasks can safely run in parallel.

## Example Usage

### Basic Usage

```terraform
data "aws_ecs_task_execution" "example" {
  cluster         = aws_ecs_cluster.example.id
  task_definition = aws_ecs_task_definition.example.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.example[*].id
    security_groups  = [aws_security_group.example.id]
    assign_public_ip = false
  }
}
```

## Argument Reference

The following arguments are required:

* `cluster` - (Required) Short name or full Amazon Resource Name (ARN) of the cluster to run the task on.
* `task_definition` - (Required) The `family` and `revision` (`family:revision`) or full ARN of the task definition to run. If a revision isn't specified, the latest `ACTIVE` revision is used.

The following arguments are optional:

* `capacity_provider_strategy` - (Optional) Set of capacity provider strategies to use for the cluster. See below.
* `client_token` - (Optional) An identifier that you provide to ensure the idempotency of the request. It must be unique and is case sensitive. Up to 64 characters are allowed. The valid characters are characters in the range of 33-126, inclusive. For more information, see [Ensuring idempotency](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/ECS_Idempotency.html).
* `desired_count` - (Optional) Number of instantiations of the specified task to place on your cluster. You can specify up to 10 tasks for each call.
* `enable_ecs_managed_tags` - (Optional) Specifies whether to enable Amazon ECS managed tags for the tasks within the service.
* `enable_execute_command` - (Optional) Specifies whether to enable Amazon ECS Exec for the tasks within the service.
* `group` - (Optional) Name of the task group to associate with the task. The default value is the family name of the task definition.
* `launch_type` - (Optional) Launch type on which to run your service. Valid values are `EC2`, `FARGATE`, and `EXTERNAL`.
* `network_configuration` - (Optional) Network configuration for the service. This parameter is required for task definitions that use the `awsvpc` network mode to receive their own Elastic Network Interface, and it is not supported for other network modes. See below.
* `overrides` - (Optional) A list of container overrides that specify the name of a container in the specified task definition and the overrides it should receive.
* `placement_constraints` - (Optional) An array of placement constraint objects to use for the task. You can specify up to 10 constraints for each task. See below.
* `placement_strategy` - (Optional) The placement strategy objects to use for the task. You can specify a maximum of 5 strategy rules for each task. See below.
* `platform_version` - (Optional) The platform version the task uses. A platform version is only specified for tasks hosted on Fargate. If one isn't specified, the `LATEST` platform version is used.
* `propagate_tags` - (Optional) Specifies whether to propagate the tags from the task definition to the task. If no value is specified, the tags aren't propagated. An error will be received if you specify the `SERVICE` option when running a task. Valid values are `TASK_DEFINITION` or `NONE`.
* `reference_id` - (Optional) The reference ID to use for the task.
* `started_by` - (Optional) An optional tag specified when a task is started.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### capacity_provider_strategy

* `capacity_provider` - (Required) Name of the capacity provider.
* `base` - (Optional) The number of tasks, at a minimum, to run on the specified capacity provider. Only one capacity provider in a capacity provider strategy can have a base defined. Defaults to `0`.
* `weight` - (Optional) The relative percentage of the total number of launched tasks that should use the specified capacity provider. The `weight` value is taken into consideration after the `base` count of tasks has been satisfied. Defaults to `0`.

### network_configuration

* `subnets` - (Required) Subnets associated with the task or service.
* `security_groups` - (Optional) Security groups associated with the task or service. If you do not specify a security group, the default security group for the VPC is used.
* `assign_public_ip` - (Optional) Assign a public IP address to the ENI (Fargate launch type only). Valid values are `true` or `false`. Default `false`.

For more information, see the [Task Networking](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-networking.html) documentation.

### overrides

* `container_overrides` - (Optional) One or more container overrides that are sent to a task. See below.
* `cpu` - (Optional) The CPU override for the task.
* `execution_role_arn` - (Optional) Amazon Resource Name (ARN) of the task execution role override for the task.
* `inference_accelerator_overrides` - (Optional) Elastic Inference accelerator override for the task. See below.
* `memory` - (Optional) The memory override for the task.
* `task_role_arn` - (Optional) Amazon Resource Name (ARN) of the role that containers in this task can assume.

### container_overrides

* `command` - (Optional) The command to send to the container that overrides the default command from the Docker image or the task definition.
* `cpu` - (Optional) The number of cpu units reserved for the container, instead of the default value from the task definition.
* `environment` - (Optional) The environment variables to send to the container. You can add new environment variables, which are added to the container at launch, or you can override the existing environment variables from the Docker image or the task definition. See below.
* `memory` - (Optional) The hard limit (in MiB) of memory to present to the container, instead of the default value from the task definition. If your container attempts to exceed the memory specified here, the container is killed.
* `memory_reservation` - (Optional) The soft limit (in MiB) of memory to reserve for the container, instead of the default value from the task definition.
* `name` - (Optional) The name of the container that receives the override. This parameter is required if any override is specified.
* `resource_requirements` - (Optional) The type and amount of a resource to assign to a container, instead of the default value from the task definition. The only supported resource is a GPU. See below.

### environment

* `key` - (Required) The name of the key-value pair. For environment variables, this is the name of the environment variable.
* `value` - (Required) The value of the key-value pair. For environment variables, this is the value of the environment variable.

### resource_requirements

* `type` - (Required) The type of resource to assign to a container. Valid values are `GPU` or `InferenceAccelerator`.
* `value` - (Required) The value for the specified resource type. If the `GPU` type is used, the value is the number of physical GPUs the Amazon ECS container agent reserves for the container. The number of GPUs that's reserved for all containers in a task can't exceed the number of available GPUs on the container instance that the task is launched on. If the `InferenceAccelerator` type is used, the value matches the `deviceName` for an InferenceAccelerator specified in a task definition.

### inference_accelerator_overrides

* `device_name` - (Optional) The Elastic Inference accelerator device name to override for the task. This parameter must match a deviceName specified in the task definition.
* `device_type` - (Optional) The Elastic Inference accelerator type to use.

### placement_constraints

* `expression` - (Optional) A cluster query language expression to apply to the constraint. The expression can have a maximum length of 2000 characters. You can't specify an expression if the constraint type is `distinctInstance`.
* `type` - (Optional) The type of constraint. Valid values are `distinctInstance` or `memberOf`. Use `distinctInstance` to ensure that each task in a particular group is running on a different container instance. Use `memberOf` to restrict the selection to a group of valid candidates.

### placement_strategy

* `field` - (Optional) The field to apply the placement strategy against.
* `type` - (Optional) The type of placement strategy. Valid values are `random`, `spread`, and `binpack`.

For more information, see the [Placement Strategy](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_PlacementStrategy.html) documentation.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `task_arns` - A list of the provisioned task ARNs.
* `id` - The unique identifier, which is a comma-delimited string joining the `cluster` and `task_definition` attributes.
