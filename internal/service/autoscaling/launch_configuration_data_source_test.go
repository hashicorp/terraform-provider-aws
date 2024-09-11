// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAutoScalingLaunchConfigurationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_launch_configuration.test"
	datasourceName := "data.aws_launch_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "associate_public_ip_address", resourceName, "associate_public_ip_address"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_block_device.#", resourceName, "ebs_block_device.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_optimized", resourceName, "ebs_optimized"),
					resource.TestCheckResourceAttrPair(datasourceName, "enable_monitoring", resourceName, "enable_monitoring"),
					resource.TestCheckResourceAttrPair(datasourceName, "ephemeral_block_device.#", resourceName, "ephemeral_block_device.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "iam_instance_profile", resourceName, "iam_instance_profile"),
					resource.TestCheckResourceAttrPair(datasourceName, "image_id", resourceName, "image_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceType, resourceName, names.AttrInstanceType),
					resource.TestCheckResourceAttrPair(datasourceName, "metadata_options.#", resourceName, "metadata_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "placement_tenancy", resourceName, "placement_tenancy"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.#", resourceName, "root_block_device.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "spot_price", resourceName, "spot_price"),
					// Resource and data source user_data have differing representations in state.
					resource.TestCheckResourceAttrSet(datasourceName, "user_data"),
				),
			},
		},
	})
}

func TestAccAutoScalingLaunchConfigurationDataSource_securityGroups(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_launch_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationDataSourceConfig_securityGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "security_groups.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccAutoScalingLaunchConfigurationDataSource_ebsNoDevice(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_launch_configuration.test"
	datasourceName := "data.aws_launch_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationDataSourceConfig_ebsNoDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_block_device.#", resourceName, "ebs_block_device.#"),
				),
			},
		},
	})
}

func TestAccAutoScalingLaunchConfigurationDataSource_metadataOptions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_launch_configuration.test"
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationDataSourceConfig_metaOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.#", resourceName, "metadata_options.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.0.http_endpoint", resourceName, "metadata_options.0.http_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.0.http_put_response_hop_limit", resourceName, "metadata_options.0.http_put_response_hop_limit"),
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.0.http_tokens", resourceName, "metadata_options.0.http_tokens"),
				),
			},
		},
	})
}

func testAccLaunchConfigurationDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name                        = %[1]q
  image_id                    = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = "m1.small"
  associate_public_ip_address = true
  user_data                   = "test-user-data"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 9
  }

  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 10
    volume_type = "gp3"
    iops        = 3000
    throughput  = 125
  }

  ephemeral_block_device {
    device_name  = "/dev/sde"
    virtual_name = "ephemeral0"
  }
}

data "aws_launch_configuration" "test" {
  name = aws_launch_configuration.test.name
}
`, rName))
}

func testAccLaunchConfigurationDataSourceConfig_securityGroups(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_launch_configuration" "test" {
  name            = %[1]q
  image_id        = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type   = "m1.small"
  security_groups = [aws_security_group.test.id]
}

data "aws_launch_configuration" "test" {
  name = aws_launch_configuration.test.name
}
`, rName))
}

func testAccLaunchConfigurationDataSourceConfig_metaOptions(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.nano"
  name          = %[1]q

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
  }
}

data "aws_launch_configuration" "test" {
  name = aws_launch_configuration.test.name
}
`, rName))
}

func testAccLaunchConfigurationDataSourceConfig_ebsNoDevice(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "m1.small"

  ebs_block_device {
    device_name = "/dev/sda2"
    no_device   = true
  }
}

data "aws_launch_configuration" "test" {
  name = aws_launch_configuration.test.name
}
`, rName))
}
