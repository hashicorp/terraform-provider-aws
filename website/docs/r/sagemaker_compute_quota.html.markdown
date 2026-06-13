---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_compute_quota"
description: |-
  Manages a SageMaker AI Compute Quota.
---

# Resource: aws_sagemaker_compute_quota

Manages a SageMaker AI Compute Quota for a SageMaker HyperPod cluster.

~> Compute quotas require task governance to be configured for the target SageMaker HyperPod cluster.

## Example Usage

```terraform
resource "aws_sagemaker_compute_quota" "example" {
  cluster_arn = var.sagemaker_hyperpod_cluster_arn
  name        = "example-quota"

  compute_quota_config {
    compute_quota_resources {
      instance_type = "ml.p4d.24xlarge"
      count         = 8
    }

    resource_sharing_config {
      strategy = "LendAndBorrow"
    }
  }

  compute_quota_target {
    team_name = "team-a"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cluster_arn` - (Required) ARN of the SageMaker HyperPod cluster.
* `compute_quota_config` - (Required) Compute allocation configuration. See [`compute_quota_config`](#compute_quota_config).
* `compute_quota_target` - (Required) Target entity for the compute quota. See [`compute_quota_target`](#compute_quota_target).
* `name` - (Required) Name of the compute quota. Must contain at least one non-whitespace character.
* `activation_state` - (Optional) Whether the compute quota is enabled. Valid values are `Enabled` and `Disabled`. Defaults to `Enabled`.
* `description` - (Optional) Description of the compute quota.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### compute_quota_config

The `compute_quota_config` block supports the following arguments:

* `compute_quota_resources` - (Required) Compute resources allocated by instance type. See [`compute_quota_resources`](#compute_quota_resources).
* `resource_sharing_config` - (Required) Idle compute sharing configuration. See [`resource_sharing_config`](#resource_sharing_config).
* `preempt_team_tasks` - (Optional) Whether workloads can preempt lower-priority same-team workloads. Valid values are `LowerPriority` and `Never`. Defaults to `LowerPriority`.

### compute_quota_resources

The `compute_quota_resources` block supports the following arguments:

* `instance_type` - (Required) Instance type of the SageMaker HyperPod cluster instance group.
* `accelerator_partition` - (Optional) Fractional GPU allocation configuration. See [`accelerator_partition`](#accelerator_partition).
* `accelerators` - (Optional) Number of accelerators to allocate.
* `count` - (Optional) Number of instances to allocate.
* `memory_in_gib` - (Optional) Amount of memory, in GiB, to allocate. Must be greater than `0` when configured.
* `vcpu` - (Optional) Number of vCPUs to allocate. Must be greater than `0` when configured.

At least one of `accelerator_partition`, `accelerators`, `count`, `memory_in_gib`, or `vcpu` must be configured.

### accelerator_partition

The `accelerator_partition` block supports the following arguments:

* `count` - (Required) Number of accelerator partitions to allocate.
* `type` - (Required) Multi-Instance GPU (MIG) profile type.

### resource_sharing_config

The `resource_sharing_config` block supports the following arguments:

* `strategy` - (Optional) Strategy for sharing idle compute within the cluster. Valid values are `DontLend`, `Lend`, and `LendAndBorrow`. Defaults to `LendAndBorrow`.
* `absolute_borrow_limits` - (Optional) Absolute limits on compute resources that can be borrowed from idle compute. Uses the same arguments as [`compute_quota_resources`](#compute_quota_resources).
* `borrow_limit` - (Optional) Percentage of idle compute the target can borrow. Valid values are between `1` and `500`. Defaults to `50`.

### compute_quota_target

The `compute_quota_target` block supports the following arguments:

* `team_name` - (Required) Name of the team to allocate compute resources to. Must contain at least one non-whitespace character.
* `fair_share_weight` - (Optional) Fair-share weight assigned to the target. Valid values are between `0` and `100`. Defaults to `0`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the compute quota.
* `compute_quota_version` - Version of the compute quota.
* `creation_time` - Creation time of the compute quota.
* `failure_reason` - Failure reason of the compute quota, if any.
* `id` - ID of the compute quota.
* `last_modified_time` - Last modified time of the compute quota.
* `status` - Status of the compute quota.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI Compute Quotas using the `id`. For example:

```terraform
import {
  to = aws_sagemaker_compute_quota.example
  id = "compute-quota-id-12345678"
}
```

Using `terraform import`, import SageMaker AI Compute Quotas using the `id`. For example:

```console
% terraform import aws_sagemaker_compute_quota.example compute-quota-id-12345678
```
