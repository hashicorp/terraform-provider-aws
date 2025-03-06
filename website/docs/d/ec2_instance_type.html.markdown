---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_instance_type"
description: |-
  Information about single EC2 Instance Type.
---


# Data Source: aws_ec2_instance_type

Get characteristics for a single EC2 Instance Type.

## Example Usage

```terraform
data "aws_ec2_instance_type" "example" {
  instance_type = "t2.micro"
}

```

## Argument Reference

This data source supports the following arguments:

* `instance_type` - (Required) Instance

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

~> **NOTE:** Not all attributes are set for every instance type.

* `auto_recovery_supported` - `true` if auto recovery is supported.
* `bandwidth_weightings` - A set of strings of valid settings for [configurable bandwidth weighting](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configure-bandwidth-weighting.html), if supported.
* `bare_metal` - `true` if it is a bare metal instance type.
* `boot_modes` - A set of strings of supported boot modes.
* `burstable_performance_supported` - `true` if the instance type is a burstable performance instance type.
* `current_generation` - `true`  if the instance type is a current generation.
* `dedicated_hosts_supported` - `true` if Dedicated Hosts are supported on the instance type.
* `default_cores` - Default number of cores for the instance type.
* `default_network_card_index` - The index of the default network card, starting at `0`.
* `default_threads_per_core` - The  default  number of threads per core for the instance type.
* `default_vcpus` - Default number of vCPUs for the instance type.
* `ebs_encryption_support` - Indicates whether Amazon EBS encryption is supported.
* `ebs_nvme_support` - Whether non-volatile memory express (NVMe) is supported.
* `ebs_optimized_support` - Indicates that the instance type is Amazon EBS-optimized.
* `ebs_performance_baseline_bandwidth` - The baseline bandwidth performance for an EBS-optimized instance type, in Mbps.
* `ebs_performance_baseline_iops` - The baseline input/output storage operations per seconds for an EBS-optimized instance type.
* `ebs_performance_baseline_throughput` - The baseline throughput performance for an EBS-optimized instance type, in MBps.
* `ebs_performance_maximum_bandwidth` - The maximum bandwidth performance for an EBS-optimized instance type, in Mbps.
* `ebs_performance_maximum_iops` - The maximum input/output storage operations per second for an EBS-optimized instance type.
* `ebs_performance_maximum_throughput` - The maximum throughput performance for an EBS-optimized instance type, in MBps.
* `efa_maximum_interfaces` - The maximum number of Elastic Fabric Adapters for the instance type.
* `efa_supported` - `true` if Elastic Fabric Adapter (EFA) is supported.
* `ena_srd_supported` - `true` if the instance type supports [ENA Express](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ena-express.html).
* `ena_support` - Indicates whether Elastic Network Adapter (ENA) is `"supported"`, `"required"`, or `"unsupported"`.
* `encryption_in_transit_supported` - `true` if encryption in-transit between instances is supported.
* `fpgas` - Describes the FPGA accelerator settings for the instance type.
    * `fpgas.#.count` - The count of FPGA accelerators for the instance type.
    * `fpgas.#.manufacturer` - The manufacturer of the FPGA accelerator.
    * `fpgas.#.memory_size` - The size (in MiB) of the memory available to the FPGA accelerator.
    * `fpgas.#.name` - The name of the FPGA accelerator.
* `free_tier_eligible` - `true` if the instance type is eligible for the free tier.
* `gpus` - Describes the GPU accelerators for the instance type.
    * `gpus.#.count` - The number of GPUs for the instance type.
    * `gpus.#.manufacturer` - The manufacturer of the GPU accelerator.
    * `gpus.#.memory_size` - The size (in MiB) of the memory available to the GPU accelerator.
    * `gpus.#.name` - The name of the GPU accelerator.
* `hibernation_supported` - `true` if On-Demand hibernation is supported.
* `hypervisor` - Hypervisor used for the instance type.
* `inference_accelerators` Describes the Inference accelerators for the instance type.
    * `inference_accelerators.#.count` - The number of Inference accelerators for the instance type.
    * `inference_accelerators.#.manufacturer` - The manufacturer of the Inference accelerator.
    * `inference_accelerators.#.memory_size` - The size (in MiB) of the memory available to the inference accelerator.
    * `inference_accelerators.#.name` - The name of the Inference accelerator.
* `instance_disks` - Describes the disks for the instance type.
    * `instance_disks.#.count` - The number of disks with this configuration.
    * `instance_disks.#.size` - The size of the disk in GB.
    * `instance_disks.#.type` - The type of disk.
* `instance_storage_supported` - `true` if instance storage is supported.
* `ipv6_supported` - `true` if IPv6 is supported.
* `maximum_ipv4_addresses_per_interface` - The maximum number of IPv4 addresses per network interface.
* `maximum_ipv6_addresses_per_interface` - The maximum number of IPv6 addresses per network interface.
* `maximum_network_cards` - The maximum number of physical network cards that can be allocated to the instance.
* `maximum_network_interfaces` - The maximum number of network interfaces for the instance type.
* `media_accelerators` -  Describes the media accelerator settings for the instance type.
    * `media_accelerators.#.count` - The number of media accelerators for the instance type.
    * `media_accelerators.#.manufacturer` - The manufacturer of the media accelerator.
    * `media_accelerators.#.memory_size` - The size (in MiB) of the memory available to each media accelerator.
    * `media_accelerators.#.name` - The name of the media accelerator.
* `memory_size` - Size of the instance memory, in MiB.
* `network_cards` - Describes the network cards for the instance type.
    * `network_cards.#.baseline_bandwidth` - The baseline network performance (in Gbps) of the network card.
    * `network_cards.#.index` - The index of the network card.
    * `network_cards.#.maximum_interfaces` - The maximum number of network interfaces for the /network card.
    * `network_cards.#.performance` - Describes the network performance of the network card.
    * `network_cards.#.peak_bandwidth` - The peak (burst) network performance (in Gbps) of the network card.
* `network_performance` - Describes the network performance.
* `neuron_devices` - Describes the Neuron accelerator settings for the instance type.
    * `neuron_devices.#.core_count` - The number of cores available to the neuron accelerator.
    * `neuron_devices.#.core_version` - A number representing the version of the neuron accelerator.
    * `neuron_devices.#.count` - The number of neuron accelerators for the instance type.
    * `neuron_devices.#.memory_size` - The size (in MiB) of the memory available to the neuron accelerator.
    * `neuron_devices.#.name` - The name of the neuron accelerator.
* `nitro_enclaves_support` - Indicates whether Nitro Enclaves is `"supported"` or `"unsupported"`.
* `nitro_tpm_support` - Indicates whether NitroTPM is `"supported"` or `"unsupported"`.
* `nitro_tpm_supported_versions` - A set of strings indicating the supported NitroTPM versions.
* `phc_support` - `true` if a local Precision Time Protocol (PTP) hardware clock (PHC) is supported.
* `supported_architectures` - A list of strings of architectures supported by the instance type.
* `supported_cpu_features` - A set of strings indicating supported CPU features.
* `supported_placement_strategies` - A list of supported placement groups types.
* `supported_root_device_types` - A list of supported root device types.
* `supported_usages_classes` - A list of supported usage classes.  Usage classes are `"spot"`, `"on-demand"`, or `"capacity-block"`.
* `supported_virtualization_types` - The supported virtualization types.
* `sustained_clock_speed` - The speed of the processor, in GHz.
* `total_fpga_memory` - Total memory of all FPGA accelerators for the instance type (in MiB).
* `total_gpu_memory` - Total size of the memory for the GPU accelerators for the instance type (in MiB).
* `total_inference_memory` - The total size of the memory for the neuron accelerators for the instance type (in MiB).
* `total_instance_storage` - The total size of the instance disks, in GB.
* `total_neuron_device_memory` - The total size of the memory for the neuron accelerators for the instance type (in MiB).
* `total_media_memory` - The total size of the memory for the media accelerators for the instance type (in MiB).
* `valid_cores` - List of the valid number of cores that can be configured for the instance type.
* `valid_threads_per_core` - List of the valid number of threads per core that can be configured for the instance type.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
