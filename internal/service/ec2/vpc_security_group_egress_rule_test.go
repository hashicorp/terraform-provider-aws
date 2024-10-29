// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCSecurityGroupEgressRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_egress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupEgressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupEgressRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupEgressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "cidr_ipv4", "10.0.0.0/8"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrTags),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8080"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupEgressRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_egress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupEgressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupEgressRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupEgressRuleExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceSecurityGroupEgressRule, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSecurityGroupEgressRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_security_group_egress_rule" {
				continue
			}

			_, err := tfec2.FindSecurityGroupEgressRuleByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Security Group Egress Rule still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSecurityGroupEgressRuleExists(ctx context.Context, n string, v *awstypes.SecurityGroupRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Security Group Egress Rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindSecurityGroupEgressRuleByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCSecurityGroupEgressRuleConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), `
resource "aws_vpc_security_group_egress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080
}
`)
}
