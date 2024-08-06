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

* `name` - (Required) Name of the job definition.
* `type` - (Required) Type of job definition. Must be `container` or `multinode`.

The following arguments are optional:

* `container_properties` - (Optional) Valid [container properties](http://docs.aws.amazon.com/batch/latest/APIReference/API_RegisterJobDefinition.html) provided as a single valid JSON document. This parameter is only valid if the `type` parameter is `container`.
* `deregister_on_new_revision` - (Optional) When updating a job definition a new revision is created. This parameter determines if the previous version is `deregistered` (`INACTIVE`) or left  `ACTIVE`. Defaults to `true`.
* `eks_properties` - (Optional) Valid [eks properties](#eks_properties). This parameter is only valid if the `type` parameter is `container`.
* `node_properties` - (Optional) Valid [node properties](http://docs.aws.amazon.com/batch/latest/APIReference/API_RegisterJobDefinition.html) provided as a single valid JSON document. This parameter is required if the `type` parameter is `multinode`.
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

* `containers` - (Optional) Properties of the container that's used on the Amazon EKS pod. See [containers](#containers) below.
* `dns_policy` - (Optional) DNS policy for the pod. The default value is `ClusterFirst`. If the `host_network` argument is not specified, the default is `ClusterFirstWithHostNet`. `ClusterFirst` indicates that any DNS query that does not match the configured cluster domain suffix is forwarded to the upstream nameserver inherited from the node. For more information, see Pod's DNS policy in the Kubernetes documentation.
* `host_network` - (Optional) Whether the pod uses the hosts' network IP address. The default value is `true`. Setting this to `false` enables the Kubernetes pod networking model. Most AWS Batch workloads are egress-only and don't require the overhead of IP allocation for each pod for incoming connections.
* `image_pull_secret` - (Optional) List of Kubernetes secret resources. See [`image_pull_secret`](#image_pull_secret) below.
* `metadata` - (Optional) Metadata about the Kubernetes pod.
* `service_account_name` - (Optional) Name of the service account that's used to run the pod.
* `volumes` - (Optional) Volumes for a job definition that uses Amazon EKS resources. AWS Batch supports [emptyDir](#eks_empty_dir), [hostPath](#eks_host_path), and [secret](#eks_secret) volume types.

#### `containers`

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

* `attempt_duration_seconds` - (Optional) Time duration in seconds after which AWS Batch terminates your jobs if they have not finished. The minimum value for the timeout is `60` seconds.

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
