package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsEc2InstanceType_attributes(t *testing.T) {
	resourceMetal := "data.aws_ec2_instance_type.metal"
	resourceGpu := "data.aws_ec2_instance_type.gpu"
	resourceFpga := "data.aws_ec2_instance_type.fpga"
	resourceAccelerator := "data.aws_ec2_instance_type.accelerator"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceEc2InstanceType,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceMetal, "auto_recovery_supported", "false"),
					resource.TestCheckResourceAttr(resourceMetal, "bare_metal", "true"),
					resource.TestCheckResourceAttr(resourceMetal, "burstable_performance_supported", "false"),
					resource.TestCheckResourceAttr(resourceMetal, "current_generation", "true"),
					resource.TestCheckResourceAttr(resourceMetal, "dedicated_hosts_supported", "true"),
					resource.TestCheckResourceAttr(resourceMetal, "default_vcpus", "96"),
					resource.TestCheckResourceAttr(resourceMetal, "ebs_encryption_support", "supported"),
					resource.TestCheckResourceAttr(resourceMetal, "ebs_optimized_support", "default"),
					resource.TestCheckResourceAttr(resourceMetal, "ena_support", "required"),
					resource.TestCheckResourceAttr(resourceMetal, "free_tier_eligible", "false"),
					resource.TestCheckResourceAttr(resourceMetal, "hibernation_supported", "false"),
					resource.TestCheckResourceAttr(resourceMetal, "instance_storage_supported", "true"),
					resource.TestCheckResourceAttr(resourceMetal, "instance_type", "i3en.metal"),
					resource.TestCheckResourceAttr(resourceMetal, "ipv6_supported", "true"),
					resource.TestCheckResourceAttr(resourceMetal, "maximum_ipv4_addresses_per_interface", "50"),
					resource.TestCheckResourceAttr(resourceMetal, "maximum_ipv6_addresses_per_interface", "50"),
					resource.TestCheckResourceAttr(resourceMetal, "maximum_network_interfaces", "15"),
					resource.TestCheckResourceAttr(resourceMetal, "memory_size", "786432"),
					resource.TestCheckResourceAttr(resourceMetal, "network_performance", "100 Gigabit"),
					resource.TestCheckResourceAttr(resourceMetal, "supported_architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceMetal, "supported_architectures.0", "x86_64"),
					resource.TestCheckResourceAttr(resourceMetal, "supported_placement_strategies.#", "3"),
					resource.TestCheckResourceAttr(resourceMetal, "supported_placement_strategies.0", "cluster"),
					resource.TestCheckResourceAttr(resourceMetal, "supported_placement_strategies.1", "partition"),
					resource.TestCheckResourceAttr(resourceMetal, "supported_placement_strategies.2", "spread"),
					resource.TestCheckResourceAttr(resourceMetal, "supported_root_device_types.#", "1"),
					resource.TestCheckResourceAttr(resourceMetal, "supported_root_device_types.0", "ebs"),
					resource.TestCheckResourceAttr(resourceMetal, "supported_usages_classes.#", "2"),
					resource.TestCheckResourceAttr(resourceMetal, "supported_usages_classes.0", "on-demand"),
					resource.TestCheckResourceAttr(resourceMetal, "supported_usages_classes.1", "spot"),
					resource.TestCheckResourceAttr(resourceMetal, "sustained_clock_speed", "3.1"),
					resource.TestCheckResourceAttr(resourceMetal, "total_instance_storage", "60000"),
					resource.TestCheckResourceAttr(resourceMetal, "instance_disks.#", "1"),
					resource.TestCheckResourceAttr(resourceMetal, "instance_disks.0.count", "8"),
					resource.TestCheckResourceAttr(resourceMetal, "instance_disks.0.size", "7500"),
					resource.TestCheckResourceAttr(resourceMetal, "instance_disks.0.type", "ssd"),
					resource.TestCheckResourceAttr(resourceGpu, "total_gpu_memory", "4096"),
					resource.TestCheckResourceAttr(resourceGpu, "hypervisor", "xen"),
					resource.TestCheckResourceAttr(resourceGpu, "gpus.#", "1"),
					resource.TestCheckResourceAttr(resourceGpu, "gpus.0.count", "1"),
					resource.TestCheckResourceAttr(resourceGpu, "gpus.0.memory_size", "4096"),
					resource.TestCheckResourceAttr(resourceGpu, "gpus.0.manufacturer", "NVIDIA"),
					resource.TestCheckResourceAttr(resourceGpu, "gpus.0.name", "K520"),
					resource.TestCheckResourceAttr(resourceGpu, "valid_threads_per_core.#", "2"),
					resource.TestCheckResourceAttr(resourceGpu, "valid_threads_per_core.0", "1"),
					resource.TestCheckResourceAttr(resourceGpu, "valid_threads_per_core.1", "2"),
					resource.TestCheckResourceAttr(resourceGpu, "default_threads_per_core", "2"),
					resource.TestCheckResourceAttr(resourceGpu, "default_cores", "4"),
					resource.TestCheckResourceAttr(resourceGpu, "default_vcpus", "8"),
					resource.TestCheckResourceAttr(resourceFpga, "fpgas.#", "1"),
					resource.TestCheckResourceAttr(resourceFpga, "fpgas.0.name", "Virtex UltraScale (VU9P)"),
					resource.TestCheckResourceAttr(resourceFpga, "fpgas.0.manufacturer", "Xilinx"),
					resource.TestCheckResourceAttr(resourceFpga, "fpgas.0.count", "1"),
					resource.TestCheckResourceAttr(resourceFpga, "fpgas.0.memory_size", "65536"),
					resource.TestCheckResourceAttr(resourceFpga, "total_fpga_memory", "65536"),
					resource.TestCheckResourceAttr(resourceAccelerator, "accelerators.#", "1"),
					resource.TestCheckResourceAttr(resourceAccelerator, "accelerators.0.count", "1"),
					resource.TestCheckResourceAttr(resourceAccelerator, "accelerators.0.name", "Inferentia"),
					resource.TestCheckResourceAttr(resourceAccelerator, "accelerators.0.manufacturer", "AWS"),
					resource.TestCheckResourceAttr(resourceAccelerator, "valid_cores.#", "1"),
					resource.TestCheckResourceAttr(resourceAccelerator, "valid_cores.0", "2"),
				),
			},
		},
	})
}

const testAccDataSourceEc2InstanceType = `
data "aws_ec2_instance_type" "metal" {
	instance_type="i3en.metal"
}
data "aws_ec2_instance_type" "gpu" {
	instance_type="g2.2xlarge"
}
data "aws_ec2_instance_type" "fpga" {
	instance_type="f1.2xlarge"
}
data "aws_ec2_instance_type" "accelerator" {
	instance_type="inf1.xlarge"
}
`
