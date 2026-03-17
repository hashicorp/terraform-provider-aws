---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_express_gateway_service"
description: |-
  Manages an ECS Express Gateway Service.
---

# Resource: aws_ecs_express_gateway_service

Manages an ECS Express service. The Express service provides a simplified way to deploy containerized applications with automatic provisioning and management of AWS infrastructure including Application Load Balancers (ALBs), target groups, security groups, and auto-scaling policies. This service offers built-in load balancing, auto-scaling, and networking capabilities with zero-downtime deployments.

Express services automatically handle infrastructure provisioning and updates through rolling deployments, ensuring high availability during service modifications. When you update an Express service, a new service revision is created and deployed with zero downtime.

## Example Usage

### Basic Usage

```terraform
resource "aws_ecs_express_gateway_service" "example" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn

  primary_container {
    image = "nginx:latest"
  }
}
```

## Service Updates and Deletion

### Updates

When you update an Express service configuration, a new service revision is created and deployed using a rolling deployment strategy with zero downtime. The service automatically manages the transition from the old configuration to the new one, ensuring continuous availability.

### Deletion

When an Express service is deleted, it enters a `DRAINING` state where existing tasks are allowed to complete gracefully before termination. The deletion process is irreversible - once initiated, the service and all its associated AWS infrastructure (load balancers, target groups, etc.) will be permanently removed. During the draining process, no new tasks are started, and the service becomes unavailable once all tasks have completed.

## Argument Reference

The following arguments are required:

* `execution_role_arn` - (Required) ARN of the IAM role that allows ECS to pull container images and publish container logs to Amazon CloudWatch.
* `infrastructure_role_arn` - (Required) ARN of the IAM role that allows ECS to manage AWS infrastructure on your behalf. **Important:** The infrastructure role cannot be modified after the service is created. Changing this forces a new resource to be created.

The following arguments are optional:

* `cluster` - (Optional) Name or ARN of the ECS cluster. Defaults to `default`.
* `cpu` - (Optional) Number of CPU units used by the task. Valid values are powers of 2 between 256 and 4096.
* `health_check_path` - (Optional) Path for health check requests. Defaults to `/ping`.
* `memory` - (Optional) Amount of memory (in MiB) used by the task. Valid values are between 512 and 8192.
* `region` - (Optional) AWS region where the service will be created. If not specified, the region configured in the provider will be used.
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

### Example with Container Logging and Environment Variables

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

    secret {
      name       = "DB_PASSWORD"
      value_from = aws_secretsmanager_secret.db_password.arn
    }
  }
}
```

### network_configuration

The `network_configuration` configuration block supports the following:

* `security_groups` - (Optional) Security groups associated with the task. If not specified, the default security group for the VPC is used.
* `subnets` - (Optional) Subnets associated with the task. At least 2 subnets must be specified when using network configuration. If not specified, default subnets will be used.

### Example

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
}
```

### scaling_target

The `scaling_target` configuration block supports the following:

* `auto_scaling_metric` - (Optional) Metric to use for auto-scaling. Valid values are `CPU` and `MEMORY`.
* `auto_scaling_target_value` - (Optional) Target value for the auto-scaling metric (as a percentage). Defaults to `60`.
* `max_task_count` - (Optional) Maximum number of tasks to run.
* `min_task_count` - (Optional) Minimum number of tasks to run.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `current_deployment` - ARN of the current deployment.
* `ingress_paths` - List of ingress paths with access type and endpoint information.
* `service_arn` - ARN of the Express Gateway Service.
* `service_revision_arn` - ARN of the service revision.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `20m`)

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
```
