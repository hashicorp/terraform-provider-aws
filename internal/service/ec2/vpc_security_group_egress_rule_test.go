// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCSecurityGroupEgressRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SecurityGroupRule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_egress_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupEgressRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupEgressRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupEgressRuleExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "ec2", "security-group-rule/{id}"),
					resource.TestCheckResourceAttr(resourceName, "cidr_ipv4", "10.0.0.0/8"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, "security_group_rule_id"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_egress_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupEgressRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupEgressRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupEgressRuleExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceSecurityGroupEgressRule, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccVPCSecurityGroupEgressRule_moveWithSingleSource(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SecurityGroupRule
	var group awstypes.SecurityGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	fromResourceName := "aws_security_group_rule.test"
	toResourceName := "aws_vpc_security_group_egress_rule.test"
	sgResourceName := "aws_security_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupEgressRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupEgressRuleConfig_moveFrom(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, t, sgResourceName, &group),
					resource.TestCheckResourceAttr(fromResourceName, names.AttrType, "egress"),
					resource.TestCheckResourceAttr(fromResourceName, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(fromResourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(fromResourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(fromResourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(fromResourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(fromResourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckNoResourceAttr(fromResourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(fromResourceName, "to_port", "8080"),
				),
			},
			{
				Config: testAccVPCSecurityGroupEgressRuleConfig_moveTo(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupEgressRuleExists(ctx, t, toResourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, toResourceName, names.AttrARN, "ec2", "security-group-rule/{id}"),
					resource.TestCheckResourceAttr(toResourceName, "cidr_ipv4", "10.0.0.0/8"),
					resource.TestCheckNoResourceAttr(toResourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(toResourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(toResourceName, "from_port", "80"),
					resource.TestCheckResourceAttrPair(toResourceName, names.AttrID, toResourceName, "security_group_rule_id"),
					resource.TestCheckResourceAttr(toResourceName, "ip_protocol", "tcp"),
					resource.TestCheckNoResourceAttr(toResourceName, "prefix_list_id"),
					resource.TestCheckNoResourceAttr(toResourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrSet(toResourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(toResourceName, names.AttrTags),
					resource.TestCheckResourceAttr(toResourceName, "to_port", "8080"),
				),
			},
		},
	})
}

func testAccVPCSecurityGroupEgressRuleConfig_moveFrom(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), `
resource "aws_security_group_rule" "test" {
  security_group_id = aws_security_group.test.id

  type        = "egress"
  cidr_blocks = ["10.0.0.0/8"]
  from_port   = 80
  protocol    = "tcp"
  to_port     = 8080
}
`)
}

func testAccVPCSecurityGroupEgressRuleConfig_moveTo(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupEgressRuleConfig_basic(rName), `
moved {
  from = aws_security_group_rule.test
  to   = aws_vpc_security_group_egress_rule.test
}
`)
}

func testAccCheckSecurityGroupEgressRuleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_security_group_egress_rule" {
				continue
			}

			_, err := tfec2.FindSecurityGroupEgressRuleByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckSecurityGroupEgressRuleExists(ctx context.Context, t *testing.T, n string, v *awstypes.SecurityGroupRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Security Group Egress Rule ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

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
