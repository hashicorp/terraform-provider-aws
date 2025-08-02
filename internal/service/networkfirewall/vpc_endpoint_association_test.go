// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkFirewallVPCEndpointAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vpcendpointassociation networkfirewall.DescribeVpcEndpointAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_vpc_endpoint_association.test"
	firewallResourceName := "aws_networkfirewall_firewall.test"
	vpcResourceName := "aws_vpc.target"
	subnetResourceName := "aws_subnet.target"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccVPCEndpointAssociationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointAssociationExists(ctx, resourceName, &vpcendpointassociation),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_arn", firewallResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_mapping.0.subnet_id", subnetResourceName, names.AttrID),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "vpc_endpoint_association_status.0.association_sync_state.*", map[string]*regexp.Regexp{
						"attachment.0.endpoint_id": regexache.MustCompile(`vpce-`),
					}),
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

func TestAccNetworkFirewallVPCEndpointAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vpcendpointassociation networkfirewall.DescribeVpcEndpointAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_vpc_endpoint_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointAssociationExists(ctx, resourceName, &vpcendpointassociation),

					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnetworkfirewall.ResourceVPCEndpointAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckVPCEndpointAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_vpc_endpoint_association" {
				continue
			}

			_, err := tfnetworkfirewall.FindVPCEndpointAssociationByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.NetworkFirewall, create.ErrActionCheckingDestroyed, tfnetworkfirewall.ResNameVPCEndpointAssociation, rs.Primary.ID, err)
			}

			return create.Error(names.NetworkFirewall, create.ErrActionCheckingDestroyed, tfnetworkfirewall.ResNameVPCEndpointAssociation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckVPCEndpointAssociationExists(ctx context.Context, name string, vpcendpointassociation *networkfirewall.DescribeVpcEndpointAssociationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.NetworkFirewall, create.ErrActionCheckingExistence, tfnetworkfirewall.ResNameVPCEndpointAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.NetworkFirewall, create.ErrActionCheckingExistence, tfnetworkfirewall.ResNameVPCEndpointAssociation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallClient(ctx)

		resp, err := tfnetworkfirewall.FindVPCEndpointAssociationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.NetworkFirewall, create.ErrActionCheckingExistence, tfnetworkfirewall.ResNameVPCEndpointAssociation, rs.Primary.ID, err)
		}

		*vpcendpointassociation = *resp

		return nil
	}
}

func testAccVPCEndpointAssociationPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallClient(ctx)

	input := &networkfirewall.ListVpcEndpointAssociationsInput{}

	_, err := conn.ListVpcEndpointAssociations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccVPCEndpointAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_baseVPC(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }
}

resource "aws_vpc" "target" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "target" {
  vpc_id            = aws_vpc.target.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.target.cidr_block, 8, 1)

  tags = {
    Name = %[1]q
  }
}
resource "aws_networkfirewall_vpc_endpoint_association" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn
  vpc_id       = aws_vpc.target.id

  subnet_mapping {
    subnet_id = aws_subnet.target.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
