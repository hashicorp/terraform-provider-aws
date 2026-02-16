// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCSecurityGroupRulesExclusive_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_rules_exclusive.test"
	securityGroupResourceName := "aws_security_group.test"
	ingressRuleResourceName := "aws_vpc_security_group_ingress_rule.test"
	egressRuleResourceName := "aws_vpc_security_group_egress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRulesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRulesExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "ingress_rule_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "egress_rule_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ingress_rule_ids.*", ingressRuleResourceName, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "egress_rule_ids.*", egressRuleResourceName, names.AttrID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "security_group_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "security_group_id",
			},
		},
	})
}

func TestAccVPCSecurityGroupRulesExclusive_disappears_SecurityGroup(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_rules_exclusive.test"
	securityGroupResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRulesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRulesExclusiveExists(ctx, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfec2.ResourceSecurityGroup(), securityGroupResourceName),
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

func TestAccVPCSecurityGroupRulesExclusive_multiple(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_rules_exclusive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRulesExclusiveConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRulesExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ingress_rule_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "egress_rule_ids.#", "2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "security_group_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "security_group_id",
			},
			{
				Config: testAccSecurityGroupRulesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRulesExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ingress_rule_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "egress_rule_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRulesExclusive_empty(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_rules_exclusive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRulesExclusiveConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRulesExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ingress_rule_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "egress_rule_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRulesExclusive_outOfBandAddition(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_rules_exclusive.test"
	securityGroupResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRulesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRulesExclusiveExists(ctx, resourceName),
					testAccCheckSecurityGroupRulesExclusiveAddOutOfBandIngressRule(ctx, securityGroupResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccSecurityGroupRulesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRulesExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ingress_rule_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "egress_rule_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRulesExclusive_ingressOnly(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_rules_exclusive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRulesExclusiveConfig_ingressOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRulesExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ingress_rule_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "egress_rule_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRulesExclusive_egressOnly(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_rules_exclusive.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRulesExclusiveConfig_egressOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRulesExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ingress_rule_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "egress_rule_ids.#", "1"),
				),
			},
		},
	})
}

func testAccCheckSecurityGroupRulesExclusiveExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		ingressRuleIDs, egressRuleIDs, err := tfec2.FindSecurityGroupRuleIDsBySecurityGroupID(ctx, conn, rs.Primary.Attributes["security_group_id"])
		if err != nil {
			return err
		}

		if strconv.Itoa(len(ingressRuleIDs)) != rs.Primary.Attributes["ingress_rule_ids.#"] {
			return fmt.Errorf("ingress rule count mismatch")
		}
		if strconv.Itoa(len(egressRuleIDs)) != rs.Primary.Attributes["egress_rule_ids.#"] {
			return fmt.Errorf("egress rule count mismatch")
		}

		return nil
	}
}

func testAccCheckSecurityGroupRulesExclusiveAddOutOfBandIngressRule(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		// Add an out-of-band ingress rule
		input := ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:    aws.String(rs.Primary.ID),
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int32(8080),
			ToPort:     aws.Int32(8080),
			CidrIp:     aws.String("10.0.0.0/8"),
		}
		_, err := conn.AuthorizeSecurityGroupIngress(ctx, &input)

		return err
	}
}

func testAccSecurityGroupRulesExclusiveConfigBase(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 0),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccSecurityGroupRulesExclusiveConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccSecurityGroupRulesExclusiveConfigBase(rName),
		`
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  to_port     = 80
  ip_protocol = "tcp"
}

resource "aws_vpc_security_group_egress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "0.0.0.0/0"
  ip_protocol = "-1"
}

resource "aws_vpc_security_group_rules_exclusive" "test" {
  security_group_id = aws_security_group.test.id
  ingress_rule_ids  = [aws_vpc_security_group_ingress_rule.test.id]
  egress_rule_ids   = [aws_vpc_security_group_egress_rule.test.id]
}
`)
}

func testAccSecurityGroupRulesExclusiveConfig_multiple(rName string) string {
	return acctest.ConfigCompose(
		testAccSecurityGroupRulesExclusiveConfigBase(rName),
		`
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  to_port     = 80
  ip_protocol = "tcp"
}

resource "aws_vpc_security_group_ingress_rule" "test2" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 443
  to_port     = 443
  ip_protocol = "tcp"
}

resource "aws_vpc_security_group_ingress_rule" "test3" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 8080
  to_port     = 8080
  ip_protocol = "tcp"
}

resource "aws_vpc_security_group_egress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "0.0.0.0/0"
  ip_protocol = "-1"
}

resource "aws_vpc_security_group_egress_rule" "test2" {
  security_group_id = aws_security_group.test.id

  cidr_ipv6   = "::/0"
  ip_protocol = "-1"
}

resource "aws_vpc_security_group_rules_exclusive" "test" {
  security_group_id = aws_security_group.test.id
  ingress_rule_ids = [
    aws_vpc_security_group_ingress_rule.test.id,
    aws_vpc_security_group_ingress_rule.test2.id,
    aws_vpc_security_group_ingress_rule.test3.id,
  ]
  egress_rule_ids = [
    aws_vpc_security_group_egress_rule.test.id,
    aws_vpc_security_group_egress_rule.test2.id,
  ]
}
`)
}

func testAccSecurityGroupRulesExclusiveConfig_empty(rName string) string {
	return acctest.ConfigCompose(
		testAccSecurityGroupRulesExclusiveConfigBase(rName),
		`
resource "aws_vpc_security_group_rules_exclusive" "test" {
  security_group_id = aws_security_group.test.id
  ingress_rule_ids  = []
  egress_rule_ids   = []
}
`)
}

func testAccSecurityGroupRulesExclusiveConfig_ingressOnly(rName string) string {
	return acctest.ConfigCompose(
		testAccSecurityGroupRulesExclusiveConfigBase(rName),
		`
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  to_port     = 80
  ip_protocol = "tcp"
}

resource "aws_vpc_security_group_rules_exclusive" "test" {
  security_group_id = aws_security_group.test.id
  ingress_rule_ids  = [aws_vpc_security_group_ingress_rule.test.id]
  egress_rule_ids   = []
}
`)
}

func testAccSecurityGroupRulesExclusiveConfig_egressOnly(rName string) string {
	return acctest.ConfigCompose(
		testAccSecurityGroupRulesExclusiveConfigBase(rName),
		`
resource "aws_vpc_security_group_egress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "0.0.0.0/0"
  ip_protocol = "-1"
}

resource "aws_vpc_security_group_rules_exclusive" "test" {
  security_group_id = aws_security_group.test.id
  ingress_rule_ids  = []
  egress_rule_ids   = [aws_vpc_security_group_egress_rule.test.id]
}
`)
}
