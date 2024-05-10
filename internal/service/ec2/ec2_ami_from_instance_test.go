// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2AMIFromInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var image ec2.Image
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ami_from_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAMIFromInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(ctx, resourceName, &image),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`image/ami-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Testing Terraform aws_ami_from_instance resource"),
					resource.TestCheckResourceAttr(resourceName, "usage_operation", "RunInstances"),
					resource.TestCheckResourceAttr(resourceName, "platform_details", "Linux/UNIX"),
					resource.TestCheckResourceAttr(resourceName, "image_type", "machine"),
					resource.TestCheckResourceAttr(resourceName, "hypervisor", "xen"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccEC2AMIFromInstance_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var image ec2.Image
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ami_from_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAMIFromInstanceConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAMIFromInstanceConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAMIFromInstanceConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2AMIFromInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var image ec2.Image
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ami_from_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAMIFromInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMIExists(ctx, resourceName, &image),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceAMIFromInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAMIFromInstanceBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAMIFromInstanceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAMIFromInstanceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ami_from_instance" "test" {
  name               = %[1]q
  description        = "Testing Terraform aws_ami_from_instance resource"
  source_instance_id = aws_instance.test.id
}
`, rName))
}

func testAccAMIFromInstanceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccAMIFromInstanceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ami_from_instance" "test" {
  name               = %[1]q
  description        = "Testing Terraform aws_ami_from_instance resource"
  source_instance_id = aws_instance.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAMIFromInstanceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccAMIFromInstanceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ami_from_instance" "test" {
  name               = %[1]q
  description        = "Testing Terraform aws_ami_from_instance resource"
  source_instance_id = aws_instance.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
