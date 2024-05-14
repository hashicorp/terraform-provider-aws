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
					resource.TestCheckResourceAttr(dataSourceName, "auto_recovery_supported", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "bare_metal", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "burstable_performance_supported", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "current_generation", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "dedicated_hosts_supported", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "default_cores", acctest.CtOne),
					resource.TestCheckResourceAttr(dataSourceName, "default_threads_per_core", acctest.CtTwo),
					resource.TestCheckResourceAttr(dataSourceName, "default_vcpus", acctest.CtTwo),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_encryption_support", "supported"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_nvme_support", "required"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_optimized_support", "default"),
					resource.TestCheckResourceAttr(dataSourceName, "efa_supported", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "ena_support", "required"),
					resource.TestCheckResourceAttr(dataSourceName, "encryption_in_transit_supported", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "free_tier_eligible", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "hibernation_supported", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "hypervisor", "nitro"),
					resource.TestCheckResourceAttr(dataSourceName, "instance_storage_supported", "false"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrInstanceType, "m5.large"),
					resource.TestCheckResourceAttr(dataSourceName, "ipv6_supported", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "maximum_ipv4_addresses_per_interface", "10"),
					resource.TestCheckResourceAttr(dataSourceName, "maximum_ipv6_addresses_per_interface", "10"),
					resource.TestCheckResourceAttr(dataSourceName, "maximum_network_cards", acctest.CtOne),
					resource.TestCheckResourceAttr(dataSourceName, "maximum_network_interfaces", acctest.CtThree),
					resource.TestCheckResourceAttr(dataSourceName, "memory_size", "8192"),
					resource.TestCheckResourceAttr(dataSourceName, "network_performance", "Up to 10 Gigabit"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_architectures.#", acctest.CtOne),
					resource.TestCheckResourceAttr(dataSourceName, "supported_architectures.0", "x86_64"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_placement_strategies.#", acctest.CtThree),
					resource.TestCheckResourceAttr(dataSourceName, "supported_placement_strategies.0", "cluster"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_placement_strategies.1", "partition"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_placement_strategies.2", "spread"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_root_device_types.#", acctest.CtOne),
					resource.TestCheckResourceAttr(dataSourceName, "supported_root_device_types.0", "ebs"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_usages_classes.#", acctest.CtTwo),
					resource.TestCheckResourceAttr(dataSourceName, "supported_usages_classes.0", "on-demand"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_usages_classes.1", "spot"),
					resource.TestCheckResourceAttr(dataSourceName, "supported_virtualization_types.#", acctest.CtOne),
					resource.TestCheckResourceAttr(dataSourceName, "supported_virtualization_types.0", "hvm"),
					resource.TestCheckResourceAttr(dataSourceName, "sustained_clock_speed", "3.1"),
					resource.TestCheckResourceAttr(dataSourceName, "valid_cores.#", acctest.CtOne),
					resource.TestCheckResourceAttr(dataSourceName, "valid_cores.0", acctest.CtOne),
					resource.TestCheckResourceAttr(dataSourceName, "valid_threads_per_core.#", acctest.CtTwo),
					resource.TestCheckResourceAttr(dataSourceName, "valid_threads_per_core.0", acctest.CtOne),
					resource.TestCheckResourceAttr(dataSourceName, "valid_threads_per_core.1", acctest.CtTwo),
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
					resource.TestCheckResourceAttr(dataSourceName, "ebs_performance_baseline_bandwidth", "19000"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_performance_baseline_throughput", "2375"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_performance_baseline_iops", "80000"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_performance_maximum_bandwidth", "19000"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_performance_maximum_throughput", "2375"),
					resource.TestCheckResourceAttr(dataSourceName, "ebs_performance_maximum_iops", "80000"),
					resource.TestCheckResourceAttr(dataSourceName, "instance_disks.#", acctest.CtOne),
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
					resource.TestCheckResourceAttr(dataSourceName, "gpus.#", acctest.CtOne),
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
					resource.TestCheckResourceAttr(dataSourceName, "fpgas.#", acctest.CtOne),
					resource.TestCheckResourceAttr(dataSourceName, "fpgas.0.count", acctest.CtOne),
					resource.TestCheckResourceAttr(dataSourceName, "fpgas.0.manufacturer", "Xilinx"),
					resource.TestCheckResourceAttr(dataSourceName, "fpgas.0.memory_size", "65536"),
					resource.TestCheckResourceAttr(dataSourceName, "fpgas.0.name", "Virtex UltraScale (VU9P)"),
					resource.TestCheckResourceAttr(dataSourceName, "total_fpga_memory", "65536"),
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
