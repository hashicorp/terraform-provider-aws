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
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `task_role_arn` - (Optional) ARN of IAM role that allows your Amazon ECS container task to make calls to other AWS services.
* `volume` - (Optional) Repeatable configuration block for [volumes](#volume) that containers in your task may use. Detailed below.

### container_definition

* `image` - (Required) The image used to start a container.
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
* `name` - (Optional) The name of a container.
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
* `name` - (Optional) Name of the volume. This name is referenced in the `sourceVolume` parameter of container definition in the `mountPoints` section.

### depends_on

* `condition` - (Required) The dependency condition of the container. Valid values: `START`, `COMPLETE`, `SUCCESS`, `HEALTHY`.
* `container_name` - (Required) The name of a container.

### environment

* `name` - (Required) The name of the environment variable.
* `value` - (Required) The value of the environment variable.

### environment_file

* `type` - (Required) The file type to use. The only supported value is `s3`.
* `value` - (Required) The ARN of the Amazon S3 object containing the environment variable file.

### firelens_configuration

* `options` - (Optional) The options to use when configuring the log router.
* `type` - (Required) The log router to use. Valid values: `fluentd`, `fluentbit`.

### health_check

* `command` - (Required) A string array representing the command that the container runs to determine if it is healthy.
* `interval` - (Optional) The time period in seconds between each health check execution. Valid range: 5–300.
* `retries` - (Optional) The number of times to retry a failed health check. Valid range: 1–10.
* `start_period` - (Optional) The grace period in seconds to provide containers time to bootstrap. Valid range: 0–300.
* `timeout` - (Optional) The time period in seconds to wait for a health check to succeed. Valid range: 2–60.

### linux_parameters

* `capabilities` - (Optional) The Linux capabilities for the container. Detailed below.
* `device` - (Optional) Any host devices to expose to the container. Detailed below.
* `init_process_enabled` - (Optional) Run an init process inside the container that forwards signals and reaps processes.
* `tmpfs` - (Optional) The container path, mount options, and size of the tmpfs mount. Detailed below.

### capabilities

* `add` - (Optional) The Linux capabilities for the container that have been added to the default configuration provided by Docker.
* `drop` - (Optional) The Linux capabilities for the container that have been removed from the default configuration provided by Docker.

### device

* `container_path` - (Optional) The path inside the container at which to expose the host device.
* `host_path` - (Required) The path for the device on the host container instance.
* `permissions` - (Optional) The explicit permissions to provide to the container for the device. Valid values: `read`, `write`, `mknod`.

### tmpfs

* `container_path` - (Required) The absolute file path where the tmpfs volume is to be mounted.
* `mount_options` - (Optional) The list of tmpfs volume mount options.
* `size` - (Required) The maximum size (in MiB) of the tmpfs volume.

### log_configuration

* `log_driver` - (Required) The log driver to use for the container. Valid values: `json-file`, `syslog`, `journald`, `gelf`, `fluentd`, `awslogs`, `splunk`, `awsfirelens`.
* `options` - (Optional) The configuration options to send to the log driver.
* `secret_option` - (Optional) The secrets to pass to the log configuration. Detailed below.

### secret_option

* `name` - (Required) The name of the secret.
* `value_from` - (Required) The secret to expose to the log configuration.

### mount_point

* `container_path` - (Optional) The path on the container to mount the host volume at.
* `read_only` - (Optional) If this value is true, the container has read-only access to the volume.
* `source_volume` - (Optional) The name of the volume to mount.

### repository_credentials

* `credentials_parameter` - (Required) The ARN of the secret containing the private repository credentials.

### restart_policy

* `enabled` - (Required) Specifies whether a restart policy is enabled for the container.
* `ignored_exit_codes` - (Optional) A list of exit codes that Amazon ECS will ignore and not attempt a restart on. Maximum of 50.
* `restart_attempt_period` - (Optional) A period of time (in seconds) that the container must run for before a restart can be attempted. Valid range: 60–1800.

### secret

* `name` - (Required) The name of the secret.
* `value_from` - (Required) The secret to expose to the container. The supported values are either the full ARN of the Secrets Manager secret or the full ARN of the parameter in the SSM Parameter Store.

### system_control

* `namespace` - (Optional) The namespaced kernel parameter to set a value for.
* `value` - (Optional) The value for the namespaced kernel parameter.

### ulimit

* `hard_limit` - (Required) The hard limit for the ulimit type.
* `name` - (Required) The type of the ulimit.
* `soft_limit` - (Required) The soft limit for the ulimit type.

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
