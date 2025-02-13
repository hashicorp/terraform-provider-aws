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
  container_properties {
    command = ["ls","-la"],
    image   = "busybox"

    resource_requirements {
      type  = "VCPU"
      value = "0.25"
    }
    resource_requirements {
      type  = "MEMORY"
      value = "512"
    }

    volumes {
      name = "tmp"
      host {
        source_path = "/tmp"
      }
    }

    environment {
      name  = "VARNAME"
      value = "VARVAL"
    }

    mount_points {
      source_volume  = "tmp"
      container_path = "/tmp"
      read_only      = false
    }

    ulimits {
      hard_limit = 1024
      name       = "nofile"
      soft_limit = 1024
    }
  }
}
```

### Job definition of type multinode

```terraform
resource "aws_batch_job_definition" "test" {
  name = "tf_test_batch_job_definition_multinode"
  type = "multinode"

  node_properties {
    main_node = 0
    num_nodes = 2
    node_range_properties {
      container {
        command = ["ls", "-la"]
        image   = "busybox"
        memory  = 128
        vcpus   = 1
      }
      target_nodes = "0:"
    }
    node_range_properties {
      container {
        command = ["echo", "test"]
        image   = "busybox"
        memory  = 128
        vcpus   = 1
      }
      target_nodes = "1:"
    }
  }
}
```

### Job Definition of type EKS

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

  container_properties {
    command      = ["echo", "test"]
    image        = "busybox"
    job_role_arn = "arn:aws:iam::123456789012:role/AWSBatchS3ReadOnly"

    fargate_platform_configuration = {
      platform_version = "LATEST"
    }

    resource_requirements {
      type  = "VCPU"
      value = "0.25"
    }
    resource_requirements {
      type  = "MEMORY"
      value = "512"
    }

    execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  }
}
```

### Job definition of type container using `ecs_properties`

```terraform
resource "aws_batch_job_definition" "test" {
  name = "tf_test_batch_job_definition"
  type = "container"

  platform_capabilities = ["FARGATE"]

  ecs_properties {
    task_properties {
      execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
      containers {
        image   = "public.ecr.aws/amazonlinux/amazonlinux:1"
        command = ["sleep", "60"]
        depends_on {
          container_name = "container_b"
          condition      = "COMPLETE"
        }

        secrets {
          name       = "TEST"
          value_from = "DUMMY"
        }

        environment {
          name  = "test 1"
          value = "Environment Variable 1"
        }

        essential = true
        log_configuration {
          log_driver = "awslogs"
          options = {
            "awslogs-group"         = %[1]q
            "awslogs-region"        = %[2]q
            "awslogs-stream-prefix" = "ecs"
          }
        }
        name                     = "container_a"
        privileged               = false
        readonly_root_filesystem = false
        resource_requirements {
          value = "1.0"
          type  = "VCPU"
        }
        resource_requirements {
          value = "2048"
          type  = "MEMORY"
        }
      }
      containers {
        image     = "public.ecr.aws/amazonlinux/amazonlinux:1"
        command   = ["sleep", "360"]
        name      = "container_b"
        essential = false
        resource_requirements {
          value = "1.0"
          type  = "VCPU"
        }
        resource_requirements {
          value = "2048"
          type  = "MEMORY"
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the job definition.
* `type` - (Required) Type of job definition. Must be `container` or `multinode`.

The following arguments are optional:

* `container_properties` - (Optional) Valid [container properties](#container-properties). This parameter is only valid if the `type` parameter is `container`.
* `deregister_on_new_revision` - (Optional) When updating a job definition a new revision is created. This parameter determines if the previous version is `deregistered` (`INACTIVE`) or left  `ACTIVE`. Defaults to `true`.
* `ecs_properties` - (Optional) Valid [ECS properties](#ecs-properties). This parameter is only valid if the `type` parameter is `container`.
* `eks_properties` - (Optional) Valid [eks properties](#eks_properties). This parameter is only valid if the `type` parameter is `container`.
* `node_properties` - (Optional) Valid [node properties](#node-properties). This parameter is required if the `type` parameter is `multinode`.
* `parameters` - (Optional) Parameter substitution placeholders to set in the job definition.
* `platform_capabilities` - (Optional) Platform capabilities required by the job definition. If no value is specified, it defaults to `EC2`. To run the job on Fargate resources, specify `FARGATE`.
* `propagate_tags` - (Optional) Whether to propagate the tags from the job definition to the corresponding Amazon ECS task. Default is `false`.
* `retry_strategy` - (Optional) Retry strategy to use for failed jobs that are submitted with this job definition. Maximum number of `retry_strategy` is `1`.  Defined below.
* `scheduling_priority` - (Optional) Scheduling priority of the job definition. This only affects jobs in job queues with a fair share policy. Jobs with a higher scheduling priority are scheduled before jobs with a lower scheduling priority. Allowed values `0` through `9999`.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `timeout` - (Optional) Timeout for jobs so that if a job runs longer, AWS Batch terminates the job. Maximum number of `timeout` is `1`. Defined below.

### `eks_properties`

* `pod_properties` - (Optional) Properties for the Kubernetes pod resources of a job. See [`pod_properties`](#pod_properties) below.

#### `pod_properties`

* `containers` - (Optional) Properties of the container that's used on the Amazon EKS pod. See [containers](#eks_containers) below.
* `dns_policy` - (Optional) DNS policy for the pod. The default value is `ClusterFirst`. If the `host_network` argument is not specified, the default is `ClusterFirstWithHostNet`. `ClusterFirst` indicates that any DNS query that does not match the configured cluster domain suffix is forwarded to the upstream nameserver inherited from the node. For more information, see Pod's DNS policy in the Kubernetes documentation.
* `host_network` - (Optional) Whether the pod uses the hosts' network IP address. The default value is `true`. Setting this to `false` enables the Kubernetes pod networking model. Most AWS Batch workloads are egress-only and don't require the overhead of IP allocation for each pod for incoming connections.
* `init_containers` - (Optional) Containers which run before application containers, always runs to completion, and must complete successfully before the next container starts. These containers are registered with the Amazon EKS Connector agent and persists the registration information in the Kubernetes backend data store. See [containers](#container) below.
* `image_pull_secret` - (Optional) List of Kubernetes secret resources. See [`image_pull_secret`](#image_pull_secret) below.
* `metadata` - (Optional) Metadata about the Kubernetes pod.
* `service_account_name` - (Optional) Name of the service account that's used to run the pod.
* `share_process_namespace` - (Optional) Indicates if the processes in a container are shared, or visible, to other containers in the same pod.
* `metadata` - [Metadata](#eks_metadata) about the Kubernetes pod.
* `volumes` - (Optional) Volumes for a job definition that uses Amazon EKS resources. AWS Batch supports [emptyDir](#eks_empty_dir), [hostPath](#eks_host_path), and [secret](#eks_secret) volume types.

#### `eks_containers`

* `args` - (Optional) Array of arguments to the entrypoint. If this isn't specified, the CMD of the container image is used. This corresponds to the args member in the Entrypoint portion of the Pod in Kubernetes. Environment variable references are expanded using the container's environment.
* `command` - (Optional) Entrypoint for the container. This isn't run within a shell. If this isn't specified, the ENTRYPOINT of the container image is used. Environment variable references are expanded using the container's environment.
* `env` - (Optional) Environment variables to pass to a container. See [EKS Environment](#eks_environment) below.
* `image` - (Required) Docker image used to start the container.
* `image_pull_policy` - (Optional) Image pull policy for the container. Supported values are `Always`, `IfNotPresent`, and `Never`.
* `name` - (Optional) Name of the container. If the name isn't specified, the default name "Default" is used. Each container in a pod must have a unique name.
* `resources` - (Optional) Type and amount of resources to assign to a container. The supported resources include `memory`, `cpu`, and `nvidia.com/gpu`.
* `security_context` - (Optional) Security context for a job.
* `volume_mounts` - (Optional) Volume mounts for the container.

#### `image_pull_secret`

* `name` - (Required) Unique identifier.

#### `eks_environment`

* `name` - (Required) Name of the environment variable.
* `value` - (Optional) Value of the environment variable.

#### `eks_empty_dir`

* `medium` - (Optional) Medium to store the volume. The default value is an empty string, which uses the storage of the node.
* `size_limit` - (Optional) Maximum size of the volume. By default, there's no maximum size defined.

#### `eks_host_path`

* `path` - (Optional) Path of the file or directory on the host to mount into containers on the pod.

#### `eks_metadata`

* `labels` - Key-value pairs used to identify, sort, and organize cube resources.

#### `eks_secret`

* `secret_name` - (Required) Name of the secret. The name must be allowed as a DNS subdomain name.
* `optional` - (Optional) Whether the secret or the secret's keys must be defined.

### `retry_strategy`

* `attempts` - (Optional) Number of times to move a job to the `RUNNABLE` status. You may specify between `1` and `10` attempts.
* `evaluate_on_exit` - (Optional) [Evaluate on exit](#evaluate_on_exit) conditions under which the job should be retried or failed. If this parameter is specified, then the `attempts` parameter must also be specified. You may specify up to 5 configuration blocks.

#### `evaluate_on_exit`

* `action` - (Required) Action to take if all of the specified conditions are met. The values are not case sensitive. Valid values: `retry`, `exit`.
* `on_exit_code` - (Optional) Glob pattern to match against the decimal representation of the exit code returned for a job.
* `on_reason` - (Optional) Glob pattern to match against the reason returned for a job.
* `on_status_reason` - (Optional) Glob pattern to match against the status reason returned for a job.

### `timeout`

* `attempt_duration_seconds` - (Optional) Time duration in seconds after which AWS Batch terminates your jobs if they have not finished. The minimum value for the timeout is `60` seconds

### `container`

- `command` - The command that's passed to the container.
- `environment` - The [environment](#environment) variables to pass to a container.
- `ephemeral_storage` - The amount of [ephemeral storage](#ephemeral_storage) to allocate for the task. This parameter is used to expand the total amount of ephemeral storage available, beyond the default amount, for tasks hosted on AWS Fargate.
- `execution_role_arn` - The Amazon Resource Name (ARN) of the execution role that AWS Batch can assume. For jobs that run on Fargate resources, you must provide an execution role.
- `fargate_platform_configuration` - The [platform configuration](#fargate_platform_configuration) for jobs that are running on Fargate resources. Jobs that are running on EC2 resources must not specify this parameter.
- `image` - The image used to start a container.
- `instance_type` - The instance type to use for a multi-node parallel job.
- `job_role_arn` - The Amazon Resource Name (ARN) of the IAM role that the container can assume for AWS permissions.
- `linux_parameters` - [Linux-specific modifications](#linux_parameters) that are applied to the container.
- `log_configuration` - The [log configuration](#log_configuration) specification for the container.
- `mount_points` - The [mount points](#mount_points) for data volumes in your container.
- `network_configuration` - The [network configuration](#network_configuration) for jobs that are running on Fargate resources.
- `privileged` - When this parameter is true, the container is given elevated permissions on the host container instance (similar to the root user).
- `readonly_root_filesystem` - When this parameter is true, the container is given read-only access to its root file system.
- `resource_requirements` - The type and amount of [resources](#resource_requirements) to assign to a container.
- `runtime_platform` - An [object](#runtime_platform) that represents the compute environment architecture for AWS Batch jobs on Fargate.
- `secrets` - The [secrets](#secrets) for the container.
- `ulimits` - A list of [ulimits](#ulimits) to set in the container.
- `user` - The user name to use inside the container.
- `volumes` - A list of data [volumes](#volumes) used in a job.

### `node_properties`

- `main_node` - Specifies the node index for the main node of a multi-node parallel job. This node index value must be fewer than the number of nodes.
- `node_range_properties` - A list of node ranges and their [properties](#node_range_properties) that are associated with a multi-node parallel job.
- `num_nodes` - The number of nodes that are associated with a multi-node parallel job.

#### `node_range_properties`

- `target_nodes` - The range of nodes, using node index values. A range of 0:3 indicates nodes with index values of 0 through 3. I
- `container` - The [container details](#container) for the node range.

#### `environment`

- `name` - The name of the key-value pair.
- `value` - The value of the key-value pair.

#### `ephemeral_storage`

- `size_in_gb` - The total amount, in GiB, of ephemeral storage to set for the task.

#### `fargate_platform_configuration`

- `platform_version` - The AWS Fargate platform version where the jobs are running. A platform version is specified only for jobs that are running on Fargate resources.

#### `linux_parameters`

- `init_process_enabled` - If true, run an init process inside the container that forwards signals and reaps processes.
- `max_swap` - The total amount of swap memory (in MiB) a container can use.
- `shared_memory_size` - The value for the size (in MiB) of the `/dev/shm` volume.
- `swappiness` - You can use this parameter to tune a container's memory swappiness behavior.
- `devices` - Any of the [host devices](#devices) to expose to the container.
- `tmpfs` - The container path, mount options, and size (in MiB) of the [tmpfs](#tmpfs) mount.

#### `log_configuration`

- `options` - The configuration options to send to the log driver.
- `log_driver` - The log driver to use for the container.
- `secret_options` - The secrets to pass to the log configuration.

#### `network_configuration`

- `assign_public_ip` - Indicates whether the job has a public IP address.

#### `mount_points`

- `container_path` - The path on the container where the host volume is mounted.
- `read_only` - If this value is true, the container has read-only access to the volume.
- `source_volume` - The name of the volume to mount.

#### `resource_requirements`

- `type` - The type of resource to assign to a container. The supported resources include `GPU`, `MEMORY`, and `VCPU`.
- `value` - The quantity of the specified resource to reserve for the container.

#### `secrets`

- `name` - The name of the secret.
- `value_from` - The secret to expose to the container.

#### `ulimits`

- `hard_limit` - The hard limit for the ulimit type.
- `name` - The type of the ulimit.
- `soft_limit` - The soft limit for the ulimit type.

#### `runtime_platform`

- `cpu_architecture` - The vCPU architecture. The default value is X86_64. Valid values are X86_64 and ARM64.
- `operating_system_family` - The operating system for the compute environment. V

#### `secret_options`

- `name` - The name of the secret.
- `value_from` - The secret to expose to the container. The supported values are either the full Amazon Resource Name (ARN) of the AWS Secrets Manager secret or the full ARN of the parameter in the AWS Systems Manager Parameter Store.

#### `devices`

- `host_path` - The path for the device on the host container instance.
- `container_path` - The path inside the container that's used to expose the host device. By default, the hostPath value is used.
- `permissions` - The explicit permissions to provide to the container for the device.

#### `tmpfs`

- `container_path` - The absolute file path in the container where the tmpfs volume is mounted.
- `size` - The size (in MiB) of the tmpfs volume.
- `mount_options` - The list of tmpfs volume mount options.

#### `volumes`

- `name` - The name of the volume.
- `host` - The contents of the host parameter determine whether your data volume persists on the host container instance and where it's stored.
- `efs_volume_configuration` - This [parameter](#efs_volume_configuration) is specified when you're using an Amazon Elastic File System file system for job storage.

#### `host`

- `source_path` - The path on the host container instance that's presented to the container.

#### `efs_volume_configuration`

- `file_system_id` - The Amazon EFS file system ID to use.
- `root_directory` - The directory within the Amazon EFS file system to mount as the root directory inside the host.
- `transit_encryption` - Determines whether to enable encryption for Amazon EFS data in transit between the Amazon ECS host and the Amazon EFS server
- `transit_encryption_port` - The port to use when sending encrypted data between the Amazon ECS host and the Amazon EFS server.
- `authorization_config` - The [authorization configuration](#authorization_config) details for the Amazon EFS file system.

#### `authorization_config`

- `access_point_id` - The Amazon EFS access point ID to use.
- `iam` - Whether or not to use the AWS Batch job IAM role defined in a job definition when mounting the Amazon EFS file system.

#### `retry_strategy`

- `attempts` - The number of times to move a job to the RUNNABLE status.
- `evaluate_on_exit` - Array of up to 5 [objects](#evaluate_on_exit) that specify the conditions where jobs are retried or failed.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the job definition, includes revision (`:#`).
* `arn_prefix` - ARN without the revision number.
* `revision` - Revision of the job definition.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
