---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_fleet"
description: |-
  Provides a resource to manage EC2 Fleets
---

# Resource: aws_ec2_fleet

Provides a resource to manage EC2 Fleets.

## Example Usage

```terraform
resource "aws_ec2_fleet" "example" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = aws_launch_template.example.id
      version            = aws_launch_template.example.latest_version
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 5
  }
}
```

## Argument Reference

The following arguments are supported:

* `launch_template_config` - (Required) Nested argument containing EC2 Launch Template configurations. Defined below.
* `target_capacity_specification` - (Required) Nested argument containing target capacity configurations. Defined below.
* `context` - (Optional) Reserved.
* `excess_capacity_termination_policy` - (Optional) Whether running instances should be terminated if the total target capacity of the EC2 Fleet is decreased below the current size of the EC2. Valid values: `no-termination`, `termination`. Defaults to `termination`.
* `on_demand_options` - (Optional) Nested argument containing On-Demand configurations. Defined below.
* `replace_unhealthy_instances` - (Optional) Whether EC2 Fleet should replace unhealthy instances. Defaults to `false`.
* `spot_options` - (Optional) Nested argument containing Spot configurations. Defined below.
* `tags` - (Optional) Map of Fleet tags. To tag instances at launch, specify the tags in the Launch Template. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `terminate_instances` - (Optional) Whether to terminate instances for an EC2 Fleet if it is deleted successfully. Defaults to `false`.
* `terminate_instances_with_expiration` - (Optional) Whether running instances should be terminated when the EC2 Fleet expires. Defaults to `false`.
* `type` - (Optional) The type of request. Indicates whether the EC2 Fleet only requests the target capacity, or also attempts to maintain it. Valid values: `maintain`, `request`. Defaults to `maintain`.

### launch_template_config

* `launch_template_specification` - (Required) Nested argument containing EC2 Launch Template to use. Defined below.
* `override` - (Optional) Nested argument(s) containing parameters to override the same parameters in the Launch Template. Defined below.

#### launch_template_specification

~> *NOTE:* Either `launch_template_id` or `launch_template_name` must be specified.

* `version` - (Required) Version number of the launch template.
* `launch_template_id` - (Optional) ID of the launch template.
* `launch_template_name` - (Optional) Name of the launch template.

#### override

Example:

```terraform
resource "aws_ec2_fleet" "example" {
  # ... other configuration ...

  launch_template_config {
    # ... other configuration ...

    override {
      instance_type     = "m4.xlarge"
      weighted_capacity = 1
    }

    override {
      instance_type     = "m4.2xlarge"
      weighted_capacity = 2
    }
  }
}
```

* `availability_zone` - (Optional) Availability Zone in which to launch the instances.
* `instance_requirements` - (Optional) Override the instance type in the Launch Template with instance types that satisfy the requirements.
* `instance_type` - (Optional) Instance type.
* `max_price` - (Optional) Maximum price per unit hour that you are willing to pay for a Spot Instance.
* `priority` - (Optional) Priority for the launch template override. If `on_demand_options` `allocation_strategy` is set to `prioritized`, EC2 Fleet uses priority to determine which launch template override to use first in fulfilling On-Demand capacity. The highest priority is launched first. The lower the number, the higher the priority. If no number is set, the launch template override has the lowest priority. Valid values are whole numbers starting at 0.
* `subnet_id` - (Optional) ID of the subnet in which to launch the instances.
* `weighted_capacity` - (Optional) Number of units provided by the specified instance type.

##### instance_requirements

This configuration block supports the following:

~> **NOTE**: Both `memory_mib.min` and `vcpu_count.min` must be specified.

* `accelerator_count` - (Optional) Block describing the minimum and maximum number of accelerators (GPUs, FPGAs, or AWS Inferentia chips). Default is no minimum or maximum.
    * `min` - (Optional) Minimum.
    * `max` - (Optional) Maximum. Set to `0` to exclude instance types with accelerators.
* `accelerator_manufacturers` - (Optional) List of accelerator manufacturer names. Default is any manufacturer.

    ```
    Valid names:
      * amazon-web-services
      * amd
      * nvidia
      * xilinx
    ```

* `accelerator_names` - (Optional) List of accelerator names. Default is any acclerator.

    ```
    Valid names:
      * a100            - NVIDIA A100 GPUs
      * v100            - NVIDIA V100 GPUs
      * k80             - NVIDIA K80 GPUs
      * t4              - NVIDIA T4 GPUs
      * m60             - NVIDIA M60 GPUs
      * radeon-pro-v520 - AMD Radeon Pro V520 GPUs
      * vu9p            - Xilinx VU9P FPGAs
    ```

* `accelerator_total_memory_mib` - (Optional) Block describing the minimum and maximum total memory of the accelerators. Default is no minimum or maximum.
    * `min` - (Optional) Minimum.
    * `max` - (Optional) Maximum.
* `accelerator_types` - (Optional) List of accelerator types. Default is any accelerator type.

    ```
    Valid types:
      * fpga
      * gpu
      * inference
    ```

* `bare_metal` - (Optional) Indicate whether bare metal instace types should be `included`, `excluded`, or `required`. Default is `excluded`.
* `baseline_ebs_bandwidth_mbps` - (Optional) Block describing the minimum and maximum baseline EBS bandwidth, in Mbps. Default is no minimum or maximum.
    * `min` - (Optional) Minimum.
    * `max` - (Optional) Maximum.
* `burstable_performance` - (Optional) Indicate whether burstable performance instance types should be `included`, `excluded`, or `required`. Default is `excluded`.
* `cpu_manufacturers` (Optional) List of CPU manufacturer names. Default is any manufacturer.

    ~> **NOTE**: Don't confuse the CPU hardware manufacturer with the CPU hardware architecture. Instances will be launched with a compatible CPU architecture based on the Amazon Machine Image (AMI) that you specify in your launch template.

    ```
    Valid names:
      * amazon-web-services
      * amd
      * intel
    ```

* `excluded_instance_types` - (Optional) List of instance types to exclude. You can use strings with one or more wild cards, represented by an asterisk (\*). The following are examples: `c5*`, `m5a.*`, `r*`, `*3*`. For example, if you specify `c5*`, you are excluding the entire C5 instance family, which includes all C5a and C5n instance types. If you specify `m5a.*`, you are excluding all the M5a instance types, but not the M5n instance types. Maximum of 400 entries in the list; each entry is limited to 30 characters. Default is no excluded instance types.
* `instance_generations` - (Optional) List of instance generation names. Default is any generation.

    ```
    Valid names:
      * current  - Recommended for best performance.
      * previous - For existing applications optimized for older instance types.
    ```

* `local_storage` - (Optional) Indicate whether instance types with local storage volumes are `included`, `excluded`, or `required`. Default is `included`.
* `local_storage_types` - (Optional) List of local storage type names. Default any storage type.

    ```
    Value names:
      * hdd - hard disk drive
      * ssd - solid state drive
    ```

* `memory_gib_per_vcpu` - (Optional) Block describing the minimum and maximum amount of memory (GiB) per vCPU. Default is no minimum or maximum.
    * `min` - (Optional) Minimum. May be a decimal number, e.g. `0.5`.
    * `max` - (Optional) Maximum. May be a decimal number, e.g. `0.5`.
* `memory_mib` - (Required) Block describing the minimum and maximum amount of memory (MiB). Default is no maximum.
    * `min` - (Required) Minimum.
    * `max` - (Optional) Maximum.
* `network_interface_count` - (Optional) Block describing the minimum and maximum number of network interfaces. Default is no minimum or maximum.
    * `min` - (Optional) Minimum.
    * `max` - (Optional) Maximum.
* `on_demand_max_price_percentage_over_lowest_price` - (Optional) The price protection threshold for On-Demand Instances. This is the maximum you’ll pay for an On-Demand Instance, expressed as a percentage higher than the cheapest M, C, or R instance type with your specified attributes. When Amazon EC2 Auto Scaling selects instance types with your attributes, we will exclude instance types whose price is higher than your threshold. The parameter accepts an integer, which Amazon EC2 Auto Scaling interprets as a percentage. To turn off price protection, specify a high value, such as 999999. Default is 20.

    If you set DesiredCapacityType to vcpu or memory-mib, the price protection threshold is applied based on the per vCPU or per memory price instead of the per instance price.
* `require_hibernate_support` - (Optional) Indicate whether instance types must support On-Demand Instance Hibernation, either `true` or `false`. Default is `false`.
* `spot_max_price_percentage_over_lowest_price` - (Optional) The price protection threshold for Spot Instances. This is the maximum you’ll pay for a Spot Instance, expressed as a percentage higher than the cheapest M, C, or R instance type with your specified attributes. When Amazon EC2 Auto Scaling selects instance types with your attributes, we will exclude instance types whose price is higher than your threshold. The parameter accepts an integer, which Amazon EC2 Auto Scaling interprets as a percentage. To turn off price protection, specify a high value, such as 999999. Default is 100.

    If you set DesiredCapacityType to vcpu or memory-mib, the price protection threshold is applied based on the per vCPU or per memory price instead of the per instance price.
* `total_local_storage_gb` - (Optional) Block describing the minimum and maximum total local storage (GB). Default is no minimum or maximum.
    * `min` - (Optional) Minimum. May be a decimal number, e.g. `0.5`.
    * `max` - (Optional) Maximum. May be a decimal number, e.g. `0.5`.
* `vcpu_count` - (Required) Block describing the minimum and maximum number of vCPUs. Default is no maximum.
    * `min` - (Required) Minimum.
    * `max` - (Optional) Maximum.

### on_demand_options

* `allocation_strategy` - (Optional) The order of the launch template overrides to use in fulfilling On-Demand capacity. Valid values: `lowestPrice`, `prioritized`. Default: `lowestPrice`.

### spot_options

* `allocation_strategy` - (Optional) How to allocate the target capacity across the Spot pools. Valid values: `diversified`, `lowestPrice`, `capacity-optimized` and `capacity-optimized-prioritized`. Default: `lowestPrice`.
* `instance_interruption_behavior` - (Optional) Behavior when a Spot Instance is interrupted. Valid values: `hibernate`, `stop`, `terminate`. Default: `terminate`.
* `instance_pools_to_use_count` - (Optional) Number of Spot pools across which to allocate your target Spot capacity. Valid only when Spot `allocation_strategy` is set to `lowestPrice`. Default: `1`.
* `maintenance_strategies` - (Optional) Nested argument containing maintenance strategies for managing your Spot Instances that are at an elevated risk of being interrupted. Defined below.


### maintenance_strategies

* `capacity_rebalance` - (Optional) Nested argument containing the capacity rebalance for your fleet request. Defined below.

### capacity_rebalance

* `replacement_strategy` - (Optional) The replacement strategy to use. Only available for fleets of `type` set to `maintain`. Valid values: `launch`.



### target_capacity_specification

* `default_target_capacity_type` - (Required) Default target capacity type. Valid values: `on-demand`, `spot`.
* `total_target_capacity` - (Required) The number of units to request, filled using `default_target_capacity_type`.
* `on_demand_target_capacity` - (Optional) The number of On-Demand units to request.
* `spot_target_capacity` - (Optional) The number of Spot units to request.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Fleet identifier
* `arn` - The ARN of the fleet
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

`aws_ec2_fleet` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `10m`) How long to wait for a fleet to be active.
* `update` - (Default `10m`) How long to wait for a fleet to be modified.
* `delete` - (Default `10m`) How long to wait for a fleet to be deleted. If `terminate_instances` is `true`, how long to wait for instances to terminate.

## Import

`aws_ec2_fleet` can be imported by using the Fleet identifier, e.g.,

```
$ terraform import aws_ec2_fleet.example fleet-b9b55d27-c5fc-41ac-a6f3-48fcc91f080c
```
