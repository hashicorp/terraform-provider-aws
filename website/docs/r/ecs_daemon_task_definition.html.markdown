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
resource "aws_ecs_daemon_task_definition" "example" {
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

resource "aws_ecs_daemon_task_definition" "example" {
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
resource "aws_ecs_daemon_task_definition" "example" {
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
    name = "data-volume"
    host {
      source_path = "/data"
    }
  }

  volume {
    name = "logs-volume"
    host {
      source_path = "/var/log"
    }
  }
}
```

### With Multiple Containers

```terraform
resource "aws_ecs_daemon_task_definition" "example" {
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
* `family` - (Required) Unique name for your daemon task definition.

The following arguments are optional:

* `cpu` - (Optional) Number of CPU units used by the task.
* `execution_role_arn` - (Optional) ARN of the task execution role that the Amazon ECS container agent and the Docker daemon can assume.
* `memory` - (Optional) Amount (in MiB) of memory used by the task.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `task_role_arn` - (Optional) ARN of IAM role that allows your Amazon ECS container task to make calls to other AWS services.
* `volume` - (Optional) Repeatable configuration block for volumes that containers in your task may use. Detailed below.

### `container_definition` Block

* `command` - (Optional) Command that is passed to the container.
* `cpu` - (Optional) Number of CPU units reserved for the container.
* `depends_on` - (Optional) Dependencies defined for container startup and shutdown. Detailed below.
* `entry_point` - (Optional) Entry point that is passed to the container.
* `environment` - (Optional) Environment variables to pass to a container. Detailed below.
* `environment_file` - (Optional) List of files containing the environment variables to pass to a container. Detailed below.
* `essential` - (Optional) If the essential parameter of a container is marked as true, and that container fails or stops for any reason, all other containers that are part of the task are stopped.
* `firelens_configuration` - (Optional) FireLens configuration for the container. Detailed below.
* `health_check` - (Optional) Container health check command and associated configuration parameters for the container. Detailed below.
* `image` - (Required) Image used to start a container.
* `interactive` - (Optional) When this parameter is true, you can deploy containerized applications that require stdin or a tty to be allocated.
* `linux_parameters` - (Optional) Linux-specific modifications that are applied to the container. Detailed below.
* `log_configuration` - (Optional) Log configuration specification for the container. Detailed below.
* `memory` - (Optional) Amount (in MiB) of memory to present to the container.
* `memory_reservation` - (Optional) Soft limit (in MiB) of memory to reserve for the container.
* `mount_point` - (Optional) Mount points for data volumes in your container. Detailed below.
* `name` - (Optional) Name of a container.
* `privileged` - (Optional) When this parameter is true, the container is given elevated privileges on the host container instance.
* `pseudo_terminal` - (Optional) When this parameter is true, a TTY is allocated.
* `readonly_root_filesystem` - (Optional) When this parameter is true, the container is given read-only access to its root file system.
* `repository_credentials` - (Optional) Private repository authentication credentials to use. Detailed below.
* `restart_policy` - (Optional) Restart policy for a container. Detailed below.
* `secret` - (Optional) Secrets to pass to the container. Detailed below.
* `start_timeout` - (Optional) Time duration (in seconds) to wait before giving up on resolving dependencies for a container.
* `stop_timeout` - (Optional) Time duration (in seconds) to wait before the container is forcefully killed if it doesn't exit normally on its own.
* `system_control` - (Optional) List of namespaced kernel parameters to set in the container. Detailed below.
* `ulimit` - (Optional) List of ulimits to set in the container. Detailed below.
* `user` - (Optional) User to use inside the container.
* `working_directory` - (Optional) Working directory to run commands inside the container.

### `volume` Block

* `host` - (Optional) Configuration for a host volume. Detailed below.
* `name` - (Required) Name of the volume. This name is referenced in the `sourceVolume` parameter of container definition in the `mountPoints` section.

### `host` Block

* `source_path` - (Optional) Path on the host container instance that is presented to the container. If not set, ECS will create a non-persistent data volume that starts empty and is deleted after the task has finished.

### `depends_on` Block

* `condition` - (Required) Dependency condition of the container. Valid values: `START`, `COMPLETE`, `SUCCESS`, `HEALTHY`.
* `container_name` - (Required) Name of a container.

### `environment` Block

* `name` - (Required) Name of the environment variable.
* `value` - (Required) Value of the environment variable.

### `environment_file` Block

* `type` - (Required) File type to use. The only supported value is `s3`.
* `value` - (Required) ARN of the Amazon S3 object containing the environment variable file.

### `firelens_configuration` Block

* `options` - (Optional) Options to use when configuring the log router.
* `type` - (Required) Log router to use. Valid values: `fluentd`, `fluentbit`.

### `health_check` Block

* `command` - (Required) String array representing the command that the container runs to determine if it is healthy.
* `interval` - (Optional) Time period in seconds between each health check execution. Valid range: 5–300.
* `retries` - (Optional) Number of times to retry a failed health check. Valid range: 1–10.
* `start_period` - (Optional) Grace period in seconds to provide containers time to bootstrap. Valid range: 0–300.
* `timeout` - (Optional) Time period in seconds to wait for a health check to succeed. Valid range: 2–60.

### `linux_parameters` Block

* `capabilities` - (Optional) Linux capabilities for the container. Detailed below.
* `device` - (Optional) Any host devices to expose to the container. Detailed below.
* `init_process_enabled` - (Optional) Run an init process inside the container that forwards signals and reaps processes.
* `tmpfs` - (Optional) Container path, mount options, and size of the tmpfs mount. Detailed below.

### `capabilities` Block

* `add` - (Optional) Linux capabilities for the container that have been added to the default configuration provided by Docker.
* `drop` - (Optional) Linux capabilities for the container that have been removed from the default configuration provided by Docker.

### `device` Block

* `container_path` - (Optional) Path inside the container at which to expose the host device.
* `host_path` - (Required) Path for the device on the host container instance.
* `permissions` - (Optional) Explicit permissions to provide to the container for the device. Valid values: `read`, `write`, `mknod`.

### `tmpfs` Block

* `container_path` - (Required) Absolute file path where the tmpfs volume is to be mounted.
* `mount_options` - (Optional) List of tmpfs volume mount options.
* `size` - (Required) Maximum size (in MiB) of the tmpfs volume.

### `log_configuration` Block

* `log_driver` - (Required) Log driver to use for the container. Valid values: `json-file`, `syslog`, `journald`, `gelf`, `fluentd`, `awslogs`, `splunk`, `awsfirelens`.
* `options` - (Optional) Configuration options to send to the log driver.
* `secret_option` - (Optional) Secrets to pass to the log configuration. Detailed below.

### `secret_option` Block

* `name` - (Required) Name of the secret.
* `value_from` - (Required) Secret to expose to the log configuration.

### `mount_point` Block

* `container_path` - (Optional) Path on the container to mount the host volume at.
* `read_only` - (Optional) If this value is true, the container has read-only access to the volume.
* `source_volume` - (Optional) Name of the volume to mount.

### `repository_credentials` Block

* `credentials_parameter` - (Required) ARN of the secret containing the private repository credentials.

### `restart_policy` Block

* `enabled` - (Required) Whether a restart policy is enabled for the container.
* `ignored_exit_codes` - (Optional) List of exit codes that Amazon ECS will ignore and not attempt a restart on. Maximum of 50.
* `restart_attempt_period` - (Optional) Period of time (in seconds) that the container must run for before a restart can be attempted. Valid range: 60–1800.

### `secret` Block

* `name` - (Required) Name of the secret.
* `value_from` - (Required) Secret to expose to the container. The supported values are either the full ARN of the Secrets Manager secret or the full ARN of the parameter in the SSM Parameter Store.

### `system_control` Block

* `namespace` - (Optional) Namespaced kernel parameter to set a value for.
* `value` - (Optional) Value for the namespaced kernel parameter.

### `ulimit` Block

* `hard_limit` - (Required) Hard limit for the ulimit type.
* `name` - (Required) Type of the ulimit.
* `soft_limit` - (Required) Soft limit for the ulimit type.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Full ARN of the Daemon Task Definition (including both `family` and `revision`).
* `revision` - Revision of the task in a particular family.
* `status` - Status of the daemon task definition.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ecs_daemon_task_definition.example
  identity = {
    arn = "arn:aws:ecs:us-east-1:012345678910:daemon-task-definition/mydaemonfamily:123"
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
