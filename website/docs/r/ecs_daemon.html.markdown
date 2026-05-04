---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_daemon"
description: |-
  Provides an ECS Daemon resource.
---

# Resource: aws_ecs_daemon

Provides an ECS Daemon resource, which manages a daemon that runs exactly one task on each container instance in an ECS cluster.

## Example Usage

### Basic Usage

```terraform
resource "aws_ecs_daemon" "example" {
  name                   = "example-daemon"
  cluster_arn            = aws_ecs_cluster.example.arn
  daemon_task_definition = aws_ecs_daemon_task_definition.example.arn
  capacity_provider_arns = [aws_ecs_capacity_provider.example.arn]
}
```

### With Deployment Configuration

```terraform
resource "aws_ecs_daemon" "example" {
  name                   = "example-daemon"
  cluster_arn            = aws_ecs_cluster.example.arn
  daemon_task_definition = aws_ecs_daemon_task_definition.example.arn
  capacity_provider_arns = [aws_ecs_capacity_provider.example.arn]

  deployment_configuration {
    drain_percent        = 50.0
    bake_time_in_minutes = 10

    alarms {
      alarm_names = ["example-alarm"]
      enable      = true
    }
  }
}
```

### With Tags

```terraform
resource "aws_ecs_daemon" "example" {
  name                   = "example-daemon"
  cluster_arn            = aws_ecs_cluster.example.arn
  daemon_task_definition = aws_ecs_daemon_task_definition.example.arn
  capacity_provider_arns = [aws_ecs_capacity_provider.example.arn]

  tags = {
    Environment = "production"
    Application = "example"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `capacity_provider_arns` - (Required) List of capacity provider ARNs to use for the daemon.
* `cluster_arn` - (Optional, Forces new resource) ARN of the ECS cluster where the daemon will run.
* `daemon_task_definition` - (Required) ARN of the daemon task definition to use for the daemon. Drift is not detected on this attribute because the API may report a stale revision while a deployment is in progress.
* `deployment_configuration` - (Optional) Configuration for daemon deployments. See [Deployment Configuration](#deployment-configuration) below.
* `enable_ecs_managed_tags` - (Optional, Write-only) Whether to enable Amazon ECS managed tags for the tasks within the daemon.
* `enable_execute_command` - (Optional, Write-only) Whether to enable Amazon ECS Exec for the tasks within the daemon.
* `name` - (Required, Forces new resource) Name of the daemon.
* `propagate_tags` - (Optional) Whether to propagate tags from the daemon to tasks. Valid values are `DAEMON` or `NONE`.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Deployment Configuration

~> **Note:** The `deployment_configuration` block is write-only. The API does not return these values, so Terraform cannot detect drift. Changes to deployment configuration will always show in the plan.

The `deployment_configuration` block supports:

* `alarms` - (Optional) Alarm configuration for deployment monitoring. See [Alarms](#alarms) below.
* `bake_time_in_minutes` - (Optional) Time in minutes to wait before considering a deployment successful. Valid values are between 0 and 1440.
* `drain_percent` - (Optional) Percentage of tasks to drain during deployment. Valid values are between 0.0 and 100.0.

### Alarms

The `alarms` block supports:

* `alarm_names` - (Required) Set of CloudWatch alarm names to monitor during deployment.
* `enable` - (Required) Whether to enable alarm monitoring for deployments.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the daemon.
* `status` - Status of the daemon. Valid values are `ACTIVE` or `DELETE_IN_PROGRESS`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)
* `update` - (Default `20m`)
* `delete` - (Default `20m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ecs_daemon.example
  identity = {
    arn = "arn:aws:ecs:us-east-1:123456789012:daemon/example-cluster/example-daemon"
  }
}

resource "aws_ecs_daemon" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the ECS Daemon.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECS Daemons using the ARN. For example:

```terraform
import {
  to = aws_ecs_daemon.example
  id = "arn:aws:ecs:us-east-1:123456789012:daemon/example-cluster/example-daemon"
}
```

Using `terraform import`, import ECS Daemons using the ARN. For example:

```console
% terraform import aws_ecs_daemon.example arn:aws:ecs:us-east-1:123456789012:daemon/example-cluster/example-daemon
```
