---
subcategory: "Batch"
layout: "aws"
page_title: "AWS: aws_batch_job_definition"
description: |-
  Terraform data source for managing an AWS Batch Job Definition.
---

# Data Source: aws_batch_job_definition

Terraform data source for managing an AWS Batch Job Definition.

## Example Usage

### Lookup via Arn

```terraform
data "aws_batch_job_definition" "arn" {
  arn = "arn:aws:batch:us-east-1:012345678910:job-definition/example"
}
```

### Lookup via Name

```terraform
data "aws_batch_job_definition" "name" {
  name     = "example"
  revision = 2
}
```

## Argument Reference

The following arguments are optional:

* `arn` - ARN of the Job Definition. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `revision` - The revision of the job definition.
* `name` - The name of the job definition to register. It can be up to 128 letters long. It can contain uppercase and lowercase letters, numbers, hyphens (-), and underscores (_).
* `status` - The status of the job definition.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `container_orchestration_type` - The orchestration type of the compute environment.
* `scheduling_priority` - The scheduling priority for jobs that are submitted with this job definition. This only affects jobs in job queues with a fair share policy. Jobs with a higher scheduling priority are scheduled before jobs with a lower scheduling priority.
* `id` - The ARN
* `eks_properties` - An [object](#eks_properties) with various properties that are specific to Amazon EKS based jobs. This must not be specified for Amazon ECS based job definitions.
* `node_properties` - An [object](#node_properties) with various properties specific to multi-node parallel jobs. If you specify node properties for a job, it becomes a multi-node parallel job. For more information, see Multi-node Parallel Jobs in the AWS Batch User Guide. If the job definition's type parameter is container, then you must specify either containerProperties or nodeProperties.
* `retry_strategy` - The [retry strategy](#retry_strategy) to use for failed jobs that are submitted with this job definition. Any retry strategy that's specified during a SubmitJob operation overrides the retry strategy defined here. If a job is terminated due to a timeout, it isn't retried.
* `timeout` - The [timeout configuration](#timeout) for jobs that are submitted with this job definition, after which AWS Batch terminates your jobs if they have not finished. If a job is terminated due to a timeout, it isn't retried. The minimum value for the timeout is 60 seconds.

### eks_properties

* `pod_properties` - The [properties](#pod_properties) for the Kubernetes pod resources of a job.

### pod_properties

* `dns_policy` - The DNS policy for the pod. The default value is ClusterFirst. If the hostNetwork parameter is not specified, the default is ClusterFirstWithHostNet. ClusterFirst indicates that any DNS query that does not match the configured cluster domain suffix is forwarded to the upstream nameserver inherited from the node.
* `host_network` - Indicates if the pod uses the hosts' network IP address. The default value is true. Setting this to false enables the Kubernetes pod networking model. Most AWS Batch workloads are egress-only and don't require the overhead of IP allocation for each pod for incoming connections.
* `service_account_name` - The name of the service account that's used to run the pod.
* `containers` - The properties of the container that's used on the Amazon EKS pod. Array of [EksContainer](#container) objects.
* `metadata` - [Metadata](#eks_metadata) about the Kubernetes pod.
* `volumes` -  Specifies the volumes for a job definition that uses Amazon EKS resources. Array of [EksVolume](#eks_volumes) objects.

### eks_container

* `args` - An array of arguments to the entrypoint
* `commands` - The entrypoint for the container. This isn't run within a shell. If this isn't specified, the ENTRYPOINT of the container image is used. Environment variable references are expanded using the container's environment.
* `env` - The environment variables to pass to a container.  Array of [EksContainerEnvironmentVariable](#eks_environment) objects.
* `image` - The Docker image used to start the container.
* `image_pull_policy` - The image pull policy for the container.
* `name` - The name of the container.
* `resources` - The type and amount of [resources](#eks_resources) to assign to a container.
* `security_context` - The [security context](#eks_security_context) for a job.
* `volume_mounts` - The [volume mounts](#eks_volume_mounts) for the container.

### eks_metadata

* `labels` - Key-value pairs used to identify, sort, and organize cube resources.

### eks_volumes

* `name` - The name of the volume. The name must be allowed as a DNS subdomain name.
* `empty_dir` - Specifies the configuration of a Kubernetes [emptyDir volume](#eks_volume_empty_dir).
* `host_path` - Specifies the configuration of a Kubernetes [hostPath volume](#eks_volume_host_path).
* `secret` - Specifies the configuration of a Kubernetes [secret volume](#eks_volume_secret).

### eks_volume_empty_dir

* `medium` - The medium to store the volume.
* `size_limit` - The maximum size of the volume. By default, there's no maximum size defined.

### eks_volume_host_path

* `path` - The path of the file or directory on the host to mount into containers on the pod.

### eks_volume_secret

* `secret_name` - The name of the secret. The name must be allowed as a DNS subdomain name
* `optional` - Specifies whether the secret or the secret's keys must be defined.

### eks_environment

* `name` - The name of the environment variable.
* `value` - The value of the environment variable.

### eks_resources

* `limits` - The type and quantity of the resources to reserve for the container.
* `requests` - The type and quantity of the resources to request for the container.

### eks_security_context

* `privileged` - When this parameter is true, the container is given elevated permissions on the host container instance. The level of permissions are similar to the root user permissions. The default value is false.
* `read_only_root_filesystem` - When this parameter is true, the container is given read-only access to its root file system. The default value is false.
* `run_as_user` - When this parameter is specified, the container is run as the specified user ID (uid). If this parameter isn't specified, the default is the user that's specified in the image metadata.
* `run_as_group` - When this parameter is specified, the container is run as the specified group ID (gid). If this parameter isn't specified, the default is the group that's specified in the image metadata.
* `run_as_non_root` - When this parameter is specified, the container is run as a user with a uid other than 0. If this parameter isn't specified, so such rule is enforced.

### eks_volume_mounts

* `mount_path` - The path on the container where the volume is mounted.
* `name` - The name the volume mount.
* `read_only` - If this value is true, the container has read-only access to the volume. Otherwise, the container can write to the volume.

### node_properties

* `main_node` - Specifies the node index for the main node of a multi-node parallel job. This node index value must be fewer than the number of nodes.
* `node_range_properties` - A list of node ranges and their [properties](#node_range_properties) that are associated with a multi-node parallel job.
* `num_nodes` - The number of nodes that are associated with a multi-node parallel job.

### node_range_properties

* `target_nodes` - The range of nodes, using node index values. A range of 0:3 indicates nodes with index values of 0 through 3. I
* `container` - The [container details](#container) for the node range.

### container

* `command` - The command that's passed to the container.
* `environment` - The [environment](#environment) variables to pass to a container.
* `ephemeral_storage` - The amount of [ephemeral storage](#ephemeral_storage) to allocate for the task. This parameter is used to expand the total amount of ephemeral storage available, beyond the default amount, for tasks hosted on AWS Fargate.
* `execution_role_arn` - The Amazon Resource Name (ARN) of the execution role that AWS Batch can assume. For jobs that run on Fargate resources, you must provide an execution role.
* `fargate_platform_configuration` - The [platform configuration](#fargate_platform_configuration) for jobs that are running on Fargate resources. Jobs that are running on EC2 resources must not specify this parameter.
* `image` - The image used to start a container.
* `instance_type` - The instance type to use for a multi-node parallel job.
* `job_role_arn` - The Amazon Resource Name (ARN) of the IAM role that the container can assume for AWS permissions.
* `linux_parameters` - [Linux-specific modifications](#linux_parameters) that are applied to the container.
* `log_configuration` - The [log configuration](#log_configuration) specification for the container.
* `mount_points` - The [mount points](#mount_points) for data volumes in your container.
* `network_configuration` - The [network configuration](#network_configuration) for jobs that are running on Fargate resources.
* `privileged` - When this parameter is true, the container is given elevated permissions on the host container instance (similar to the root user).
* `readonly_root_filesystem` - When this parameter is true, the container is given read-only access to its root file system.
* `resource_requirements` - The type and amount of [resources](#resource_requirements) to assign to a container.
* `runtime_platform` - An [object](#runtime_platform) that represents the compute environment architecture for AWS Batch jobs on Fargate.
* `secrets` - The [secrets](#secrets) for the container.
* `ulimits` - A list of [ulimits](#ulimits) to set in the container.
* `user` - The user name to use inside the container.
* `volumes` - A list of data [volumes](#volumes) used in a job.

### environment

* `name` - The name of the key-value pair.
* `value` - The value of the key-value pair.

### ephemeral_storage

* `size_in_gb` - The total amount, in GiB, of ephemeral storage to set for the task.

### fargate_platform_configuration

* `platform_version` - The AWS Fargate platform version where the jobs are running. A platform version is specified only for jobs that are running on Fargate resources.

### linux_parameters

* `init_process_enabled` - If true, run an init process inside the container that forwards signals and reaps processes.
* `max_swap` - The total amount of swap memory (in MiB) a container can use.
* `shared_memory_size` - The value for the size (in MiB) of the `/dev/shm` volume.
* `swappiness` - You can use this parameter to tune a container's memory swappiness behavior.
* `devices` - Any of the [host devices](#devices) to expose to the container.
* `tmpfs` - The container path, mount options, and size (in MiB) of the [tmpfs](#tmpfs) mount.

### log_configuration

* `options` - The configuration options to send to the log driver.
* `log_driver` - The log driver to use for the container.
* `secret_options` - The secrets to pass to the log configuration.

### network_configuration

* `assign_public_ip` - Indicates whether the job has a public IP address.

### mount_points

* `container_path` - The path on the container where the host volume is mounted.
* `read_only` - If this value is true, the container has read-only access to the volume.
* `source_volume` - The name of the volume to mount.

### resource_requirements

* `type` - The type of resource to assign to a container. The supported resources include `GPU`, `MEMORY`, and `VCPU`.
* `value` - The quantity of the specified resource to reserve for the container.

### secrets

* `name` - The name of the secret.
* `value_from` - The secret to expose to the container.

### ulimits

* `hard_limit` - The hard limit for the ulimit type.
* `name` - The type of the ulimit.
* `soft_limit` - The soft limit for the ulimit type.

### runtime_platform

* `cpu_architecture` - The vCPU architecture. The default value is X86_64. Valid values are X86_64 and ARM64.
* `operating_system_family` - The operating system for the compute environment. V

### secret_options

* `name` - The name of the secret.
* `value_from` - The secret to expose to the container. The supported values are either the full Amazon Resource Name (ARN) of the AWS Secrets Manager secret or the full ARN of the parameter in the AWS Systems Manager Parameter Store.

### devices

* `host_path` - The path for the device on the host container instance.
* `container_path` - The path inside the container that's used to expose the host device. By default, the hostPath value is used.
* `permissions` - The explicit permissions to provide to the container for the device.

### tmpfs

* `container_path` - The absolute file path in the container where the tmpfs volume is mounted.
* `size` - The size (in MiB) of the tmpfs volume.
* `mount_options` - The list of tmpfs volume mount options.

### volumes

* `name` - The name of the volume.
* `host` - The contents of the host parameter determine whether your data volume persists on the host container instance and where it's stored.
* `efs_volume_configuration` - This [parameter](#efs_volume_configuration) is specified when you're using an Amazon Elastic File System file system for job storage.

### host

* `source_path` - The path on the host container instance that's presented to the container.

### efs_volume_configuration

* `file_system_id` - The Amazon EFS file system ID to use.
* `root_directory` - The directory within the Amazon EFS file system to mount as the root directory inside the host.
* `transit_encryption` - Determines whether to enable encryption for Amazon EFS data in transit between the Amazon ECS host and the Amazon EFS server
* `transit_encryption_port` - The port to use when sending encrypted data between the Amazon ECS host and the Amazon EFS server.
* `authorization_config` - The [authorization configuration](#authorization_config) details for the Amazon EFS file system.

### authorization_config

* `access_point_id` - The Amazon EFS access point ID to use.
* `iam` - Whether or not to use the AWS Batch job IAM role defined in a job definition when mounting the Amazon EFS file system.

### retry_strategy

* `attempts` - The number of times to move a job to the RUNNABLE status.
* `evaluate_on_exit` - Array of up to 5 [objects](#evaluate_on_exit) that specify the conditions where jobs are retried or failed.

### evaluate_on_exit

* `action` - Specifies the action to take if all of the specified conditions (onStatusReason, onReason, and onExitCode) are met. The values aren't case sensitive.
* `on_exit_code` - Contains a glob pattern to match against the decimal representation of the ExitCode returned for a job.
* `on_reason` - Contains a glob pattern to match against the Reason returned for a job.
* `on_status_reason` - Contains a glob pattern to match against the StatusReason returned for a job.

### timeout

* `attempt_duration_seconds` - The job timeout time (in seconds) that's measured from the job attempt's startedAt timestamp.
