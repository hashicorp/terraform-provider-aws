// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCDefaultSecurityGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_default_security_group.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultSecurityGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "name", "default"),
					resource.TestCheckResourceAttr(resourceName, "description", "default VPC security group"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"protocol":      "tcp",
						"from_port":     "80",
						"to_port":       "8000",
						"cidr_blocks.#": "1",
						"cidr_blocks.0": "10.0.0.0/8",
					}),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"protocol":      "tcp",
						"from_port":     "80",
						"to_port":       "8000",
						"cidr_blocks.#": "1",
						"cidr_blocks.0": "10.0.0.0/8",
					}),
					testAccCheckDefaultSecurityGroupARN(resourceName, &group),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config:   testAccVPCDefaultSecurityGroupConfig_basic(rName),
				PlanOnly: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCDefaultSecurityGroup_empty(t *testing.T) {
	ctx := acctest.Context(t)
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_default_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultSecurityGroupConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func testAccCheckDefaultSecurityGroupARN(resourceName string, group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ec2", fmt.Sprintf("security-group/%s", aws.StringValue(group.GroupId)))(s)
	}
}

func testAccVPCDefaultSecurityGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }
}
`, rName)
}

func testAccVPCDefaultSecurityGroupConfig_empty(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}
