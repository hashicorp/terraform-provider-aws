// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccVPCNetworkACLRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resource1Name := "aws_network_acl_rule.test1"
	resource2Name := "aws_network_acl_rule.test2"
	resource3Name := "aws_network_acl_rule.test3"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLRuleExists(ctx, resource1Name),
					testAccCheckNetworkACLRuleExists(ctx, resource2Name),
					testAccCheckNetworkACLRuleExists(ctx, resource3Name),

					resource.TestCheckResourceAttr(resource1Name, names.AttrCIDRBlock, "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resource1Name, "egress", acctest.CtFalse),
					resource.TestCheckResourceAttr(resource1Name, "from_port", "22"),
					resource.TestCheckResourceAttr(resource1Name, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resource1Name, names.AttrProtocol, "6"),
					resource.TestCheckResourceAttr(resource1Name, "rule_action", "allow"),
					resource.TestCheckResourceAttr(resource1Name, "rule_number", "200"),
					resource.TestCheckResourceAttr(resource1Name, "to_port", "22"),

					resource.TestCheckResourceAttr(resource2Name, names.AttrCIDRBlock, "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resource2Name, "egress", acctest.CtFalse),
					resource.TestCheckResourceAttr(resource2Name, "icmp_code", "-1"),
					resource.TestCheckResourceAttr(resource2Name, "icmp_type", acctest.Ct0),
					resource.TestCheckResourceAttr(resource2Name, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resource2Name, names.AttrProtocol, acctest.Ct1),
					resource.TestCheckResourceAttr(resource2Name, "rule_action", "allow"),
					resource.TestCheckResourceAttr(resource2Name, "rule_number", "300"),

					resource.TestCheckResourceAttr(resource3Name, names.AttrCIDRBlock, "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resource3Name, "egress", acctest.CtFalse),
					resource.TestCheckResourceAttr(resource3Name, "icmp_code", "-1"),
					resource.TestCheckResourceAttr(resource3Name, "icmp_type", "-1"),
					resource.TestCheckResourceAttr(resource3Name, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resource3Name, names.AttrProtocol, acctest.Ct1),
					resource.TestCheckResourceAttr(resource3Name, "rule_action", "allow"),
					resource.TestCheckResourceAttr(resource3Name, "rule_number", "400"),
				),
			},
			{
				ResourceName:      resource1Name,
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc(resource1Name, "tcp"),
				ImportStateVerify: true,
			},
			{
				ResourceName:      resource2Name,
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc(resource2Name, "icmp"),
				ImportStateVerify: true,
			},
			{
				ResourceName:      resource3Name,
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc(resource3Name, "icmp"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCNetworkACLRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_network_acl_rule.test1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLRuleExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceNetworkACLRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNetworkACLRule_Disappears_networkACL(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_network_acl_rule.test1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLRuleExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceNetworkACL(), "aws_network_acl.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNetworkACLRule_Disappears_ingressEgressSameNumber(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_network_acl_rule.test1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLRuleConfig_ingressEgressSameNumberMissing(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLRuleExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceNetworkACLRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNetworkACLRule_ipv6(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_network_acl_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLRuleConfig_ipv6(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, ""),
					resource.TestCheckResourceAttr(resourceName, "egress", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "from_port", "22"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", "::/0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "6"),
					resource.TestCheckResourceAttr(resourceName, "rule_action", "allow"),
					resource.TestCheckResourceAttr(resourceName, "rule_number", "150"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "22"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc(resourceName, "tcp"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCNetworkACLRule_ipv6ICMP(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_network_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLRuleConfig_ipv6ICMP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, ""),
					resource.TestCheckResourceAttr(resourceName, "egress", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "icmp_code", "-1"),
					resource.TestCheckResourceAttr(resourceName, "icmp_type", "-1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", "::/0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "58"),
					resource.TestCheckResourceAttr(resourceName, "rule_action", "allow"),
					resource.TestCheckResourceAttr(resourceName, "rule_number", "150"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc(resourceName, "58"),
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/6710
func TestAccVPCNetworkACLRule_ipv6VPCAssignGeneratedIPv6CIDRBlockUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Vpc
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	vpcResourceName := "aws_vpc.test"
	resourceName := "aws_network_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLRuleConfig_ipv6NotAssignGeneratedIPv6CIDRBlockUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, vpcResourceName, &v),
					resource.TestCheckResourceAttr(vpcResourceName, "assign_generated_ipv6_cidr_block", acctest.CtFalse),
					resource.TestCheckResourceAttr(vpcResourceName, "ipv6_cidr_block", ""),
				),
			},
			{
				Config: testAccVPCNetworkACLRuleConfig_ipv6AssignGeneratedIPv6CIDRBlockUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, vpcResourceName, &v),
					testAccCheckNetworkACLRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(vpcResourceName, "assign_generated_ipv6_cidr_block", acctest.CtTrue),
					resource.TestMatchResourceAttr(vpcResourceName, "ipv6_cidr_block", regexache.MustCompile(`/56$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc(resourceName, "tcp"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCNetworkACLRule_allProtocol(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_network_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLRuleConfig_allProtocol(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "egress", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "from_port", "22"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "-1"),
					resource.TestCheckResourceAttr(resourceName, "rule_action", "allow"),
					resource.TestCheckResourceAttr(resourceName, "rule_number", "150"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "22"),
				),
			},
			{
				Config:   testAccVPCNetworkACLRuleConfig_allProtocolNoRealUpdate(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCNetworkACLRule_tcpProtocol(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_network_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkACLRuleConfig_tcpProtocol(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkACLRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "egress", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "from_port", "22"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "6"),
					resource.TestCheckResourceAttr(resourceName, "rule_action", "deny"),
					resource.TestCheckResourceAttr(resourceName, "rule_number", "150"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "22"),
				),
			},
			{
				Config:   testAccVPCNetworkACLRuleConfig_tcpProtocolNoRealUpdate(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCNetworkACLRule_duplicate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCNetworkACLRuleConfig_duplicate(rName),
				ExpectError: regexache.MustCompile(`NetworkAclEntryAlreadyExists: EC2 Network ACL .* already exists`),
			},
		},
	})
}

func testAccCheckNetworkACLRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_network_acl_rule" {
				continue
			}

			egress, err := strconv.ParseBool(rs.Primary.Attributes["egress"])

			if err != nil {
				return err
			}

			naclID := rs.Primary.Attributes["network_acl_id"]

			ruleNumber, err := strconv.Atoi(rs.Primary.Attributes["rule_number"])

			if err != nil {
				return err
			}

			_, err = tfec2.FindNetworkACLEntryByThreePartKey(ctx, conn, naclID, egress, ruleNumber)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Network ACL Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNetworkACLRuleExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Network ACL Rule ID is set: %s", n)
		}

		egress, err := strconv.ParseBool(rs.Primary.Attributes["egress"])

		if err != nil {
			return err
		}

		naclID := rs.Primary.Attributes["network_acl_id"]

		ruleNumber, err := strconv.Atoi(rs.Primary.Attributes["rule_number"])

		if err != nil {
			return err
		}

		_, err = tfec2.FindNetworkACLEntryByThreePartKey(ctx, conn, naclID, egress, ruleNumber)

		return err
	}
}

func testAccVPCNetworkACLRuleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl_rule" "test1" {
  network_acl_id = aws_network_acl.test.id
  rule_number    = 200
  egress         = false
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}

resource "aws_network_acl_rule" "test2" {
  network_acl_id = aws_network_acl.test.id
  rule_number    = 300
  protocol       = "icmp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  icmp_type      = 0
  icmp_code      = -1
}

resource "aws_network_acl_rule" "test3" {
  network_acl_id = aws_network_acl.test.id
  rule_number    = 400
  protocol       = "icmp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  icmp_type      = -1
  icmp_code      = -1
}
`, rName)
}

func testAccVPCNetworkACLRuleConfig_allProtocolNoRealUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl_rule" "test" {
  network_acl_id = aws_network_acl.test.id
  rule_number    = 150
  egress         = false
  protocol       = "all"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`, rName)
}

func testAccVPCNetworkACLRuleConfig_tcpProtocolNoRealUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl_rule" "test" {
  network_acl_id = aws_network_acl.test.id
  rule_number    = 150
  egress         = true
  protocol       = "tcp"
  rule_action    = "deny"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`, rName)
}

func testAccVPCNetworkACLRuleConfig_allProtocol(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl_rule" "test" {
  network_acl_id = aws_network_acl.test.id
  rule_number    = 150
  egress         = false
  protocol       = "-1"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`, rName)
}

func testAccVPCNetworkACLRuleConfig_tcpProtocol(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl_rule" "test" {
  network_acl_id = aws_network_acl.test.id
  rule_number    = 150
  egress         = true
  protocol       = "6"
  rule_action    = "deny"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`, rName)
}

func testAccVPCNetworkACLRuleConfig_ipv6(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl_rule" "test" {
  network_acl_id  = aws_network_acl.test.id
  rule_number     = 150
  egress          = false
  protocol        = "tcp"
  rule_action     = "allow"
  ipv6_cidr_block = "::/0"
  from_port       = 22
  to_port         = 22
}
`, rName)
}

func testAccVPCNetworkACLRuleConfig_ingressEgressSameNumberMissing(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl_rule" "test1" {
  network_acl_id = aws_network_acl.test.id
  rule_number    = 100
  egress         = false
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}

resource "aws_network_acl_rule" "test2" {
  network_acl_id = aws_network_acl.test.id
  rule_number    = 100
  egress         = true
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`, rName)
}

func testAccVPCNetworkACLRuleConfig_ipv6ICMP(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl_rule" "test" {
  icmp_code       = -1
  icmp_type       = -1
  ipv6_cidr_block = "::/0"
  network_acl_id  = aws_network_acl.test.id
  protocol        = 58
  rule_action     = "allow"
  rule_number     = 150
}
`, rName, rName)
}

func testAccVPCNetworkACLRuleConfig_ipv6AssignGeneratedIPv6CIDRBlockUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl_rule" "test" {
  from_port       = 22
  ipv6_cidr_block = aws_vpc.test.ipv6_cidr_block
  network_acl_id  = aws_network_acl.test.id
  protocol        = "tcp"
  rule_action     = "allow"
  rule_number     = 150
  to_port         = 22
}
`, rName)
}

func testAccVPCNetworkACLRuleConfig_ipv6NotAssignGeneratedIPv6CIDRBlockUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = false
  cidr_block                       = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}`, rName)
}

func testAccVPCNetworkACLRuleConfig_duplicate(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl_rule" "test1" {
  network_acl_id = aws_network_acl.test.id
  rule_number    = 200
  egress         = false
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}

resource "aws_network_acl_rule" "test2" {
  network_acl_id = aws_network_acl_rule.test1.network_acl_id
  rule_number    = 200
  egress         = false
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`, rName)
}

func testAccNetworkACLRuleImportStateIdFunc(resourceName, resourceProtocol string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		naclID := rs.Primary.Attributes["network_acl_id"]
		ruleNumber := rs.Primary.Attributes["rule_number"]
		protocol := rs.Primary.Attributes[names.AttrProtocol]
		// Ensure the resource's ID will be determined from the original protocol value set in the resource's config
		if protocol != resourceProtocol {
			protocol = resourceProtocol
		}
		egress := rs.Primary.Attributes["egress"]

		return strings.Join([]string{naclID, ruleNumber, protocol, egress}, tfec2.NetworkACLRuleImportIDSeparator), nil
	}
}
