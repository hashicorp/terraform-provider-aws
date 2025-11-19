---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_express_gateway_service"
description: |-
  Manages an ECS Express Gateway Service.
---

# Resource: aws_ecs_express_gateway_service

-> **Note:** When creating IAM roles for the Express Gateway Service in the same Terraform configuration, you should add a `time_sleep` resource to ensure proper role propagation. See the [IAM Role Timing](#iam-role-timing) section below for details.

Manages an ECS Express Gateway Service. The Express Gateway Service provides a simplified way to deploy containerized applications with built-in load balancing, auto-scaling, and networking capabilities.

## Example Usage

### Basic Usage

```terraform
resource "aws_ecs_express_gateway_service" "example" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn

  primary_container {
    image = "nginx:latest"
  }

  depends_on = [time_sleep.wait_for_iam]
}
```

### With Network Configuration

```terraform
resource "aws_ecs_express_gateway_service" "example" {
  service_name            = "my-express-service"
  cluster                 = aws_ecs_cluster.main.name
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn
  cpu                     = "256"
  memory                  = "512"

  primary_container {
    image          = "nginx:latest"
    container_port = 80
  }

  network_configuration {
    subnets         = [aws_subnet.private_a.id, aws_subnet.private_b.id]
    security_groups = [aws_security_group.app.id]
  }

  depends_on = [time_sleep.wait_for_iam]
}
```

### With Container Logging and Environment Variables

```terraform
resource "aws_ecs_express_gateway_service" "example" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn
  health_check_path       = "/health"

  primary_container {
    image          = "my-app:latest"
    container_port = 8080
    command        = ["./start.sh"]

    aws_logs_configuration {
      log_group = aws_cloudwatch_log_group.app.name
    }

    environment {
      name  = "ENV"
      value = "production"
    }

    environment {
      name  = "PORT"
      value = "8080"
    }

    secrets {
      name       = "DB_PASSWORD"
      value_from = aws_secretsmanager_secret.db_password.arn
    }
  }

  depends_on = [time_sleep.wait_for_iam]
}
```

### With Auto-Scaling

```terraform
resource "aws_ecs_express_gateway_service" "example" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn

  primary_container {
    image = "my-app:latest"
  }

  scaling_target {
    min_task_count             = 2
    max_task_count             = 10
    auto_scaling_metric        = "CPU"
    auto_scaling_target_value  = 70
  }

  depends_on = [time_sleep.wait_for_iam]
}
```

### IAM Role Timing

When creating IAM roles in the same Terraform configuration, add a `time_sleep` resource to ensure proper role propagation:

```terraform
resource "time_sleep" "wait_for_iam" {
  depends_on      = [aws_iam_role_policy_attachment.infrastructure]
  create_duration = "7s"
}

resource "aws_ecs_express_gateway_service" "example" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn

  primary_container {
    image = "nginx:latest"
  }

  depends_on = [time_sleep.wait_for_iam]
}
```

## Argument Reference

The following arguments are required:

* `execution_role_arn` - (Required) ARN of the IAM role that allows ECS to pull container images and publish container logs to Amazon CloudWatch.
* `infrastructure_role_arn` - (Required) ARN of the IAM role that allows ECS to manage AWS infrastructure on your behalf. Changing this forces a new resource to be created.

The following arguments are optional:

* `cluster` - (Optional) Name or ARN of the ECS cluster. Defaults to `default`.
* `cpu` - (Optional) Number of CPU units used by the task. Valid values are powers of 2 between 256 and 4096.
* `health_check_path` - (Optional) Path for health check requests. Defaults to `/`.
* `memory` - (Optional) Amount of memory (in MiB) used by the task. Valid values are between 512 and 8192.
* `service_name` - (Optional) Name of the service. If not specified, a name will be generated. Changing this forces a new resource to be created.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `task_role_arn` - (Optional) ARN of the IAM role that allows your Amazon ECS container task to make calls to other AWS services.
* `wait_for_steady_state` - (Optional) Whether to wait for the service to reach a steady state before considering the operation complete. Defaults to `false`.

### primary_container

The `primary_container` configuration block supports the following:

* `image` - (Required) Docker image to use for the container.
* `command` - (Optional) Command to run in the container. Overrides the default command from the Docker image.
* `container_port` - (Optional) Port on which the container listens for connections.

#### aws_logs_configuration

The `aws_logs_configuration` configuration block supports the following:

* `log_group` - (Required) CloudWatch log group name.
* `log_stream_prefix` - (Optional) Prefix for log stream names. If not specified, a default prefix will be used.

#### environment

The `environment` configuration block supports the following:

* `name` - (Required) Name of the environment variable.
* `value` - (Required) Value of the environment variable.

#### repository_credentials

The `repository_credentials` configuration block supports the following:

* `credentials_parameter` - (Required) ARN of the AWS Systems Manager parameter containing the repository credentials.

#### secrets

The `secrets` configuration block supports the following:

* `name` - (Required) Name of the secret.
* `value_from` - (Required) ARN of the AWS Secrets Manager secret or AWS Systems Manager parameter containing the secret value.

### network_configuration

The `network_configuration` configuration block supports the following:

* `security_groups` - (Optional) Security groups associated with the task. If not specified, the default security group for the VPC is used.
* `subnets` - (Optional) Subnets associated with the task. At least 2 subnets must be specified when using network configuration. If not specified, default subnets will be used.

### scaling_target

The `scaling_target` configuration block supports the following:

* `auto_scaling_metric` - (Optional) Metric to use for auto-scaling. Valid values are `CPU` and `MEMORY`.
* `auto_scaling_target_value` - (Optional) Target value for the auto-scaling metric (as a percentage).
* `max_task_count` - (Optional) Maximum number of tasks to run.
* `min_task_count` - (Optional) Minimum number of tasks to run.

### timeouts

The `timeouts` configuration block supports the following:

* `create` - (Optional) Maximum time to wait for the service to be created. Defaults to `20m`.
* `delete` - (Optional) Maximum time to wait for the service to be deleted. Defaults to `20m`.
* `update` - (Optional) Maximum time to wait for the service to be updated. Defaults to `20m`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `active_configurations` - List of active service configurations. Each configuration contains:
  * `auto_scaling_metric` - Auto-scaling metric being used.
  * `auto_scaling_target_value` - Target value for auto-scaling.
  * `cpu` - CPU units allocated to the service.
  * `created_at` - Time when the configuration was created.
  * `execution_role_arn` - Execution role ARN.
  * `health_check_path` - Health check path.
  * `ingress_paths` - List of ingress paths with access type and endpoint information.
  * `max_task_count` - Maximum number of tasks.
  * `memory` - Memory allocated to the service.
  * `min_task_count` - Minimum number of tasks.
  * `network_configuration` - Network configuration details.
  * `primary_container` - Primary container configuration.
  * `scaling_target` - Scaling target configuration.
  * `service_revision_arn` - ARN of the service revision.
  * `task_role_arn` - Task role ARN.
* `created_at` - Time when the service was created.
* `current_deployment` - ARN of the current deployment.
* `id` - ARN of the Express Gateway Service.
* `service_arn` - ARN of the Express Gateway Service.
* `status` - Current status of the service. Contains:
  * `status_code` - Status code of the service.
  * `status_reason` - Reason for the current status.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `updated_at` - Time when the service was last updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)
* `delete` - (Default `20m`)
* `update` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECS Express Gateway Services using the service ARN. For example:

```terraform
import {
  to = aws_ecs_express_gateway_service.example
  id = "arn:aws:ecs:us-west-2:123456789012:service/my-cluster/my-express-gateway-service"
}
```

Using `terraform import`, import ECS Express Gateway Services using the service ARN. For example:

```console
% terraform import aws_ecs_express_gateway_service.example arn:aws:ecs:us-west-2:123456789012:service/my-cluster/my-express-gateway-service
