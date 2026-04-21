---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_daemon_task_definition"
description: |-
  Manages a revision of an ECS daemon task definition.
---

# Resource: aws_ecs_daemon_task_definition

Manages a revision of an ECS daemon task definition for use with daemon scheduling strategy.

## Example Usage

### Basic Example

```terraform
resource "aws_ecs_daemon_task_definition" "service" {
  family = "my-daemon-service"
  cpu    = "512"
  memory = "1024"

  container_definition {
    name      = "app"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }
}
```

### With IAM Roles

```terraform
resource "aws_iam_role" "task_execution" {
  name = "daemon-task-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role" "task" {
  name = "daemon-task-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
  })
}

resource "aws_ecs_daemon_task_definition" "service" {
  family             = "my-daemon-service"
  execution_role_arn = aws_iam_role.task_execution.arn
  task_role_arn      = aws_iam_role.task.arn
  cpu                = "512"
  memory             = "1024"

  container_definition {
    name      = "app"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }
}
```

### With Volumes

```terraform
resource "aws_ecs_daemon_task_definition" "service" {
  family = "my-daemon-service"
  cpu    = "512"
  memory = "1024"

  container_definition {
    name      = "app"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }

  volume {
    name      = "data-volume"
    host_path = "/data"
  }

  volume {
    name      = "logs-volume"
    host_path = "/var/log"
  }
}
```

### With Multiple Containers

```terraform
resource "aws_ecs_daemon_task_definition" "service" {
  family = "my-daemon-service"
  cpu    = "512"
  memory = "1024"

  container_definition {
    name      = "app"
    image     = "my-app:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }

  container_definition {
    name      = "sidecar"
    image     = "datadog/agent:latest"
    cpu       = 128
    memory    = 256
    essential = false
  }
}
```

## Argument Reference

The following arguments are required:

* `container_definition` - (Required) One or more container definition blocks. Detailed below.
* `family` - (Required) A unique name for your daemon task definition.

The following arguments are optional:

* `cpu` - (Optional) Number of cpu units used by the task.
* `execution_role_arn` - (Optional) ARN of the task execution role that the Amazon ECS container agent and the Docker daemon can assume.
* `memory` - (Optional) Amount (in MiB) of memory used by the task.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `task_role_arn` - (Optional) ARN of IAM role that allows your Amazon ECS container task to make calls to other AWS services.
* `volume` - (Optional) Repeatable configuration block for [volumes](#volume) that containers in your task may use. Detailed below.

### container_definition

* `image` - (Required) The image used to start a container.
* `name` - (Required) The name of a container.
* `command` - (Optional) The command that is passed to the container.
* `cpu` - (Optional) The number of cpu units reserved for the container.
* `depends_on` - (Optional) The dependencies defined for container startup and shutdown. Detailed below.
* `entry_point` - (Optional) The entry point that is passed to the container.
* `environment` - (Optional) The environment variables to pass to a container. Detailed below.
* `environment_file` - (Optional) A list of files containing the environment variables to pass to a container. Detailed below.
* `essential` - (Optional) If the essential parameter of a container is marked as true, and that container fails or stops for any reason, all other containers that are part of the task are stopped.
* `firelens_configuration` - (Optional) The FireLens configuration for the container. Detailed below.
* `health_check` - (Optional) The container health check command and associated configuration parameters for the container. Detailed below.
* `interactive` - (Optional) When this parameter is true, you can deploy containerized applications that require stdin or a tty to be allocated.
* `linux_parameters` - (Optional) Linux-specific modifications that are applied to the container. Detailed below.
* `log_configuration` - (Optional) The log configuration specification for the container. Detailed below.
* `memory` - (Optional) The amount (in MiB) of memory to present to the container.
* `memory_reservation` - (Optional) The soft limit (in MiB) of memory to reserve for the container.
* `mount_point` - (Optional) The mount points for data volumes in your container. Detailed below.
* `privileged` - (Optional) When this parameter is true, the container is given elevated privileges on the host container instance.
* `pseudo_terminal` - (Optional) When this parameter is true, a TTY is allocated.
* `readonly_root_filesystem` - (Optional) When this parameter is true, the container is given read-only access to its root file system.
* `repository_credentials` - (Optional) The private repository authentication credentials to use. Detailed below.
* `restart_policy` - (Optional) The restart policy for a container. Detailed below.
* `secret` - (Optional) The secrets to pass to the container. Detailed below.
* `start_timeout` - (Optional) Time duration (in seconds) to wait before giving up on resolving dependencies for a container.
* `stop_timeout` - (Optional) Time duration (in seconds) to wait before the container is forcefully killed if it doesn't exit normally on its own.
* `system_control` - (Optional) A list of namespaced kernel parameters to set in the container. Detailed below.
* `ulimit` - (Optional) A list of ulimits to set in the container. Detailed below.
* `user` - (Optional) The user to use inside the container.
* `working_directory` - (Optional) The working directory to run commands inside the container.

### volume

* `host_path` - (Optional) Path on the host container instance that is presented to the container. If not set, ECS will create a nonpersistent data volume that starts empty and is deleted after the task has finished.
* `name` - (Required) Name of the volume. This name is referenced in the `sourceVolume` parameter of container definition in the `mountPoints` section.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Full ARN of the Daemon Task Definition (including both `family` and `revision`).
* `delete_requested_at` - Timestamp when deletion was requested (if applicable).
* `registered_at` - Timestamp when the daemon task definition was registered.
* `registered_by` - Principal that registered the daemon task definition.
* `revision` - Revision of the task in a particular family.
* `status` - Status of the daemon task definition.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ecs_daemon_task_definition.example
  identity = {
    "arn" = "arn:aws:ecs:us-east-1:012345678910:daemon-task-definition/mydaemonfamily:123"
  }
}

resource "aws_ecs_daemon_task_definition" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the ECS Daemon Task Definition.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECS Daemon Task Definitions using their ARNs. For example:

```terraform
import {
  to = aws_ecs_daemon_task_definition.example
  id = "arn:aws:ecs:us-east-1:012345678910:daemon-task-definition/mydaemonfamily:123"
}
```

Using `terraform import`, import ECS Daemon Task Definitions using their ARNs. For example:

```console
% terraform import aws_ecs_daemon_task_definition.example arn:aws:ecs:us-east-1:012345678910:daemon-task-definition/mydaemonfamily:123
```
