---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_task_set"
description: |-
  Provides an ECS task set.
---

# Resource: aws_ecs_task_set

Provides an ECS task set - effectively a task that is expected to run until an error occurs or a user terminates it (typically a webserver or a database).

See [ECS Task Set section in AWS developer guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/deployment-type-external.html).

## Example Usage

```terraform
resource "aws_ecs_task_set" "example" {
  service         = aws_ecs_service.example.id
  cluster         = aws_ecs_cluster.example.id
  task_definition = aws_ecs_task_definition.example.arn

  load_balancer {
    target_group_arn = aws_lb_target_group.example.arn
    container_name   = "mongo"
    container_port   = 8080
  }
}
```

### Ignoring Changes to Scale

You can utilize the generic Terraform resource [lifecycle configuration block](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html) with `ignore_changes` to create an ECS service with an initial count of running instances, then ignore any changes to that count caused externally (e.g. Application Autoscaling).

```terraform
resource "aws_ecs_task_set" "example" {
  # ... other configurations ...

  # Example: Run 50% of the servcie's desired count
  scale {
    value = 50.0
  }

  # Optional: Allow external changes without Terraform plan difference
  lifecycle {
    ignore_changes = ["scale"]
  }
}
```

## Argument Reference

The following arguments are required:

* `cluster` - (Required) Short name or ARN of the cluster that hosts the service to create the task set in.
* `service` - (Required) Short name or ARN of the ECS service.
* `task_definition` - (Required) Family and revision (`family:revision`) or full ARN of the task definition to run in your service.

The following arguments are optional:

* `capacity_provider_strategy` - (Optional) Capacity provider strategy to use for the service. Can be one or more. [Defined below](#capacity_provider_strategy).
* `external_id` - (Optional) External ID associated with the task set.
* `force_delete` - (Optional) Whether to allow deleting the task set without waiting for scaling down to 0. You can force a task set to delete even if it's in the process of scaling a resource. Normally, Terraform drains all the tasks before deleting the task set. This bypasses that behavior and potentially leaves resources dangling.
* `launch_type` - (Optional) Launch type on which to run your service. Valid values are `EC2`, `FARGATE`, and `EXTERNAL`. Defaults to `EC2`.
* `load_balancer` - (Optional) Details on load balancers that are used with a task set. [Detailed below](#load_balancer).
* `network_configuration` - (Optional) Network configuration for the service. Required for task definitions that use the `awsvpc` network mode to receive their own Elastic Network Interface, and not supported for other network modes. [Detailed below](#network_configuration).
* `platform_version` - (Optional) Platform version on which to run your service. Only applicable for `launch_type` set to `FARGATE`. Defaults to `LATEST`. More information about Fargate platform versions can be found in the [AWS ECS User Guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/platform_versions.html).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `scale` - (Optional) Floating-point percentage of the desired number of tasks to place and keep running in the task set. [Detailed below](#scale).
* `service_registries` - (Optional) Service discovery registries for the service. The maximum number of `service_registries` blocks is `1`. [Detailed below](#service_registries).
* `tags` - (Optional) Map of tags to assign to the file system. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level. If you have set `copy_tags_to_backups` to true, and you specify one or more tags, no existing file system tags are copied from the file system to the backup.
* `wait_until_stable` - (Optional) Whether `terraform` should wait until the task set has reached `STEADY_STATE`.
* `wait_until_stable_timeout` - (Optional) Wait timeout for task set to reach `STEADY_STATE`. Valid time units include `ns`, `us` (or `µs`), `ms`, `s`, `m`, and `h`. Default `10m`.

### `capacity_provider_strategy` Block

The `capacity_provider_strategy` configuration block supports the following:

* `base` - (Optional) Number of tasks, at a minimum, to run on the specified capacity provider. Only one capacity provider in a capacity provider strategy can have a base defined.
* `capacity_provider` - (Required) Short name or full Amazon Resource Name (ARN) of the capacity provider.
* `weight` - (Required) Relative percentage of the total number of launched tasks that should use the specified capacity provider.

### `load_balancer` Block

The `load_balancer` configuration block supports the following:

* `container_name` - (Required) Name of the container to associate with the load balancer (as it appears in a container definition).
* `container_port` - (Optional) Port on the container to associate with the load balancer. Defaults to `0` if not specified.
* `load_balancer_name` - (Optional, Required for ELB Classic) Name of the ELB (Classic) to associate with the service.
* `target_group_arn` - (Optional, Required for ALB/NLB) ARN of the Load Balancer target group to associate with the service.

~> **Note:** Specifying multiple `load_balancer` configurations is still not supported by AWS for ECS task set.

### `network_configuration` Block

The `network_configuration` configuration block supports the following. For more information, see [Task Networking](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-networking.html).

* `assign_public_ip` - (Optional) Whether to assign a public IP address to the ENI (`FARGATE` launch type only). Valid values are `true` or `false`. Default `false`.
* `security_groups` - (Optional) Security groups associated with the task or service. If you do not specify a security group, the default security group for the VPC is used. Maximum of 5.
* `subnets` - (Required) Subnets associated with the task or service. Maximum of 16.

### `scale` Block

The `scale` configuration block supports the following:

* `unit` - (Optional) Unit of measure for the scale value. Default: `PERCENT`.
* `value` - (Optional) Value, specified as a percent total of a service's `desiredCount`, to scale the task set. Defaults to `0` if not specified. Accepted values are numbers between 0.0 and 100.0.

### `service_registries` Block

The `service_registries` configuration block supports the following:

* `container_name` - (Optional) Container name value, already specified in the task definition, to be used for your service discovery service.
* `container_port` - (Optional) Port value, already specified in the task definition, to be used for your service discovery service.
* `port` - (Optional) Port value used if your Service Discovery service specified an SRV record.
* `registry_arn` - (Required) ARN of the Service Registry. The currently supported service registry is Amazon Route 53 Auto Naming Service ([`aws_service_discovery_service` resource](/docs/providers/aws/r/service_discovery_service.html)). For more information, see [Service](https://docs.aws.amazon.com/Route53/latest/APIReference/API_autonaming_Service.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) that identifies the task set.
* `id` - `task_set_id`, `service` and `cluster` separated by commas (`,`).
* `stability_status` - Stability status. This indicates whether the task set has reached a steady state.
* `status` - Status of the task set.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `task_set_id` - ID of the task set.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECS Task Sets using the `task_set_id`, `service`, and `cluster` separated by commas (`,`). For example:

```terraform
import {
  to = aws_ecs_task_set.example
  id = "ecs-svc/7177320696926227436,arn:aws:ecs:us-west-2:123456789101:service/example/example-1234567890,arn:aws:ecs:us-west-2:123456789101:cluster/example"
}
```

Using `terraform import`, import ECS Task Sets using the `task_set_id`, `service`, and `cluster` separated by commas (`,`). For example:

```console
% terraform import aws_ecs_task_set.example ecs-svc/7177320696926227436,arn:aws:ecs:us-west-2:123456789101:service/example/example-1234567890,arn:aws:ecs:us-west-2:123456789101:cluster/example
```
