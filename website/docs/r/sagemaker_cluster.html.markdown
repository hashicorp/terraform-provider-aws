---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_cluster"
description: |-
  Terraform resource for managing an AWS SageMaker AI Cluster.
---
# Resource: aws_sagemaker_cluster

Terraform resource for managing an AWS SageMaker AI Cluster.

## Example Usage

### Basic Usage for Slurm Orchestrator
```terraform
resource "aws_sagemaker_cluster" "example" {
  cluster_name = "name"
  instance_group {
    execution_role      = aws_iam_role.example.arn
    instance_count      = 2
    instance_group_name = "example"
    instance_type       = "ml.t3.medium"
    lifecycle_config {
      on_create     = "on_create.sh"
      source_s3_uri = "s3://${aws_s3_bucket.example.id}/sagemaker-lifecycle/root"
    }
    instance_storage_config {
      ebs_volume_config {
        volume_size_in_gb = 20
      }
    }
    threads_per_core = 1
  }
  instance_group {
    execution_role      = aws_iam_role.example.arn
    instance_count      = 1
    instance_group_name = "example-worker"
    instance_type       = "ml.t3.medium"
    lifecycle_config {
      on_create     = "on_create.sh"
      source_s3_uri = "s3://${aws_s3_bucket.example.id}/sagemaker-lifecycle/root"
    }
    instance_storage_config {
      ebs_volume_config {
        volume_size_in_gb = 20
      }
    }
    threads_per_core = 1
  }
  vpc_config {
    security_group_ids = [aws_security_group.example.id]
    subnets            = [aws_subnet.example.id]
  }
}
```

### Basic Usage for EKS Orchestrator
```terraform
resource "aws_sagemaker_cluster" "example" {
  cluster_name = "name"
  instance_group {
    execution_role      = aws_iam_role.example.arn
    instance_count      = 2
    instance_group_name = "example"
    instance_type       = "ml.t3.medium"
    lifecycle_config {
      on_create     = "on_create.sh"
      source_s3_uri = "s3://${aws_s3_bucket.example.id}/sagemaker-lifecycle/root"
    }
    instance_storage_config {
      ebs_volume_config {
        volume_size_in_gb = 20
      }
    }
    threads_per_core = 1
  }
  instance_group {
    execution_role      = aws_iam_role.example.arn
    instance_count      = 1
    instance_group_name = "example-worker"
    instance_type       = "ml.t3.medium"
    lifecycle_config {
      on_create     = "on_create.sh"
      source_s3_uri = "s3://${aws_s3_bucket.example.id}/sagemaker-lifecycle/root"
    }
    instance_storage_config {
      ebs_volume_config {
        volume_size_in_gb = 20
      }
    }
    threads_per_core = 1
  }
  orchestrator {
    eks {
      cluster_arn = aws_eks_cluster.example.arn
    }
  }
  vpc_config {
    security_group_ids = [aws_security_group.example.id]
    subnets            = [aws_subnet.example.id]
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `cluster_name` - (Required, Forces new resource) The name of the HyperPod Cluster. Must be 1-63 characters and can contain alphanumeric characters and hyphens, but cannot start or end with a hyphen.
* `instance_group` - (Required) The instance groups of the SageMaker HyperPod cluster. You can specify between 1 and 100 instance groups. See [instance_group](#instance_group) below.
* `node_recovery` - (Optional, Computed) Valid options are `Automatic` and `None`. If node auto-recovery is set to `Automatic`, faulty nodes will be replaced or rebooted when a failure is detected. If set to `None`, nodes will be labelled when a fault is detected. Defaults to `Automatic`.
* `orchestrator` - (Optional, Forces new resource) Specifies parameter(s) specific to the orchestrator. Omit to use Slurm orchestrator. See [orchestrator](#orchestrator) below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_config` - (Optional, Forces new resource) Specifies an Amazon Virtual Private Cloud (VPC) that cluster resources have access to. See [vpc_config](#vpc_config) below.

### instance_group

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

**Note:** Modifying certain attributes of an instance group (such as `execution_role`, `instance_type`, `instance_storage_config`, `override_vpc_config`, `scheduled_update_config`, `threads_per_core`, or `training_plan_arn`) will force replacement of the instance group.

The following arguments are required:

* `execution_role` - (Required, Forces new resource) The ARN of the execution role for the instance group to assume.
* `instance_count` - (Required) The number of instances to add to the instance group of a SageMaker HyperPod cluster. Must be between 0 and 6758.
* `instance_group_name` - (Required, Forces new resource) The name of the instance group of a SageMaker HyperPod cluster. Must be 1-63 characters and can contain alphanumeric characters and hyphens, but cannot start or end with a hyphen.
* `instance_type` - (Required, Forces new resource) The instance type of the instance group of a SageMaker HyperPod cluster.
* `lifecycle_config` - (Required) The lifecycle configuration for a SageMaker HyperPod cluster. See [lifecycle_config](#lifecycle_config) below.

The following arguments are optional:

* `instance_storage_config` - (Optional, Forces new resource) The instance storage configuration for the instance group. See [instance_storage_config](#instance_storage_config) below.
* `on_start_deep_healthchecks` - (Optional) Nodes will undergo advanced stress test to detect and replace faulty instances, based on the type of deep health check(s) passed in. Valid options are `InstanceStress` and `InstanceConnectivity`. Only specify if using EKS cluster.
* `override_vpc_config` - (Optional, Forces new resource) Specifies an Amazon Virtual Private Cloud (VPC) at the instance group level that overrides the default Amazon VPC configuration of the SageMaker HyperPod cluster. See [vpc_config](#vpc_config) below.
* `scheduled_update_config` - (Optional, Forces new resource) The configuration object of the schedule that SageMaker follows when updating the AMI. See [scheduled_update_config](#scheduled_update_config) below.
* `threads_per_core` - (Optional, Computed, Forces new resource) Enabling or disabling multithreading. For instance types that support multithreading, specify `1` for disabling multithreading and `2` for enabling multithreading.
* `training_plan_arn` - (Optional, Forces new resource) The ARN of the training plan associated with the cluster instance group.

### lifecycle_config

The following arguments are required:

* `on_create` - (Required) The file name of the entrypoint script of lifecycle scripts under `source_s3_uri`. This entrypoint script runs during cluster creation. Must be between 1 and 128 characters.
* `source_s3_uri` - (Required) An Amazon S3 bucket path where the lifecycle scripts are stored. Must be a valid S3 or HTTPS URI with maximum length of 1024 characters.

### instance_storage_config

The following arguments are optional:

* `ebs_volume_config` - (Optional, Forces new resource) Defines the configuration for attaching additional Amazon Elastic Block Store (EBS) volumes to the instances in the SageMaker HyperPod cluster instance group. The additional EBS volume is attached to each instance within the SageMaker HyperPod cluster instance group and mounted to `/opt/sagemaker`. See [ebs_volume_config](#ebs_volume_config) below.

### ebs_volume_config

The following arguments are required:

* `volume_size_in_gb` - (Required, Forces new resource) The size in gigabytes (GB) of the additional EBS volume to be attached to the instances in the SageMaker HyperPod cluster instance group. Must be between 1 and 16384 GB.

### scheduled_update_config

The following arguments are required:

* `schedule_expression` - (Required, Forces new resource) A cron expression that specifies the schedule that SageMaker follows when updating the AMI. Must be between 1 and 256 characters.

The following arguments are optional:

* `deployment_config` - (Optional, Forces new resource) The configuration to use when updating the AMI versions. See [deployment_config](#deployment_config) below.

### deployment_config

The following arguments are optional:

* `auto_rollback_configuration` - (Optional, Forces new resource) A list of alarms that SageMaker monitors to know whether to roll back the AMI update. You can specify between 1 and 10 alarms. See [auto_rollback_configuration](#auto_rollback_configuration) below.
* `rolling_update_policy` - (Optional, Forces new resource) The policy that SageMaker uses when updating the AMI versions of the cluster. See [rolling_update_policy](#rolling_update_policy) below.
* `wait_interval` - (Optional, Forces new resource) The duration in seconds that SageMaker waits before updating more instances in the cluster. Must be between 0 and 3600 seconds.

### auto_rollback_configuration

The following arguments are required:

* `alarm_name` - (Required, Forces new resource) The name of the alarm. Must be between 1 and 255 characters and cannot contain whitespace.

### rolling_update_policy

The following arguments are required:

* `maximum_batch_size` - (Required, Forces new resource) The maximum amount of instances in the cluster that SageMaker can update at a time. See [maximum_batch_size](#maximum_batch_size) below.

The following arguments are optional:

* `rollback_maximum_batch_size` - (Optional, Forces new resource) The maximum amount of instances in the cluster that SageMaker can rollback at a time. See [maximum_batch_size](#maximum_batch_size) below.

### maximum_batch_size

The following arguments are required:

* `type` - (Required, Forces new resource) Specifies whether SageMaker should process the update by amount or percentage of instances. Valid values are `INSTANCE_COUNT` and `CAPACITY_PERCENTAGE`.
* `value` - (Required, Forces new resource) Specifies the amount or percentage of instances SageMaker updates at a time.

### orchestrator

The following arguments are optional:

* `eks` - (Optional) Specifies parameter(s) related to EKS as orchestrator. See [eks](#eks) below.

### eks

The following arguments are required:

* `cluster_arn` - (Required, Forces new resource) The ARN of the EKS cluster.

### vpc_config

The following arguments are required:

* `security_group_ids` - (Required, Forces new resource) Security groups for the VPC that is specified in the `subnets` field. You can specify between 1 and 5 security group IDs.
* `subnets` - (Required, Forces new resource) The ID of the subnets in the VPC to connect the cluster to. You can specify between 1 and 16 subnet IDs.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Cluster.
* `cluster_status` - The status of the cluster.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

The `instance_group` configuration block exports the following additional attributes:

* `nodes` - Information about the cluster nodes. See [nodes](#nodes) below.

### nodes

* `instance_group_name` - The name of the instance group the node is part of.
* `instance_id` - The instance ID of the cluster node.
* `instance_status` - The status of the instance. See [instance_status](#instance_status) below.
* `instance_storage_configs` - The instance storage configuration of the cluster node.
* `instance_type` - The instance type of the cluster node.
* `last_software_update_time` - The last time the cluster node software was updated.
* `launch_time` - The time when the cluster node was launched.
* `lifecycle_config` - The lifecycle configuration of the cluster node.
* `override_vpc_config` - The VPC configuration override of the cluster node.
* `placement` - The placement configuration of the cluster node. See [placement](#placement) below.
* `private_dns_hostname` - The private DNS hostname of the cluster node.
* `private_primary_ip` - The private primary IP address of the cluster node.
* `private_primary_ipv6` - The private primary IPv6 address of the cluster node.
* `threads_per_core` - The number of threads per core of the cluster node.

### instance_status

* `message` - A message describing the instance status.
* `status` - The status of the instance.

### placement

* `availability_zone` - The Availability Zone of the instance.
* `availability_zone_id` - The Availability Zone ID of the instance.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI Cluster using the `cluster_name`. For example:

```terraform
import {
  to = aws_sagemaker_cluster.example
  id = "my-cluster"
}
```

Using `terraform import`, import SageMaker AI Cluster using the `cluster_name`. For example:

```console
% terraform import aws_sagemaker_cluster.example my-cluster
```
