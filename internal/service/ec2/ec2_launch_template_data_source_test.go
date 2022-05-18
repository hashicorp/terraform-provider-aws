package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2LaunchTemplateDataSource_name(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateDataSourceNameConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "block_device_mappings.#", dataSourceName, "block_device_mappings.#"),
					resource.TestCheckResourceAttrPair(resourceName, "capacity_reservation_specification.#", dataSourceName, "capacity_reservation_specification.#"),
					resource.TestCheckResourceAttrPair(resourceName, "cpu_options.#", dataSourceName, "cpu_options.#"),
					resource.TestCheckResourceAttrPair(resourceName, "credit_specification.#", dataSourceName, "credit_specification.#"),
					resource.TestCheckResourceAttrPair(resourceName, "default_version", dataSourceName, "default_version"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "disable_api_termination", dataSourceName, "disable_api_termination"),
					resource.TestCheckResourceAttrPair(resourceName, "ebs_optimized", dataSourceName, "ebs_optimized"),
					resource.TestCheckResourceAttrPair(resourceName, "elastic_gpu_specifications.#", dataSourceName, "elastic_gpu_specifications.#"),
					resource.TestCheckResourceAttrPair(resourceName, "elastic_inference_accelerator.#", dataSourceName, "elastic_inference_accelerator.#"),
					resource.TestCheckResourceAttrPair(resourceName, "enclave_options.#", dataSourceName, "enclave_options.#"),
					resource.TestCheckResourceAttrPair(resourceName, "hibernation_options.#", dataSourceName, "hibernation_options.#"),
					resource.TestCheckResourceAttrPair(resourceName, "iam_instance_profile.#", dataSourceName, "iam_instance_profile.#"),
					resource.TestCheckResourceAttrPair(resourceName, "image_id", dataSourceName, "image_id"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_initiated_shutdown_behavior", dataSourceName, "instance_initiated_shutdown_behavior"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_market_options.#", dataSourceName, "instance_market_options.#"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_requirements.#", dataSourceName, "instance_requirements.#"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_type", dataSourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(resourceName, "kernel_id", dataSourceName, "kernel_id"),
					resource.TestCheckResourceAttrPair(resourceName, "key_name", dataSourceName, "key_name"),
					resource.TestCheckResourceAttrPair(resourceName, "latest_version", dataSourceName, "latest_version"),
					resource.TestCheckResourceAttrPair(resourceName, "license_specification.#", dataSourceName, "license_specification.#"),
					resource.TestCheckResourceAttrPair(resourceName, "maintenance_options.#", dataSourceName, "maintenance_options.#"),
					resource.TestCheckResourceAttrPair(resourceName, "metadata_options.#", dataSourceName, "metadata_options.#"),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring.#", dataSourceName, "monitoring.#"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "network_interfaces.#", dataSourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(resourceName, "placement.#", dataSourceName, "placement.#"),
					resource.TestCheckResourceAttrPair(resourceName, "private_dns_name_options.#", dataSourceName, "private_dns_name_options.#"),
					resource.TestCheckResourceAttrPair(resourceName, "ram_disk_id", dataSourceName, "ram_disk_id"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_names.#", dataSourceName, "security_group_names.#"),
					resource.TestCheckResourceAttrPair(resourceName, "tag_specifications.#", dataSourceName, "tag_specifications.#"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "user_data", dataSourceName, "user_data"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_security_group_ids.#", dataSourceName, "vpc_security_group_ids.#"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplateDataSource_id(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateDataSourceIDConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplateDataSource_filter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateDataSourceFilterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
				),
			},
		},
	})
}

func TestAccEC2LaunchTemplateDataSource_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchTemplateDataSourceTagsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccLaunchTemplateDataSourceNameConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  name          = %[1]q

  block_device_mappings {
    device_name = "/dev/xvda"

    ebs {
      iops        = 4000
      throughput  = 500
      volume_size = 15
      volume_type = "gp3"
    }
  }

  elastic_inference_accelerator {
    type = "eia1.medium"
  }

  elastic_gpu_specifications {
    type = "test"
  }

  iam_instance_profile {
    name = "test"
  }

  instance_initiated_shutdown_behavior = "terminate"

  instance_market_options {
    market_type = "spot"
  }

  maintenance_options {
    auto_recovery = "disabled"
  }

  disable_api_termination = true
  ebs_optimized           = false

  kernel_id = "aki-a12bc3de"
  key_name  = "test"

  placement {
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  ram_disk_id            = "ari-a12bc3de"
  vpc_security_group_ids = ["sg-12a3b45c"]

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = "test"
    }
  }

  tag_specifications {
    resource_type = "volume"

    tags = {
      Name = "test"
    }
  }

  tags = {
    Name = %[1]q
  }

  capacity_reservation_specification {
    capacity_reservation_preference = "open"
  }

  cpu_options {
    core_count       = 4
    threads_per_core = 2
  }

  credit_specification {
    cpu_credits = "unlimited"
  }
}

data "aws_launch_template" "test" {
  name = aws_launch_template.test.name
}
`, rName))
}

func testAccLaunchTemplateDataSourceIDConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q
}

data "aws_launch_template" "test" {
  id = aws_launch_template.test.id
}
`, rName)
}

func testAccLaunchTemplateDataSourceFilterConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q
}

data "aws_launch_template" "test" {
  filter {
    name   = "launch-template-name"
    values = [aws_launch_template.test.name]
  }
}
`, rName)
}

func testAccLaunchTemplateDataSourceTagsConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_launch_template" "test" {
  tags = {
    Name = aws_launch_template.test.tags["Name"]
  }
}
`, rName)
}
