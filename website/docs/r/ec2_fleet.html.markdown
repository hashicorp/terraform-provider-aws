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

This resource supports the following arguments:

* `context` - (Optional) Reserved.
* `excess_capacity_termination_policy` - (Optional) Whether running instances should be terminated if the total target capacity of the EC2 Fleet is decreased below the current size of the EC2. Valid values: `no-termination`, `termination`. Defaults to `termination`. Supported only for fleets of type `maintain`.
* `launch_template_config` - (Required) Nested argument containing EC2 Launch Template configurations. Defined below.
* `on_demand_options` - (Optional) Nested argument containing On-Demand configurations. Defined below.
* `replace_unhealthy_instances` - (Optional) Whether EC2 Fleet should replace unhealthy instances. Defaults to `false`. Supported only for fleets of type `maintain`.
* `spot_options` - (Optional) Nested argument containing Spot configurations. Defined below.
* `tags` - (Optional) Map of Fleet tags. To tag instances at launch, specify the tags in the Launch Template. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `target_capacity_specification` - (Required) Nested argument containing target capacity configurations. Defined below.
* `terminate_instances` - (Optional) Whether to terminate instances for an EC2 Fleet if it is deleted successfully. Defaults to `false`.
* `terminate_instances_with_expiration` - (Optional) Whether running instances should be terminated when the EC2 Fleet expires. Defaults to `false`.
* `type` - (Optional) The type of request. Indicates whether the EC2 Fleet only requests the target capacity, or also attempts to maintain it. Valid values: `maintain`, `request`, `instant`. Defaults to `maintain`.
* `valid_from` - (Optional) The start date and time of the request, in UTC format (for example, YYYY-MM-DDTHH:MM:SSZ). The default is to start fulfilling the request immediately.
* `valid_until` - (Optional) The end date and time of the request, in UTC format (for example, YYYY-MM-DDTHH:MM:SSZ). At this point, no new EC2 Fleet requests are placed or able to fulfill the request. If no value is specified, the request remains until you cancel it.

### launch_template_config

Describes a launch template and overrides.

* `launch_template_specification` - (Optional) Nested argument containing EC2 Launch Template to use. Defined below.
* `override` - (Optional) Nested argument(s) containing parameters to override the same parameters in the Launch Template. Defined below.

#### launch_template_specification

The launch template to use. You must specify either the launch template ID or launch template name in the request.

* `launch_template_id` - (Optional) The ID of the launch template.
* `launch_template_name` - (Optional) The name of the launch template.
* `version` - (Required) The launch template version number, `$Latest`, or `$Default.`

#### override

Any parameters that you specify override the same parameters in the launch template. For fleets of type `request` and `maintain`, a maximum of 300 items is allowed across all launch templates.

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

The attributes for the instance types. For a list of currently supported values, please see ['InstanceRequirementsRequest'](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_InstanceRequirementsRequest.html).

This configuration block supports the following:

~> **NOTE:** Both `memory_mib.min` and `vcpu_count.min` must be specified.

* `accelerator_count` - (Optional) Block describing the minimum and maximum number of accelerators (GPUs, FPGAs, or AWS Inferentia chips). Default is no minimum or maximum limits.
    * `min` - (Optional) Minimum.
    * `max` - (Optional) Maximum. Set to `0` to exclude instance types with accelerators.
* `accelerator_manufacturers` - (Optional) List of accelerator manufacturer names. Default is any manufacturer.
* `accelerator_names` - (Optional) List of accelerator names. Default is any acclerator.
* `accelerator_total_memory_mib` - (Optional) Block describing the minimum and maximum total memory of the accelerators. Default is no minimum or maximum.
    * `min` - (Optional) The minimum amount of accelerator memory, in MiB. To specify no minimum limit, omit this parameter.
    * `max` - (Optional) The maximum amount of accelerator memory, in MiB. To specify no maximum limit, omit this parameter.
* `accelerator_types` - (Optional) The accelerator types that must be on the instance type. Default is any accelerator type.
* `allowed_instance_types` - (Optional) The instance types to apply your specified attributes against. All other instance types are ignored, even if they match your specified attributes. You can use strings with one or more wild cards,represented by an asterisk (\*). The following are examples: `c5*`, `m5a.*`, `r*`, `*3*`. For example, if you specify `c5*`, you are excluding the entire C5 instance family, which includes all C5a and C5n instance types. If you specify `m5a.*`, you are excluding all the M5a instance types, but not the M5n instance types. Maximum of 400 entries in the list; each entry is limited to 30 characters. Default is no excluded instance types. Default is any instance type.

    If you specify `AllowedInstanceTypes`, you can't specify `ExcludedInstanceTypes`.

* `bare_metal` - (Optional) Indicate whether bare metal instace types should be `included`, `excluded`, or `required`. Default is `excluded`.
* `baseline_ebs_bandwidth_mbps` - (Optional) Block describing the minimum and maximum baseline EBS bandwidth, in Mbps. Default is no minimum or maximum.
    * `min` - (Optional) The minimum baseline bandwidth, in Mbps. To specify no minimum limit, omit this parameter..
    * `max` - (Optional) The maximum baseline bandwidth, in Mbps. To specify no maximum limit, omit this parameter..
* `burstable_performance` - (Optional) Indicates whether burstable performance T instance types are `included`, `excluded`, or `required`. Default is `excluded`.
* `cpu_manufacturers` (Optional) The CPU manufacturers to include. Default is any manufacturer.
    ~> **NOTE:** Don't confuse the CPU hardware manufacturer with the CPU hardware architecture. Instances will be launched with a compatible CPU architecture based on the Amazon Machine Image (AMI) that you specify in your launch template.
* `excluded_instance_types` - (Optional) The instance types to exclude. You can use strings with one or more wild cards, represented by an asterisk (\*). The following are examples: `c5*`, `m5a.*`, `r*`, `*3*`. For example, if you specify `c5*`, you are excluding the entire C5 instance family, which includes all C5a and C5n instance types. If you specify `m5a.*`, you are excluding all the M5a instance types, but not the M5n instance types. Maximum of 400 entries in the list; each entry is limited to 30 characters. Default is no excluded instance types.

    If you specify `AllowedInstanceTypes`, you can't specify `ExcludedInstanceTypes`.

* `instance_generations` - (Optional) Indicates whether current or previous generation instance types are included. The current generation instance types are recommended for use. Valid values are `current` and `previous`. Default is `current` and `previous` generation instance types.
* `local_storage` - (Optional) Indicate whether instance types with local storage volumes are `included`, `excluded`, or `required`. Default is `included`.
* `local_storage_types` - (Optional) List of local storage type names. Valid values are `hdd` and `ssd`. Default any storage type.
* `max_spot_price_as_percentage_of_optimal_on_demand_price` - (Optional) The price protection threshold for Spot Instances. This is the maximum you’ll pay for a Spot Instance, expressed as a percentage higher than the cheapest M, C, or R instance type with your specified attributes. When Amazon EC2 Auto Scaling selects instance types with your attributes, we will exclude instance types whose price is higher than your threshold. The parameter accepts an integer, which Amazon EC2 Auto Scaling interprets as a percentage. To turn off price protection, specify a high value, such as 999999. Conflicts with `spot_max_price_percentage_over_lowest_price`
* `memory_gib_per_vcpu` - (Optional) Block describing the minimum and maximum amount of memory (GiB) per vCPU. Default is no minimum or maximum.
    * `min` - (Optional) The minimum amount of memory per vCPU, in GiB. To specify no minimum limit, omit this parameter.
    * `max` - (Optional) The maximum amount of memory per vCPU, in GiB. To specify no maximum limit, omit this parameter.
* `memory_mib` - (Required) The minimum and maximum amount of memory per vCPU, in GiB. Default is no minimum or maximum limits.
    * `min` - (Required) The minimum amount of memory, in MiB. To specify no minimum limit, specify `0`.
    * `max` - (Optional) The maximum amount of memory, in MiB. To specify no maximum limit, omit this parameter.
* `network_bandwidth_gbps` - (Optional) The minimum and maximum amount of network bandwidth, in gigabits per second (Gbps). Default is No minimum or maximum.
    * `min` - (Optional) The minimum amount of network bandwidth, in Gbps. To specify no minimum limit, omit this parameter.
    * `max` - (Optional) The maximum amount of network bandwidth, in Gbps. To specify no maximum limit, omit this parameter.
* `network_interface_count` - (Optional) Block describing the minimum and maximum number of network interfaces. Default is no minimum or maximum.
    * `min` - (Optional) The minimum number of network interfaces. To specify no minimum limit, omit this parameter.
    * `max` - (Optional) The maximum number of network interfaces. To specify no maximum limit, omit this parameter.
* `on_demand_max_price_percentage_over_lowest_price` - (Optional) The price protection threshold for On-Demand Instances. This is the maximum you’ll pay for an On-Demand Instance, expressed as a percentage higher than the cheapest M, C, or R instance type with your specified attributes. When Amazon EC2 Auto Scaling selects instance types with your attributes, we will exclude instance types whose price is higher than your threshold. The parameter accepts an integer, which Amazon EC2 Auto Scaling interprets as a percentage. To turn off price protection, specify a high value, such as 999999. Default is 20.

    If you set `target_capacity_unit_type` to `vcpu` or `memory-mib`, the price protection threshold is applied based on the per-vCPU or per-memory price instead of the per-instance price.

* `require_hibernate_support` - (Optional) Indicate whether instance types must support On-Demand Instance Hibernation, either `true` or `false`. Default is `false`.
* `spot_max_price_percentage_over_lowest_price` - (Optional) The price protection threshold for Spot Instances. This is the maximum you’ll pay for a Spot Instance, expressed as a percentage higher than the cheapest M, C, or R instance type with your specified attributes. When Amazon EC2 Auto Scaling selects instance types with your attributes, we will exclude instance types whose price is higher than your threshold. The parameter accepts an integer, which Amazon EC2 Auto Scaling interprets as a percentage. To turn off price protection, specify a high value, such as 999999. Default is 100. Conflicts with `max_spot_price_as_percentage_of_optimal_on_demand_price`

    If you set DesiredCapacityType to vcpu or memory-mib, the price protection threshold is applied based on the per vCPU or per memory price instead of the per instance price.

* `total_local_storage_gb` - (Optional) Block describing the minimum and maximum total local storage (GB). Default is no minimum or maximum.
    * `min` - (Optional) The minimum amount of total local storage, in GB. To specify no minimum limit, omit this parameter.
    * `max` - (Optional) The maximum amount of total local storage, in GB. To specify no maximum limit, omit this parameter.
* `vcpu_count` - (Required) Block describing the minimum and maximum number of vCPUs. Default is no maximum.
    * `min` - (Required) The minimum number of vCPUs. To specify no minimum limit, specify `0`.
    * `max` - (Optional) The maximum number of vCPUs. To specify no maximum limit, omit this parameter.

### on_demand_options

* `allocation_strategy` - (Optional) The order of the launch template overrides to use in fulfilling On-Demand capacity. Valid values: `lowestPrice`, `prioritized`. Default: `lowestPrice`.
* `capacity_reservation_options` (Optional) The strategy for using unused Capacity Reservations for fulfilling On-Demand capacity. Supported only for fleets of type `instant`.
    * `usage_strategy` - (Optional) Indicates whether to use unused Capacity Reservations for fulfilling On-Demand capacity. Valid values: `use-capacity-reservations-first`.
* `max_total_price` - (Optional) The maximum amount per hour for On-Demand Instances that you're willing to pay.
* `min_target_capacity` - (Optional) The minimum target capacity for On-Demand Instances in the fleet. If the minimum target capacity is not reached, the fleet launches no instances. Supported only for fleets of type `instant`.
    If you specify `min_target_capacity`, at least one of the following must be specified: `single_availability_zone` or `single_instance_type`.

* `single_availability_zone` - (Optional) Indicates that the fleet launches all On-Demand Instances into a single Availability Zone. Supported only for fleets of type `instant`.
* `single_instance_type` - (Optional) Indicates that the fleet uses a single instance type to launch all On-Demand Instances in the fleet. Supported only for fleets of type `instant`.

### spot_options

* `allocation_strategy` - (Optional) How to allocate the target capacity across the Spot pools. Valid values: `diversified`, `lowestPrice`, `capacity-optimized`, `capacity-optimized-prioritized` and `price-capacity-optimized`. Default: `lowestPrice`.
* `instance_interruption_behavior` - (Optional) Behavior when a Spot Instance is interrupted. Valid values: `hibernate`, `stop`, `terminate`. Default: `terminate`.
* `instance_pools_to_use_count` - (Optional) Number of Spot pools across which to allocate your target Spot capacity. Valid only when Spot `allocation_strategy` is set to `lowestPrice`. Default: `1`.
* `maintenance_strategies` - (Optional) Nested argument containing maintenance strategies for managing your Spot Instances that are at an elevated risk of being interrupted. Defined below.
* `max_total_price` - (Optional) The maximum amount per hour for Spot Instances that you're willing to pay.
* `min_target_capacity` - (Optional) The minimum target capacity for Spot Instances in the fleet. If the minimum target capacity is not reached, the fleet launches no instances. Supported only for fleets of type `instant`.
* `single_availability_zone` - (Optional) Indicates that the fleet launches all Spot Instances into a single Availability Zone. Supported only for fleets of type `instant`.
* `single_instance_type` - (Optional) Indicates that the fleet uses a single instance type to launch all Spot Instances in the fleet. Supported only for fleets of type `instant`.

### maintenance_strategies

* `capacity_rebalance` - (Optional) Nested argument containing the capacity rebalance for your fleet request. Defined below.

### capacity_rebalance

* `replacement_strategy` - (Optional) The replacement strategy to use. Only available for fleets of `type` set to `maintain`. Valid values: `launch`.

### target_capacity_specification

* `default_target_capacity_type` - (Required) Default target capacity type. Valid values: `on-demand`, `spot`.
* `on_demand_target_capacity` - (Optional) The number of On-Demand units to request.
* `spot_target_capacity` - (Optional) The number of Spot units to request.
* `target_capacity_unit_type` - (Optional) The unit for the target capacity.
    If you specify `target_capacity_unit_type`, `instance_requirements` must be specified.

* `total_target_capacity` - (Required) The number of units to request, filled using `default_target_capacity_type`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Fleet identifier
* `arn` - The ARN of the fleet
* `fleet_instance_set` - Information about the instances that were launched by the fleet. Available only when `type` is set to `instant`.
    * `instance_ids` - The IDs of the instances.
    * `instance_type` - The instance type.
    * `lifecycle` - Indicates if the instance that was launched is a Spot Instance or On-Demand Instance.
    * `platform` - The value is `Windows` for Windows instances. Otherwise, the value is blank.
* `fleet_state` - The state of the EC2 Fleet.
* `fulfilled_capacity` - The number of units fulfilled by this request compared to the set target capacity.
* `fulfilled_on_demand_capacity` - The number of units fulfilled by this request compared to the set target On-Demand capacity.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_fleet` using the Fleet identifier. For example:

```terraform
import {
  to = aws_ec2_fleet.example
  id = "fleet-b9b55d27-c5fc-41ac-a6f3-48fcc91f080c"
}
```

Using `terraform import`, import `aws_ec2_fleet` using the Fleet identifier. For example:

```console
% terraform import aws_ec2_fleet.example fleet-b9b55d27-c5fc-41ac-a6f3-48fcc91f080c
```
