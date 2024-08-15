// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccClientVPNAuthorizationRule_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.AuthorizationRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_authorization_rule.test"
	subnetResourceName := "aws_subnet.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNAuthorizationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNAuthorizationRuleConfig_basic(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNAuthorizationRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "target_network_cidr", subnetResourceName, names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(resourceName, "authorize_all_groups", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "access_group_id", ""),
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

func testAccClientVPNAuthorizationRule_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.AuthorizationRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_authorization_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNAuthorizationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNAuthorizationRuleConfig_basic(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNAuthorizationRuleExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceClientVPNAuthorizationRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccClientVPNAuthorizationRule_Disappears_endpoint(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.AuthorizationRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_authorization_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNAuthorizationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNAuthorizationRuleConfig_basic(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNAuthorizationRuleExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceClientVPNEndpoint(), "aws_ec2_client_vpn_endpoint.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccClientVPNAuthorizationRule_groups(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.AuthorizationRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_ec2_client_vpn_authorization_rule.test1"
	resource2Name := "aws_ec2_client_vpn_authorization_rule.test2"
	subnetResourceName := "aws_subnet.test.0"
	group1Name := "group_one"
	group2Name := "group_two"
	groups1 := map[string]string{
		"test1": group1Name,
	}
	groups2 := map[string]string{
		"test1": group1Name,
		"test2": group2Name,
	}
	groups3 := map[string]string{
		"test2": group2Name,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNAuthorizationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNAuthorizationRuleConfig_groups(t, rName, groups1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNAuthorizationRuleExists(ctx, resource1Name, &v),
					resource.TestCheckResourceAttrPair(resource1Name, "target_network_cidr", subnetResourceName, names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(resource1Name, "authorize_all_groups", acctest.CtFalse),
					resource.TestCheckResourceAttr(resource1Name, "access_group_id", group1Name),
				),
			},
			{
				ResourceName:      resource1Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientVPNAuthorizationRuleConfig_groups(t, rName, groups2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNAuthorizationRuleExists(ctx, resource1Name, &v),
					resource.TestCheckResourceAttrPair(resource1Name, "target_network_cidr", subnetResourceName, names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(resource1Name, "authorize_all_groups", acctest.CtFalse),
					resource.TestCheckResourceAttr(resource1Name, "access_group_id", group1Name),

					testAccCheckClientVPNAuthorizationRuleExists(ctx, resource2Name, &v),
					resource.TestCheckResourceAttrPair(resource2Name, "target_network_cidr", subnetResourceName, names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(resource2Name, "authorize_all_groups", acctest.CtFalse),
					resource.TestCheckResourceAttr(resource2Name, "access_group_id", group2Name),
				),
			},
			{
				ResourceName:      resource2Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientVPNAuthorizationRuleConfig_groups(t, rName, groups3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNAuthorizationRuleExists(ctx, resource2Name, &v),
					resource.TestCheckResourceAttrPair(resource2Name, "target_network_cidr", subnetResourceName, names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(resource2Name, "authorize_all_groups", acctest.CtFalse),
					resource.TestCheckResourceAttr(resource2Name, "access_group_id", group2Name),
				),
			},
		},
	})
}

func testAccClientVPNAuthorizationRule_subnets(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.AuthorizationRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_ec2_client_vpn_authorization_rule.test1"
	resource2Name := "aws_ec2_client_vpn_authorization_rule.test2"
	subnetCount := 2
	subnetIndex1 := 0
	subnetIndex2 := 1
	case1 := map[string]int{
		"test1": subnetIndex1,
		"test2": subnetIndex2,
	}
	case2 := map[string]int{
		"test2": subnetIndex2,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNAuthorizationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNAuthorizationRuleConfig_subnets(t, rName, subnetCount, case1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNAuthorizationRuleExists(ctx, resource1Name, &v),
					resource.TestCheckResourceAttrPair(resource1Name, "target_network_cidr", fmt.Sprintf("aws_subnet.test.%d", subnetIndex1), names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(resource1Name, "authorize_all_groups", acctest.CtTrue),
					resource.TestCheckResourceAttr(resource1Name, "access_group_id", ""),

					testAccCheckClientVPNAuthorizationRuleExists(ctx, resource2Name, &v),
					resource.TestCheckResourceAttrPair(resource2Name, "target_network_cidr", fmt.Sprintf("aws_subnet.test.%d", subnetIndex2), names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(resource2Name, "authorize_all_groups", acctest.CtTrue),
					resource.TestCheckResourceAttr(resource2Name, "access_group_id", ""),
				),
			},
			{
				ResourceName:      resource2Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientVPNAuthorizationRuleConfig_subnets(t, rName, subnetCount, case2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNAuthorizationRuleExists(ctx, resource2Name, &v),
					resource.TestCheckResourceAttrPair(resource2Name, "target_network_cidr", fmt.Sprintf("aws_subnet.test.%d", subnetIndex2), names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(resource2Name, "authorize_all_groups", acctest.CtTrue),
					resource.TestCheckResourceAttr(resource2Name, "access_group_id", ""),
				),
			},
		},
	})
}

func testAccCheckClientVPNAuthorizationRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_client_vpn_authorization_rule" {
				continue
			}

			_, err := tfec2.FindClientVPNAuthorizationRuleByThreePartKey(ctx, conn, rs.Primary.Attributes["client_vpn_endpoint_id"], rs.Primary.Attributes["target_network_cidr"], rs.Primary.Attributes["access_group_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Client VPN Authorization Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClientVPNAuthorizationRuleExists(ctx context.Context, name string, v *awstypes.AuthorizationRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindClientVPNAuthorizationRuleByThreePartKey(ctx, conn, rs.Primary.Attributes["client_vpn_endpoint_id"], rs.Primary.Attributes["target_network_cidr"], rs.Primary.Attributes["access_group_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClientVPNAuthorizationRuleConfig_base(t *testing.T, rName string, subnetCount int) string {
	return acctest.ConfigCompose(
		testAccClientVPNEndpointConfig_basic(t, rName),
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count                   = %[2]d
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}
`, rName, subnetCount))
}

func testAccClientVPNAuthorizationRuleConfig_basic(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccClientVPNAuthorizationRuleConfig_base(t, rName, 1), `
resource "aws_ec2_client_vpn_authorization_rule" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  target_network_cidr    = aws_subnet.test[0].cidr_block
  authorize_all_groups   = true
}
`)
}

func testAccClientVPNAuthorizationRuleConfig_groups(t *testing.T, rName string, groupNames map[string]string) string {
	var b strings.Builder
	for k, v := range groupNames {
		fmt.Fprintf(&b, `
resource "aws_ec2_client_vpn_authorization_rule" %[1]q {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  target_network_cidr    = aws_subnet.test[0].cidr_block
  access_group_id        = %[2]q
}
`, k, v)
	}

	return acctest.ConfigCompose(testAccClientVPNAuthorizationRuleConfig_base(t, rName, 1), b.String())
}

func testAccClientVPNAuthorizationRuleConfig_subnets(t *testing.T, rName string, subnetCount int, groupNames map[string]int) string {
	var b strings.Builder
	for k, v := range groupNames {
		fmt.Fprintf(&b, `
resource "aws_ec2_client_vpn_authorization_rule" %[1]q {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  target_network_cidr    = aws_subnet.test[%[2]d].cidr_block
  authorize_all_groups   = true
}
`, k, v)
	}

	return acctest.ConfigCompose(testAccClientVPNAuthorizationRuleConfig_base(t, rName, subnetCount), b.String())
}
