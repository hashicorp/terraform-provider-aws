---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_capacity_provider"
description: |-
  Provides an ECS cluster capacity provider.
---

# Resource: aws_ecs_capacity_provider

Provides an ECS cluster capacity provider. More information can be found on the [ECS Developer Guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/cluster-capacity-providers.html).

~> **NOTE:** Associating an ECS Capacity Provider to an Auto Scaling Group will automatically add the `AmazonECSManaged` tag to the Auto Scaling Group. This tag should be included in the `aws_autoscaling_group` resource configuration to prevent Terraform from removing it in subsequent executions as well as ensuring the `AmazonECSManaged` tag is propagated to all EC2 Instances in the Auto Scaling Group if `min_size` is above 0 on creation. Any EC2 Instances in the Auto Scaling Group without this tag must be manually be updated, otherwise they may cause unexpected scaling behavior and metrics.

~> **NOTE:** You must specify exactly one of `auto_scaling_group_provider` or `managed_instances_provider`. When using `managed_instances_provider`, the `cluster` parameter is required. When using `auto_scaling_group_provider`, the `cluster` parameter must not be set.

## Example Usage

### Auto Scaling Group Provider

```terraform
resource "aws_autoscaling_group" "example" {
  # ... other configuration, including potentially other tags ...

  tag {
    key                 = "AmazonECSManaged"
    value               = true
    propagate_at_launch = true
  }
}

resource "aws_ecs_capacity_provider" "example" {
  name = "example"

  auto_scaling_group_provider {
    auto_scaling_group_arn         = aws_autoscaling_group.example.arn
    managed_termination_protection = "ENABLED"

    managed_scaling {
      maximum_scaling_step_size = 1000
      minimum_scaling_step_size = 1
      status                    = "ENABLED"
      target_capacity           = 10
    }
  }
}
```

### Managed Instances Provider

```terraform
resource "aws_ecs_capacity_provider" "example" {
  name    = "example"
  cluster = "my-cluster"

  managed_instances_provider {
    infrastructure_role_arn = aws_iam_role.ecs_infrastructure.arn
    propagate_tags          = "CAPACITY_PROVIDER"

    instance_launch_template {
      ec2_instance_profile_arn = aws_iam_instance_profile.ecs_instance.arn
      monitoring               = "ENABLED"

      network_configuration {
        subnets         = [aws_subnet.example.id]
        security_groups = [aws_security_group.example.id]
      }

      storage_configuration {
        storage_size_gib = 30
      }

      instance_requirements {
        memory_mib {
          min = 1024
          max = 8192
        }

        vcpu_count {
          min = 1
          max = 4
        }

        instance_generations = ["current"]
        cpu_manufacturers    = ["intel", "amd"]
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `auto_scaling_group_provider` - (Optional) Configuration block for the provider for the ECS auto scaling group. Detailed below. Exactly one of `auto_scaling_group_provider` or `managed_instances_provider` must be specified.
* `cluster` - (Optional) Name of the ECS cluster. Required when using `managed_instances_provider`. Must not be set when using `auto_scaling_group_provider`.
* `managed_instances_provider` - (Optional) Configuration block for the managed instances provider. Detailed below. Exactly one of `auto_scaling_group_provider` or `managed_instances_provider` must be specified.
* `name` - (Required) Name of the capacity provider.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `auto_scaling_group_provider`

* `auto_scaling_group_arn` - (Required) - ARN of the associated auto scaling group.
* `managed_draining` - (Optional) - Enables or disables a graceful shutdown of instances without disturbing workloads. Valid values are `ENABLED` and `DISABLED`. The default value is `ENABLED` when a capacity provider is created.
* `managed_scaling` - (Optional) - Configuration block defining the parameters of the auto scaling. Detailed below.
* `managed_termination_protection` - (Optional) - Enables or disables container-aware termination of instances in the auto scaling group when scale-in happens. Valid values are `ENABLED` and `DISABLED`.

### `managed_scaling`

* `instance_warmup_period` - (Optional) Period of time, in seconds, after a newly launched Amazon EC2 instance can contribute to CloudWatch metrics for Auto Scaling group. If this parameter is omitted, the default value of 300 seconds is used.

  For more information on how the instance warmup period contributes to managed scale-out behavior, see [Control the instances Amazon ECS terminates](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/managed-termination-protection.html) in the _Amazon Elastic Container Service Developer Guide_.
* `maximum_scaling_step_size` - (Optional) Maximum step adjustment size. A number between 1 and 10,000.
* `minimum_scaling_step_size` - (Optional) Minimum step adjustment size. A number between 1 and 10,000.
* `status` - (Optional) Whether auto scaling is managed by ECS. Valid values are `ENABLED` and `DISABLED`.
* `target_capacity` - (Optional) Target utilization for the capacity provider. A number between 1 and 100.

### `managed_instances_provider`

* `infrastructure_role_arn` - (Required) The Amazon Resource Name (ARN) of the infrastructure role that Amazon ECS uses to manage instances on your behalf. This role must have permissions to launch, terminate, and manage Amazon EC2 instances, as well as access to other AWS services required for Amazon ECS Managed Instances functionality. For more information, see [Amazon ECS infrastructure IAM role](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/infrastructure_IAM_role.html) in the Amazon ECS Developer Guide.
* `instance_launch_template` - (Required) The launch template configuration that specifies how Amazon ECS should launch Amazon EC2 instances. This includes the instance profile, network configuration, storage settings, and instance requirements for attribute-based instance type selection. For more information, see [Store instance launch parameters in Amazon EC2 launch templates](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-launch-templates.html) in the Amazon EC2 User Guide. Detailed below.
* `propagate_tags` - (Optional) Specifies whether to propagate tags from the capacity provider to the Amazon ECS Managed Instances. When enabled, tags applied to the capacity provider are automatically applied to all instances launched by this provider. Valid values are `CAPACITY_PROVIDER` and `NONE`.
* `infrastructure_optimization` - (Optional) Defines how Amazon ECS Managed Instances optimizes the infrastructure in your capacity provider. Configure it to turn on or off the infrastructure optimization in your capacity provider, and to control the idle EC2 instances optimization delay.

### `instance_launch_template`

* `capacity_option_type` - (Optional) The purchasing option for the EC2 instances used in the capacity provider. Determines whether to use On-Demand or Spot instances. Valid values are `ON_DEMAND` and `SPOT`. Defaults to `ON_DEMAND` when not specified. Changing this value will trigger replacement of the capacity provider. For more information, see [Amazon EC2 billing and purchasing options](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-purchasing-options.html) in the Amazon EC2 User Guide.
* `ec2_instance_profile_arn` - (Required) The Amazon Resource Name (ARN) of the instance profile that Amazon ECS applies to Amazon ECS Managed Instances. This instance profile must include the necessary permissions for your tasks to access AWS services and resources. For more information, see [Amazon ECS instance profile for Managed Instances](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/instance_IAM_role.html) in the Amazon ECS Developer Guide.
* `instance_requirements` - (Optional) The instance requirements. You can specify the instance types and instance requirements such as vCPU count, memory, network performance, and accelerator specifications. Amazon ECS automatically selects the instances that match the specified criteria. Detailed below.
* `monitoring` - (Optional) CloudWatch provides two categories of monitoring: basic monitoring and detailed monitoring. By default, your managed instance is configured for basic monitoring. You can optionally enable detailed monitoring to help you more quickly identify and act on operational issues. You can enable or turn off detailed monitoring at launch or when the managed instance is running or stopped. For more information, see [Detailed monitoring for Amazon ECS Managed Instances](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/cloudwatch-metrics.html) in the Amazon ECS Developer Guide. Valid values are `BASIC` and `DETAILED`.
* `network_configuration` - (Required) The network configuration for Amazon ECS Managed Instances. This specifies the subnets and security groups that instances use for network connectivity. Detailed below.
* `storage_configuration` - (Optional) The storage configuration for Amazon ECS Managed Instances. This defines the root volume size and type for the instances. Detailed below.

### `network_configuration`

* `security_groups` - (Optional) The list of security group IDs to apply to Amazon ECS Managed Instances. These security groups control the network traffic allowed to and from the instances.
* `subnets` - (Required) The list of subnet IDs where Amazon ECS can launch Amazon ECS Managed Instances. Instances are distributed across the specified subnets for high availability. All subnets must be in the same VPC.

### `storage_configuration`

* `storage_size_gib` - (Required) The size of the tasks volume in GiB. Must be at least 1.

### `instance_requirements`

* `accelerator_count` - (Optional) The minimum and maximum number of accelerators for the instance types. This is used when you need instances with specific numbers of GPUs or other accelerators.
* `accelerator_manufacturers` - (Optional) The accelerator manufacturers to include. You can specify `nvidia`, `amd`, `amazon-web-services`, `xilinx`, or `habana` depending on your accelerator requirements. Valid values are `amazon-web-services`, `amd`, `nvidia`, `xilinx`, `habana`.
* `accelerator_names` - (Optional) The specific accelerator names to include. For example, you can specify `a100`, `v100`, `k80`, or other specific accelerator models. Valid values are `a100`, `inferentia`, `k520`, `k80`, `m60`, `radeon-pro-v520`, `t4`, `vu9p`, `v100`, `a10g`, `h100`, `t4g`.
* `accelerator_total_memory_mib` - (Optional) The minimum and maximum total accelerator memory in mebibytes (MiB). This is important for GPU workloads that require specific amounts of video memory.
* `accelerator_types` - (Optional) The accelerator types to include. You can specify `gpu` for graphics processing units, `fpga` for field programmable gate arrays, or `inference` for machine learning inference accelerators. Valid values are `gpu`, `fpga`, `inference`.
* `allowed_instance_types` - (Optional) The instance types to include in the selection. When specified, Amazon ECS only considers these instance types, subject to the other requirements specified. Maximum of 400 instance types. You can specify instance type patterns using wildcards (e.g., `m5.*`).
* `bare_metal` - (Optional) Indicates whether to include bare metal instance types. Set to `included` to allow bare metal instances, `excluded` to exclude them, or `required` to use only bare metal instances. Valid values are `included`, `excluded`, `required`.
* `baseline_ebs_bandwidth_mbps` - (Optional) The minimum and maximum baseline Amazon EBS bandwidth in megabits per second (Mbps). This is important for workloads with high storage I/O requirements.
* `burstable_performance` - (Optional) Indicates whether to include burstable performance instance types (T2, T3, T3a, T4g). Set to `included` to allow burstable instances, `excluded` to exclude them, or `required` to use only burstable instances. Valid values are `included`, `excluded`, `required`.
* `cpu_manufacturers` - (Optional) The CPU manufacturers to include or exclude. You can specify `intel`, `amd`, or `amazon-web-services` to control which CPU types are used for your workloads. Valid values are `intel`, `amd`, `amazon-web-services`.
* `excluded_instance_types` - (Optional) The instance types to exclude from selection. Use this to prevent Amazon ECS from selecting specific instance types that may not be suitable for your workloads. Maximum of 400 instance types.
* `instance_generations` - (Optional) The instance generations to include. You can specify `current` to use the latest generation instances, or `previous` to include previous generation instances for cost optimization. Valid values are `current`, `previous`.
* `local_storage` - (Optional) Indicates whether to include instance types with local storage. Set to `included` to allow local storage, `excluded` to exclude it, or `required` to use only instances with local storage. Valid values are `included`, `excluded`, `required`.
* `local_storage_types` - (Optional) The local storage types to include. You can specify `hdd` for hard disk drives, `ssd` for solid state drives, or both. Valid values are `hdd`, `ssd`.
* `max_spot_price_as_percentage_of_optimal_on_demand_price` - (Optional) The maximum price for Spot instances as a percentage of the optimal On-Demand price. This provides more precise cost control for Spot instance selection.
* `memory_gib_per_vcpu` - (Optional) The minimum and maximum amount of memory per vCPU in gibibytes (GiB). This helps ensure that instance types have the appropriate memory-to-CPU ratio for your workloads.
* `memory_mib` - (Required) The minimum and maximum amount of memory in mebibytes (MiB) for the instance types. Amazon ECS selects instance types that have memory within this range.
* `network_bandwidth_gbps` - (Optional) The minimum and maximum network bandwidth in gigabits per second (Gbps). This is crucial for network-intensive workloads that require high throughput.
* `network_interface_count` - (Optional) The minimum and maximum number of network interfaces for the instance types. This is useful for workloads that require multiple network interfaces.
* `on_demand_max_price_percentage_over_lowest_price` - (Optional) The price protection threshold for On-Demand Instances, as a percentage higher than an identified On-Demand price. The identified On-Demand price is the price of the lowest priced current generation C, M, or R instance type with your specified attributes. When Amazon ECS selects instance types with your attributes, it will exclude instance types whose price exceeds your specified threshold.
* `require_hibernate_support` - (Optional) Indicates whether the instance types must support hibernation. When set to `true`, only instance types that support hibernation are selected.
* `spot_max_price_percentage_over_lowest_price` - (Optional) The maximum price for Spot instances as a percentage over the lowest priced On-Demand instance. This helps control Spot instance costs while maintaining access to capacity.
* `total_local_storage_gb` - (Optional) The minimum and maximum total local storage in gigabytes (GB) for instance types with local storage.
* `vcpu_count` - (Required) The minimum and maximum number of vCPUs for the instance types. Amazon ECS selects instance types that have vCPU counts within this range.

### `infrastructure_optimization`

* `scale_in_after` - (Optional) This parameter defines the number of seconds Amazon ECS Managed Instances waits before optimizing EC2 instances that have become idle or underutilized. A longer delay increases the likelihood of placing new tasks on idle instances, reducing startup time. A shorter delay helps reduce infrastructure costs by optimizing idle instances more quickly. Valid values are:
    * Not set (null) - Uses the default optimization behavior.
    * `-1` - Disables automatic infrastructure optimization.
    * `0` to `3600` (inclusive) - Specifies the number of seconds to wait before optimizing instances.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN that identifies the capacity provider.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ecs_capacity_provider.example
  identity = {
    "arn" = "arn:aws:ecs:us-west-2:123456789012:capacity-provider/example"
  }
}

resource "aws_ecs_capacity_provider" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) Amazon Resource Name (ARN) of the ECS capacity provider.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECS Capacity Providers using the `arn`. For example:

```terraform
import {
  to = aws_ecs_capacity_provider.example
  id = "arn:aws:ecs:us-west-2:123456789012:capacity-provider/example"
}
```

Using `terraform import`, import ECS Capacity Providers using the `arn`. For example:

```console
% terraform import aws_ecs_capacity_provider.example arn:aws:ecs:us-west-2:123456789012:capacity-provider/example
```
