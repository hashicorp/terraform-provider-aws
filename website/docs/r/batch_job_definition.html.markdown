---
subcategory: "Batch"
layout: "aws"
page_title: "AWS: aws_batch_job_definition"
description: |-
  Provides a Batch Job Definition resource.
---

# Resource: aws_batch_job_definition

Provides a Batch Job Definition resource.

## Example Usage

### Job definition of type container

```terraform
resource "aws_batch_job_definition" "test" {
  name = "tf_test_batch_job_definition"
  type = "container"
  container_properties = jsonencode({
    command = ["ls", "-la"],
    image   = "busybox"

    resourceRequirements = [
      {
        type  = "VCPU"
        value = "0.25"
      },
      {
        type  = "MEMORY"
        value = "512"
      }
    ]

    volumes = [
      {
        host = {
          sourcePath = "/tmp"
        }
        name = "tmp"
      }
    ]

    environment = [
      {
        name  = "VARNAME"
        value = "VARVAL"
      }
    ]

    mountPoints = [
      {
        sourceVolume  = "tmp"
        containerPath = "/tmp"
        readOnly      = false
      }
    ]

    ulimits = [
      {
        hardLimit = 1024
        name      = "nofile"
        softLimit = 1024
      }
    ]
  })
}
```

### Job definition of type multinode

```terraform
resource "aws_batch_job_definition" "test" {
  name = "tf_test_batch_job_definition_multinode"
  type = "multinode"

  node_properties = jsonencode({
    mainNode = 0
    nodeRangeProperties = [
      {
        container = {
          command = ["ls", "-la"]
          image   = "busybox"
          memory  = 128
          vcpus   = 1
        }
        targetNodes = "0:"
      },
      {
        container = {
          command = ["echo", "test"]
          image   = "busybox"
          memory  = 128
          vcpus   = 1
        }
        targetNodes = "1:"
      }
    ]
    numNodes = 2
  })
}
```

### Job Definitionn of type EKS

```terraform
resource "aws_batch_job_definition" "test" {
  name = " tf_test_batch_job_definition_eks"
  type = "container"
  eks_properties {
    pod_properties {
      host_network = true
      containers {
        image = "public.ecr.aws/amazonlinux/amazonlinux:1"
        command = [
          "sleep",
          "60"
        ]
        resources {
          limits = {
            cpu    = "1"
            memory = "1024Mi"
          }
        }
      }
      metadata {
        labels = {
          environment = "test"
        }
      }
    }
  }
}
```

### Job Definitionn of type EKS

```terraform
resource "aws_batch_job_definition" "test" {
  name = " tf_test_batch_job_definition_ecs"
  type = "container"
  ecs_properties {
    task_properties {
      host_network = true
      containers {
        essential = true
        image = "public.ecr.aws/amazonlinux/amazonlinux:1"
        command = [
          "sleep",
          "60"
        ]
        resource_requirements {
          value = "1.0"
          type = "VCPU"
        }
        resource_requirements {
          value = "2048"
          type = "MEMORY"
        }
        environment {
		  name = "test"
		  value = "Environment Variable"
        }
      }
    }
  }
}
```

### Fargate Platform Capability

```terraform
resource "aws_iam_role" "ecs_task_execution_role" {
  name               = "tf_test_batch_exec_role"
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_batch_job_definition" "test" {
  name = "tf_test_batch_job_definition"
  type = "container"

  platform_capabilities = [
    "FARGATE",
  ]

  container_properties = jsonencode({
    command    = ["echo", "test"]
    image      = "busybox"
    jobRoleArn = "arn:aws:iam::123456789012:role/AWSBatchS3ReadOnly"

    fargatePlatformConfiguration = {
      platformVersion = "LATEST"
    }

    resourceRequirements = [
      {
        type  = "VCPU"
        value = "0.25"
      },
      {
        type  = "MEMORY"
        value = "512"
      }
    ]

    executionRoleArn = aws_iam_role.ecs_task_execution_role.arn
  })
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Specifies the name of the job definition.
* `type` - (Required) The type of job definition. Must be `container` or `multinode`.

The following arguments are optional:

* `container_properties` - (Optional) A valid [container properties](http://docs.aws.amazon.com/batch/latest/APIReference/API_RegisterJobDefinition.html) provided as a single valid JSON document. This parameter is only valid if the `type` parameter is `container`.
* `deregister_on_new_revision` - (Optional) When updating a job definition a new revision is created. This parameter determines if the previous version is `deregistered` (`INACTIVE`) or left  `ACTIVE`. Defaults to `true`.
* `node_properties` - (Optional) A valid [node properties](http://docs.aws.amazon.com/batch/latest/APIReference/API_RegisterJobDefinition.html) provided as a single valid JSON document. This parameter is required if the `type` parameter is `multinode`.
* `eks_properties` - (Optional) A valid [eks properties](#eks_properties). This parameter is only valid if the `type` parameter is `container`.
* `parameters` - (Optional) Specifies the parameter substitution placeholders to set in the job definition.
* `platform_capabilities` - (Optional) The platform capabilities required by the job definition. If no value is specified, it defaults to `EC2`. To run the job on Fargate resources, specify `FARGATE`.
* `propagate_tags` - (Optional) Specifies whether to propagate the tags from the job definition to the corresponding Amazon ECS task. Default is `false`.
* `retry_strategy` - (Optional) Specifies the retry strategy to use for failed jobs that are submitted with this job definition. Maximum number of `retry_strategy` is `1`.  Defined below.
* `scheduling_priority` - (Optional) The scheduling priority of the job definition. This only affects jobs in job queues with a fair share policy. Jobs with a higher scheduling priority are scheduled before jobs with a lower scheduling priority. Allowed values `0` through `9999`.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `timeout` - (Optional) Specifies the timeout for jobs so that if a job runs longer, AWS Batch terminates the job. Maximum number of `timeout` is `1`. Defined below.

### `eks_properties`

* `pod_properties` - The properties for the Kubernetes pod resources of a job. See [`pod_properties`](#pod_properties) below.

### `pod_properties`

* `containers` - The properties of the container that's used on the Amazon EKS pod. See [containers](#containers) below.
* `dns_policy` - (Optional) The DNS policy for the pod. The default value is `ClusterFirst`. If the `host_network` argument is not specified, the default is `ClusterFirstWithHostNet`. `ClusterFirst` indicates that any DNS query that does not match the configured cluster domain suffix is forwarded to the upstream nameserver inherited from the node. For more information, see Pod's DNS policy in the Kubernetes documentation.
* `host_network` - (Optional) Indicates if the pod uses the hosts' network IP address. The default value is `true`. Setting this to `false` enables the Kubernetes pod networking model. Most AWS Batch workloads are egress-only and don't require the overhead of IP allocation for each pod for incoming connections.
* `metadata` - (Optional) Metadata about the Kubernetes pod.
* `service_account_name` - (Optional) The name of the service account that's used to run the pod.
* `volumes` - (Optional) Specifies the volumes for a job definition that uses Amazon EKS resources. AWS Batch supports [emptyDir](#eks_empty_dir), [hostPath](#eks_host_path), and [secret](#eks_secret) volume types.

### `containers`

* `image` - The Docker image used to start the container.
* `args` - An array of arguments to the entrypoint. If this isn't specified, the CMD of the container image is used. This corresponds to the args member in the Entrypoint portion of the Pod in Kubernetes. Environment variable references are expanded using the container's environment.
* `command` - The entrypoint for the container. This isn't run within a shell. If this isn't specified, the ENTRYPOINT of the container image is used. Environment variable references are expanded using the container's environment.
* `env` - The environment variables to pass to a container. See [EKS Environment](#eks_environment) below.
* `image_pull_policy` - The image pull policy for the container. Supported values are `Always`, `IfNotPresent`, and `Never`.
* `name` - The name of the container. If the name isn't specified, the default name "Default" is used. Each container in a pod must have a unique name.
* `resources` - The type and amount of resources to assign to a container. The supported resources include `memory`, `cpu`, and `nvidia.com/gpu`.
* `security_context` - The security context for a job.
* `volume_mounts` - The volume mounts for the container.

### `eks_environment`

* `name` - The name of the environment variable.
* `value` - The value of the environment variable.

### `eks_empty_dir`

* `medium` - (Optional) The medium to store the volume. The default value is an empty string, which uses the storage of the node.
* `size_limit` - The maximum size of the volume. By default, there's no maximum size defined.

### `eks_host_path`

* `path` - The path of the file or directory on the host to mount into containers on the pod.

### `eks_secret`

* `secret_name` - The name of the secret. The name must be allowed as a DNS subdomain name.
* `optional` - (Optional) Specifies whether the secret or the secret's keys must be defined.

### `ecs_properties`

* `task_properties` - The properties of the Amazon ECS task definition on a job. See [`task_properties`](#task_properties) below.

### `task_properties`

* `containers` - The properties of the container. See [containers](#ecs_containers) below.
* `ephemeral_storage` - (Optional) The ephemeral storage settings for the task. See [ephemeral_storage](#ephemeral_storage) below.
* `execution_role_arn` - (Optional) The Amazon Resource Name (ARN) of the IAM role that the AWS Batch can assume.
* `ipc_mode` - (Optional) The IPC resource namespace to use for container in the task. The valid values are `host`, `task`, or `none`.
* `network_configuration` - (Optional) The network configuration for the Amazon ECS task. See [network_configuration](#network_configuration) below.
* `pic_mode` - (Optional) The process namespace to use for the container in the task. The valid values are `host` or `task`.
* `platform_version` - (Optional) The platform version on which to run your task. If one is not specified, the LATEST platform version is used by default.
* `runtime_platform` - (Optional) The compute environment architecture for AWS Batch job on Fargate. See [runtime_platform](#runtime_platform) below.
* `task_role_arn` - The Amazon Resource Name (ARN) of the IAM role that the Amazon ECS task..
* `volume` - (Optional) The volumes to mount on the container in the task. See [volume](#volume) below.

### `containers`

* `image` - The image used to start the container.
* `command` - (Optional) The command that's passed to the container.
* `depends_on` - (Optional) The properties of containers that this container depends on. See [depends_on](#depends_on) below.
* `environment` - (Optional) The environment variables to pass to a container. See [environment](#environment) below.
* `essential` - (Optional) The parameter this container is essential or not.
* `linux_parameters` - (Optional) Linux-specific modifications that are applied to the container. See [linux_parameters](#linux_parameters) below.
* `log_configuration` - (Optional) The log configuration specification for the container. See [log_configuration](#log_configuration) below.
* `mount_points` - (Optional) The mount points for data volumes in your container. See [mount_points](#mount_points) below.
* `name` - (Optional) The name of the container.
* `privileged` - (Optional) When this parameter is true, the container is given elevated privileges on the host container instance.
* `readonly_root_filesystem` - (Optional) When this parameter is true, the container is given read-only access to its root file system.
* `repository_credentials` - (Optional) The private repository authentication credentials to use. See [repository_credentials](#repository_credentials) below.
* `resource_requirements` - (Optional) The type and amount of resources to assign to a container. The supported resources include `GPU`, `MEMORY`, and `VCPU`. See [resource_requirements](#resource_requirements) below.
* `secrets` - (Optional) The secrets to pass to the container. See [secrets](#secrets) below.
* `ulimits` - (Optional) A list of `ulimits` to set in the container. This parameter maps to `Ulimits` in the Create a container section of the Docker Remote API and the --ulimit option to docker run. See [ulimits](#ulimits) below.
* `user` - (Optional) The user to use inside the container. This parameter maps to User in the Create a container section of the Docker Remote API and the --user option to docker run.

### `depends_on`

* `container_name` - The name of a container in the task definition to depend on.
* `condition` - The dependency condition of the dependent container. The valid values are `START`, `COMPLETE`, `SUCCESS`, and `HEALTHY`.

### `environment`

* `name` - The name of the environment variable.
* `value` - The value of the environment variable.

### `linux_parameters`

* `devices` - (Optional) Any host devices to expose to the container. This parameter maps to Devices in the Create a container section of the Docker Remote API and the --device option to docker run. See [devices](#devices) below.
* `init_process_enabled` - (Optional) Whether the init process is enabled in the container. This parameter maps to `init` in the Create a container section of the Docker Remote API and the --init option to docker run.
* `max_swap` - (Optional) The total amount of swap memory (in MiB) a container can use. This parameter will be translated to the --memory-swap option to docker run where the value would be the sum of the container memory plus the maxSwap value.
* `shared_memory_size` - (Optional) The value for the size (in MiB) of the `/dev/shm` volume. This parameter maps to `ShmSize` in the Create a container section of the Docker Remote API and the --shm-size option to docker run.
* `swappiness` - (Optional) This allows you to tune a container's memory swappiness behavior. A swappiness value is a percentage from `0` to `100`.
* `tmpfs` - (Optional) The container path, mount options, and size (in MiB) of a tmpfs mount. This parameter maps to `Tmpfs` in the Create a container section of the Docker Remote API and the --tmpfs option to docker run. See [tmpfs](#tmpfs) below.

### `devices`

* `host_path` - The path for the device on the host.
* `container_path` - The path inside the container at which to expose the host device.
* `permissions` - The explicit permissions to provide to the container for the device. By default, the container has permissions for `read`, `write`, and `mknod` for the device.

### `tmpfs`

* `container_path` - The absolute file path in the container where the tmpfs volume is mounted.
* `size` - The size (in MiB) of the tmpfs volume.
* `mount_options` - (Optional) The list of tmpfs volume mount options.

### `log_configuration`

* `log_driver` - The log driver to use for the container. The valid values listed for this parameter are log drivers that the Amazon ECS container agent can communicate with by default.
* `options` - The configuration options to send to the log driver. This parameter requires a map of key-value pairs.
* `secret_options` - The secrets to pass to the log configuration. See [secret_options](#secret_options) below.

### `secret_options`

* `name` - The name of the secret.
* `value_from` - The value to assign to the secret.

### `mount_points`

* `container_path` - The path in the container at which to mount the host volume.
* `read_only` - If this value is `true`, the container has read-only access to the volume. If this value is `false`, then the container can write to the volume. The default value is `false`.
* `source_volume` - The name of the volume to mount.

### `repository_credentials`

* `credentials_parameter` - The Amazon Resource Name (ARN) or name of the secret in Secrets Manager that stores the private repository authentication credentials.

### `resource_requirements`

* `type` - The type of resource to assign to a container. The supported resources include `GPU`, `MEMORY`, and `VCPU`.
* `value` - The value for the specified resource type.

### `secrets`

* `name` - The name of the secret.
* `value_from` - The value to assign to the secret.

### `ulimits`

* `hard_limit` - The hard limit for the ulimit type.
* `name` - The type of the ulimit. For valid values, see the [ulimit](https://docs.docker.com/engine/reference/run/#ulimit) documentation.
* `soft_limit` - The soft limit for the ulimit type.

### `ephemeral_storage`

* `size_in_gib` - The size (in GiB) of the ephemeral storage volume.

### `network_configuration`

* `assign_public_ip` - (Optional) Assign a public IP address to the Amazon ECS task. The default value is `DISABLED`.

### `runtime_platform`

* `cpu_architecture` - The CPU architecture to use for the task. The valid values are `X86_64` and `ARM64`.
* `operating_system_family` - The operating system family to use for the task. The valid values are `LINUX` and `WINDOWS`.

### `volume`

* `efs_volume_configuration` - (Optional) The Amazon Elastic File System (Amazon EFS) volume configuration to use for the Amazon ECS task. See [efs_volume_configuration](#efs_volume_configuration) below.
* `host` - (Optional) The host volume configuration to use for the Amazon ECS task. See [host](#host) below.
* `name` - The name of the volume.

### `efs_volume_configuration`

* `file_system_id` - The Amazon EFS file system ID to use.
* `authorization_config` - (Optional) The authorization configuration details for the Amazon EFS file system. See [authorization_config](#authorization_config) below.
* `root_directory` - (Optional) The root directory to mount to the Amazon ECS task.
* `transit_encryption` - (Optional) Whether to enable encryption for Amazon EFS data in transit between the Amazon ECS host and the Amazon EFS server. The default value is `ENABLED`.
* `transit_encryption_port` - (Optional) The port to use when sending encrypted data between the Amazon ECS host and the Amazon EFS server. The default value is `PORT_443`.

### `authorization_config`

* `access_point_id` - The Amazon EFS access point ID to use.
* `iam` - The Amazon EFS authorization method to use.

### `host`

* `source_path` - The path on the host container instance that's presented to the container. If this parameter is empty, then the Docker daemon has assigned a host path for you.

### `retry_strategy`

* `attempts` - (Optional) The number of times to move a job to the `RUNNABLE` status. You may specify between `1` and `10` attempts.
* `evaluate_on_exit` - (Optional) The [evaluate on exit](#evaluate_on_exit) conditions under which the job should be retried or failed. If this parameter is specified, then the `attempts` parameter must also be specified. You may specify up to 5 configuration blocks.

#### `evaluate_on_exit`

* `action` - (Required) Specifies the action to take if all of the specified conditions are met. The values are not case sensitive. Valid values: `retry`, `exit`.
* `on_exit_code` - (Optional) A glob pattern to match against the decimal representation of the exit code returned for a job.
* `on_reason` - (Optional) A glob pattern to match against the reason returned for a job.
* `on_status_reason` - (Optional) A glob pattern to match against the status reason returned for a job.

### `timeout`

* `attempt_duration_seconds` - (Optional) The time duration in seconds after which AWS Batch terminates your jobs if they have not finished. The minimum value for the timeout is `60` seconds.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name of the job definition, includes revision (`:#`).
* `arn_prefix` - The ARN without the revision number.
* `revision` - The revision of the job definition.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Batch Job Definition using the `arn`. For example:

```terraform
import {
  to = aws_batch_job_definition.test
  id = "arn:aws:batch:us-east-1:123456789012:job-definition/sample"
}
```

Using `terraform import`, import Batch Job Definition using the `arn`. For example:

```console
% terraform import aws_batch_job_definition.test arn:aws:batch:us-east-1:123456789012:job-definition/sample
```
