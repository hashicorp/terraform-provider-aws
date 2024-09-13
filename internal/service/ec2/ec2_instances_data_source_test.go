// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2InstancesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_ids(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instances.test", "ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr("data.aws_instances.test", "ipv6_addresses.#", acctest.Ct2),
					resource.TestCheckResourceAttr("data.aws_instances.test", "private_ips.#", acctest.Ct2),
					// Public IP values are flakey for new EC2 instances due to eventual consistency
					resource.TestCheckResourceAttrSet("data.aws_instances.test", "public_ips.#"),
				),
			},
		},
	})
}

func TestAccEC2InstancesDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instances.test", "ids.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccEC2InstancesDataSource_instanceStateNames(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_instanceStateNames(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instances.test", "ids.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccEC2InstancesDataSource_empty(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instances.test", "ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr("data.aws_instances.test", "ipv6_addresses.#", acctest.Ct0),
					resource.TestCheckResourceAttr("data.aws_instances.test", "private_ips.#", acctest.Ct0),
					resource.TestCheckResourceAttr("data.aws_instances.test", "public_ips.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccEC2InstancesDataSource_timeout(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_timeout(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instances.test", "ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet("data.aws_instances.test", "ipv6_addresses.#"),
					resource.TestCheckResourceAttr("data.aws_instances.test", "private_ips.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet("data.aws_instances.test", "public_ips.#"),
				),
			},
		},
	})
}

func testAccInstancesDataSourceConfig_ids(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		acctest.ConfigVPCWithSubnetsIPv6(rName, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  count              = 2
  ami                = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type      = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id          = aws_subnet.test[0].id
  ipv6_address_count = 1

  tags = {
    Name = %[1]q
  }
}

data "aws_instances" "test" {
  filter {
    name   = "instance-id"
    values = aws_instance.test[*].id
  }
}
`, rName))
}

func testAccInstancesDataSourceConfig_tags(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  count         = 2
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name      = %[1]q
    SecondTag = "%[1]s-2"
  }
}

data "aws_instances" "test" {
  instance_tags = {
    Name      = aws_instance.test[0].tags["Name"]
    SecondTag = aws_instance.test[0].tags["SecondTag"]
  }

  depends_on = [aws_instance.test]
}
`, rName))
}

func testAccInstancesDataSourceConfig_instanceStateNames(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  count         = 2
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = %[1]q
  }
}

data "aws_instances" "test" {
  instance_tags = {
    Name = aws_instance.test[0].tags["Name"]
  }

  instance_state_names = ["pending", "running"]
  depends_on           = [aws_instance.test]
}
`, rName))
}

func testAccInstancesDataSourceConfig_empty(rName string) string {
	return fmt.Sprintf(`
data "aws_instances" "test" {
  instance_tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccInstancesDataSourceConfig_timeout(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  count         = 2
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = %[1]q
  }
}

data "aws_instances" "test" {
  filter {
    name   = "instance-id"
    values = aws_instance.test[*].id
  }

  timeouts {
    read = "60m"
  }
}
`, rName))
}
