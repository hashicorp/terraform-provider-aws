// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2InstanceTypeDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_instance_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeDataSourceConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "auto_recovery_supported", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "bandwidth_weightings.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "bare_metal", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "boot_modes.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "boot_modes.0", "legacy-bios"),
					resource.TestCheckResourceAttr(dataSourceName, "boot_modes.1", "uefi"),
					resource.TestCheckResourceAttr(dataSourceName, "burstable_performance_supported", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "current_generation", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "dedicated_hosts_supported", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "default_cores", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "default_network_card_index", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "default_threads_per_core", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "default_vcpus", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_encryption_support", "supported"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_nvme_support", "required"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_optimized_support", "default"),
					resource.TestCheckResourceAttr(dataSourceName, "efa_supported", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "ena_srd_supported", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "ena_support", "required"),
					resource.TestCheckResourceAttr(dataSourceName, "encryption_in_transit_supported", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "free_tier_eligible", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "hibernation_supported", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "hypervisor", "nitro"),
					resource.TestCheckResourceAttr(dataSourceName, "instance_storage_supported", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrInstanceType, "m5.large"),
					resource.TestCheckResourceAttr(dataSourceName, "ipv6_supported", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "maximum_ipv4_addresses_per_interface", "10"),
					resource.TestCheckResourceAttr(dataSourceName, "maximum_ipv6_addresses_per_interface", "10"),
					resource.TestCheckResourceAttr(dataSourceName, "maximum_network_cards", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "maximum_network_interfaces", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "media_accelerators.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "memory_size", "8192"),
					resource.TestCheckResourceAttr(dataSourceName, "network_cards.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "network_cards.0.baseline_bandwidth", "0.75"),
					resource.TestCheckResourceAttr(dataSourceName, "network_cards.0.index", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "network_cards.0.maximum_interfaces", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "network_cards.0.performance", "Up to 10 Gigabit"),
					resource.TestCheckResourceAttr(dataSourceName, "network_cards.0.peak_bandwidth", "10"),
					resource.TestCheckResourceAttr(dataSourceName, "network_performance", "Up to 10 Gigabit"),
					resource.TestCheckResourceAttr(dataSourceName, "neuron_devices.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "nitro_enclaves_support", "unsupported"),
					resource.TestCheckResourceAttr(dataSourceName, "nitro_tpm_support", "supported"),
					resource.TestCheckResourceAttr(dataSourceName, "nitro_tpm_supported_versions.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "nitro_tpm_supported_versions.0", "2.0"),
					resource.TestCheckResourceAttr(dataSourceName, "phc_support", "unsupported"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_architectures.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_architectures.0", "x86_64"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_cpu_features.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_placement_strategies.#", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_placement_strategies.0", "cluster"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_placement_strategies.1", "partition"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_placement_strategies.2", "spread"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_root_device_types.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_root_device_types.0", "ebs"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_usages_classes.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_usages_classes.0", "on-demand"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_usages_classes.1", "spot"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_virtualization_types.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_virtualization_types.0", "hvm"),
					resource.TestCheckResourceAttr(dataSourceName, "sustained_clock_speed", "3.1"),
					resource.TestCheckResourceAttr(dataSourceName, "valid_cores.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "valid_cores.0", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "valid_threads_per_core.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "valid_threads_per_core.0", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "valid_threads_per_core.1", "2"),
				),
			},
		},
	})
}

func TestAccEC2InstanceTypeDataSource_metal(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_instance_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeDataSourceConfig_metal,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrInstanceType, "i3en.metal"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_performance_baseline_bandwidth", "19000"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_performance_baseline_throughput", "2375"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_performance_baseline_iops", "80000"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_performance_maximum_bandwidth", "19000"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_performance_maximum_throughput", "2375"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_performance_maximum_iops", "80000"),
					resource.TestCheckResourceAttr(dataSourceName, "instance_disks.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "instance_disks.0.count", "8"),
					resource.TestCheckResourceAttr(dataSourceName, "instance_disks.0.size", "7500"),
					resource.TestCheckResourceAttr(dataSourceName, "instance_disks.0.type", "ssd"),
					resource.TestCheckResourceAttr(dataSourceName, "total_instance_storage", "60000"),
				),
			},
		},
	})
}

func TestAccEC2InstanceTypeDataSource_gpu(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_instance_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeDataSourceConfig_gpu,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrInstanceType, "p5.48xlarge"),
					resource.TestCheckResourceAttr(dataSourceName, "efa_maximum_interfaces", "32"),
					resource.TestCheckResourceAttr(dataSourceName, "efa_supported", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "gpus.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "gpus.0.count", "8"),
					resource.TestCheckResourceAttr(dataSourceName, "gpus.0.manufacturer", "NVIDIA"),
					resource.TestCheckResourceAttr(dataSourceName, "gpus.0.memory_size", "81920"),
					resource.TestCheckResourceAttr(dataSourceName, "gpus.0.name", "H100"),
					resource.TestCheckResourceAttr(dataSourceName, "maximum_network_cards", "32"),
					resource.TestCheckResourceAttr(dataSourceName, "maximum_network_interfaces", "64"),
				),
			},
		},
	})
}

func TestAccEC2InstanceTypeDataSource_fpga(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_instance_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeDataSourceConfig_fgpa,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrInstanceType, "f1.2xlarge"),
					resource.TestCheckResourceAttr(dataSourceName, "fpgas.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "fpgas.0.count", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "fpgas.0.manufacturer", "Xilinx"),
					resource.TestCheckResourceAttr(dataSourceName, "fpgas.0.memory_size", "65536"),
					resource.TestCheckResourceAttr(dataSourceName, "fpgas.0.name", "Virtex UltraScale (VU9P)"),
					resource.TestCheckResourceAttr(dataSourceName, "total_fpga_memory", "65536"),
				),
			},
		},
	})
}

func TestAccEC2InstanceTypeDataSource_neuron(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_instance_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeDataSourceConfig_neuron,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrInstanceType, "inf1.xlarge"),
					resource.TestCheckResourceAttr(dataSourceName, "inference_accelerators.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "inference_accelerators.0.count", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "inference_accelerators.0.memory_size", "8192"),
					resource.TestCheckResourceAttr(dataSourceName, "inference_accelerators.0.manufacturer", "AWS"),
					resource.TestCheckResourceAttr(dataSourceName, "inference_accelerators.0.name", "Inferentia"),
					resource.TestCheckResourceAttr(dataSourceName, "neuron_devices.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "neuron_devices.0.core_count", "4"),
					resource.TestCheckResourceAttr(dataSourceName, "neuron_devices.0.core_version", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "neuron_devices.0.count", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "neuron_devices.0.memory_size", "8192"),
					resource.TestCheckResourceAttr(dataSourceName, "neuron_devices.0.name", "Inferentia"),
					resource.TestCheckResourceAttr(dataSourceName, "total_inference_memory", "8192"),
					resource.TestCheckResourceAttr(dataSourceName, "total_neuron_device_memory", "8192"),
				),
			},
		},
	})
}

func TestAccEC2InstanceTypeDataSource_media_accelerator(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_instance_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeDataSourceConfig_media_accelerator,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrInstanceType, "vt1.6xlarge"),
					resource.TestCheckResourceAttr(dataSourceName, "media_accelerators.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "media_accelerators.0.count", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "media_accelerators.0.manufacturer", "Xilinx"),
					resource.TestCheckResourceAttr(dataSourceName, "media_accelerators.0.memory_size", "24576"),
					resource.TestCheckResourceAttr(dataSourceName, "media_accelerators.0.name", "U30"),
					resource.TestCheckResourceAttr(dataSourceName, "total_media_memory", "49152"),
				),
			},
		},
	})
}

const testAccInstanceTypeDataSourceConfig_basic = `
data "aws_ec2_instance_type" "test" {
  instance_type = "m5.large"
}
`

const testAccInstanceTypeDataSourceConfig_metal = `
data "aws_ec2_instance_type" "test" {
  instance_type = "i3en.metal"
}
`

const testAccInstanceTypeDataSourceConfig_gpu = `
data "aws_ec2_instance_type" "test" {
  instance_type = "p5.48xlarge"
}
`

const testAccInstanceTypeDataSourceConfig_fgpa = `
data "aws_ec2_instance_type" "test" {
  instance_type = "f1.2xlarge"
}
`

const testAccInstanceTypeDataSourceConfig_neuron = `
data "aws_ec2_instance_type" "test" {
  instance_type = "inf1.xlarge"
}
`

const testAccInstanceTypeDataSourceConfig_media_accelerator = `
data "aws_ec2_instance_type" "test" {
  instance_type = "vt1.6xlarge"
}
`
