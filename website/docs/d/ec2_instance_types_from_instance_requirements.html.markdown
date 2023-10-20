---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_instance_types_from_instance_requirements"
description: |-
  EC2 Instance Types that match requirements.
---

# Data Source: aws_ec2_instance_types_from_instance_requirements

Get a list of EC2 Instance Types that match requirements.

## Example Usage

### Basic Usage

```terraform
data "aws_ec2_instance_types_from_instance_requirements" "example" {
  architecture_types   = ["x86_64"]
  virtualization_types = ["hvm"]

  instance_requirements {
    memory_mib {
      min = 1024
    }
    vcpu_count {
      min = 2
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `architecture_types` - (Required) List of architecture types allowed. Possible values are `i386`, `x86_64`, `arm64`, `x86_64_mac` and `arm64_mac`
* `instance_requirements` - (Required) The attribute requirements for the type of instance.
* `virtualization_types` - (Required) List of virtualization types allowed. Possible values are `hvm` and `paravirtual`

### instance_requirements

This configuration block supports the following:

~> **NOTE:** Both `memory_mib.min` and `vcpu_count.min` must be specified.

- `accelerator_count` - (Optional) Block describing the minimum and maximum number of accelerators (GPUs, FPGAs, or AWS Inferentia chips). Default is no minimum or maximum.
    - `min` - (Optional) Minimum.
    - `max` - (Optional) Maximum. Set to `0` to exclude instance types with accelerators.
- `accelerator_manufacturers` - (Optional) List of accelerator manufacturer names. Default is any manufacturer.

  ```
  Valid names:
    * amazon-web-services
    * amd
    * nvidia
    * xilinx
  ```

- `accelerator_names` - (Optional) List of accelerator names. Default is any acclerator.

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

- `accelerator_total_memory_mib` - (Optional) Block describing the minimum and maximum total memory of the accelerators. Default is no minimum or maximum.

    - `min` - (Optional) Minimum.
    - `max` - (Optional) Maximum.

- `accelerator_types` - (Optional) List of accelerator types. Default is any accelerator type.

  ```
  Valid types:
    * fpga
    * gpu
    * inference
  ```

- `allowed_instance_types` - (Optional) List of instance types to apply your specified attributes against. All other instance types are ignored, even if they match your specified attributes. You can use strings with one or more wild cards, represented by an asterisk (\*), to allow an instance type, size, or generation. The following are examples: `m5.8xlarge`, `c5*.*`, `m5a.*`, `r*`, `*3*`. For example, if you specify `c5*`, you are allowing the entire C5 instance family, which includes all C5a and C5n instance types. If you specify `m5a.*`, you are allowing all the M5a instance types, but not the M5n instance types. Maximum of 400 entries in the list; each entry is limited to 30 characters. Default is all instance types.

  ~> **NOTE:** If you specify `allowed_instance_types`, you can't specify `excluded_instance_types`.

- `bare_metal` - (Optional) Indicate whether bare metal instace types should be `included`, `excluded`, or `required`. Default is `excluded`.
- `baseline_ebs_bandwidth_mbps` - (Optional) Block describing the minimum and maximum baseline EBS bandwidth, in Mbps. Default is no minimum or maximum.
    - `min` - (Optional) Minimum.
    - `max` - (Optional) Maximum.
- `burstable_performance` - (Optional) Indicate whether burstable performance instance types should be `included`, `excluded`, or `required`. Default is `excluded`.
- `cpu_manufacturers` (Optional) List of CPU manufacturer names. Default is any manufacturer.

  ~> **NOTE:** Don't confuse the CPU hardware manufacturer with the CPU hardware architecture. Instances will be launched with a compatible CPU architecture based on the Amazon Machine Image (AMI) that you specify in your launch template.

  ```
  Valid names:
    * amazon-web-services
    * amd
    * intel
  ```

- `excluded_instance_types` - (Optional) List of instance types to exclude. You can use strings with one or more wild cards, represented by an asterisk (\*), to exclude an instance type, size, or generation. The following are examples: `m5.8xlarge`, `c5*.*`, `m5a.*`, `r*`, `*3*`. For example, if you specify `c5*`, you are excluding the entire C5 instance family, which includes all C5a and C5n instance types. If you specify `m5a.*`, you are excluding all the M5a instance types, but not the M5n instance types. Maximum of 400 entries in the list; each entry is limited to 30 characters. Default is no excluded instance types.

  ~> **NOTE:** If you specify `excluded_instance_types`, you can't specify `allowed_instance_types`.

- `instance_generations` - (Optional) List of instance generation names. Default is any generation.

  ```
  Valid names:
    * current  - Recommended for best performance.
    * previous - For existing applications optimized for older instance types.
  ```

- `local_storage` - (Optional) Indicate whether instance types with local storage volumes are `included`, `excluded`, or `required`. Default is `included`.
- `local_storage_types` - (Optional) List of local storage type names. Default any storage type.

  ```
  Value names:
    * hdd - hard disk drive
    * ssd - solid state drive
  ```

- `memory_gib_per_vcpu` - (Optional) Block describing the minimum and maximum amount of memory (GiB) per vCPU. Default is no minimum or maximum.
    - `min` - (Optional) Minimum. May be a decimal number, e.g. `0.5`.
    - `max` - (Optional) Maximum. May be a decimal number, e.g. `0.5`.
- `memory_mib` - (Required) Block describing the minimum and maximum amount of memory (MiB). Default is no maximum.
    - `min` - (Required) Minimum.
    - `max` - (Optional) Maximum.
- `network_bandwidth_gbps` - (Optional) Block describing the minimum and maximum amount of network bandwidth, in gigabits per second (Gbps). Default is no minimum or maximum.
    - `min` - (Optional) Minimum.
    - `max` - (Optional) Maximum.
- `network_interface_count` - (Optional) Block describing the minimum and maximum number of network interfaces. Default is no minimum or maximum.
    - `min` - (Optional) Minimum.
    - `max` - (Optional) Maximum.
- `on_demand_max_price_percentage_over_lowest_price` - (Optional) Price protection threshold for On-Demand Instances. This is the maximum you’ll pay for an On-Demand Instance, expressed as a percentage higher than the cheapest M, C, or R instance type with your specified attributes. When Amazon EC2 Auto Scaling selects instance types with your attributes, we will exclude instance types whose price is higher than your threshold. The parameter accepts an integer, which Amazon EC2 Auto Scaling interprets as a percentage. To turn off price protection, specify a high value, such as 999999. Default is 20.

  If you set DesiredCapacityType to vcpu or memory-mib, the price protection threshold is applied based on the per vCPU or per memory price instead of the per instance price.

- `require_hibernate_support` - (Optional) Indicate whether instance types must support On-Demand Instance Hibernation, either `true` or `false`. Default is `false`.
- `spot_max_price_percentage_over_lowest_price` - (Optional) Price protection threshold for Spot Instances. This is the maximum you’ll pay for a Spot Instance, expressed as a percentage higher than the cheapest M, C, or R instance type with your specified attributes. When Amazon EC2 Auto Scaling selects instance types with your attributes, we will exclude instance types whose price is higher than your threshold. The parameter accepts an integer, which Amazon EC2 Auto Scaling interprets as a percentage. To turn off price protection, specify a high value, such as 999999. Default is 100.

  If you set DesiredCapacityType to vcpu or memory-mib, the price protection threshold is applied based on the per vCPU or per memory price instead of the per instance price.

- `total_local_storage_gb` - (Optional) Block describing the minimum and maximum total local storage (GB). Default is no minimum or maximum.
    - `min` - (Optional) Minimum. May be a decimal number, e.g. `0.5`.
    - `max` - (Optional) Maximum. May be a decimal number, e.g. `0.5`.
- `vcpu_count` - (Required) Block describing the minimum and maximum number of vCPUs. Default is no maximum.
    - `min` - (Required) Minimum.
    - `max` - (Optional) Maximum.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `instance_types` - List of EC2 Instance Types.
