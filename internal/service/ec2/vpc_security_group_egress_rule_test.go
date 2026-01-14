// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
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
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceSecurityGroupEgressRule, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupEgressRule_ReferencedSecurityGroupID_peerVPCUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_egress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckSecurityGroupEgressRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupEgressRuleConfig_referencedSecurityGroupIDPeerVPCWithDescription(rName, "original description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupEgressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "original description"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSecurityGroupEgressRuleConfig_referencedSecurityGroupIDPeerVPCWithDescription(rName, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupEgressRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated description"),
				),
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

func testAccVPCSecurityGroupEgressRuleConfig_referencedSecurityGroupIDPeerVPCWithDescription(rName, description string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), testAccVPCSecurityGroupRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpc" "peer" {
	provider = "awsalternate"

	cidr_block = "10.1.0.0/16"

	tags = {
		Name = %[1]q
	}
}

resource "aws_security_group" "peer" {
	provider = "awsalternate"

	vpc_id = aws_vpc.peer.id
	name   = %[1]q

	tags = {
		Name = %[1]q
	}
}

data "aws_caller_identity" "peer" {
	provider = "awsalternate"
}

resource "aws_vpc_peering_connection" "test" {
	vpc_id        = aws_vpc.test.id
	peer_vpc_id   = aws_vpc.peer.id
	peer_owner_id = data.aws_caller_identity.peer.account_id
	peer_region   = %[2]q
	auto_accept   = false

	tags = {
		Name = %[1]q
	}
}

resource "aws_vpc_peering_connection_accepter" "peer" {
	provider = "awsalternate"

	vpc_peering_connection_id = aws_vpc_peering_connection.test.id
	auto_accept               = true

	tags = {
		Name = %[1]q
	}
}

resource "aws_vpc_security_group_egress_rule" "test" {
	security_group_id = aws_security_group.test.id

	referenced_security_group_id = "${data.aws_caller_identity.peer.account_id}/${aws_security_group.peer.id}"
	description                  = %[3]q
	from_port                    = 80
	ip_protocol                  = "tcp"
	to_port                      = 8080

	depends_on = [aws_vpc_peering_connection_accepter.peer]
}
`, rName, acctest.Region(), description))
}
