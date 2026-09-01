// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkFirewallVPCEndpointAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v networkfirewall.DescribeVpcEndpointAssociationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_vpc_endpoint_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccVPCEndpointAssociationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointAssociationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("subnet_mapping"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							names.AttrIPAddressType: tfknownvalue.StringExact(awstypes.IPAddressTypeIpv4),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_endpoint_association_arn"), tfknownvalue.RegionalARNRegexp("network-firewall", regexache.MustCompile(`vpc-endpoint-association/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_endpoint_association_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vpc_endpoint_association_status"), knownvalue.ListSizeExact(1)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "vpc_endpoint_association_arn",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "vpc_endpoint_association_arn"),
			},
		},
	})
}

func TestAccNetworkFirewallVPCEndpointAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v networkfirewall.DescribeVpcEndpointAssociationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_vpc_endpoint_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccVPCEndpointAssociationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointAssociationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworkfirewall.ResourceVPCEndpointAssociation, resourceName),
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

func TestAccNetworkFirewallVPCEndpointAssociation_full(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v networkfirewall.DescribeVpcEndpointAssociationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_vpc_endpoint_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccVPCEndpointAssociationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointAssociationConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointAssociationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("testing")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "vpc_endpoint_association_arn",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "vpc_endpoint_association_arn"),
			},
		},
	})
}

func TestAccNetworkFirewallVPCEndpointAssociation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v networkfirewall.DescribeVpcEndpointAssociationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_vpc_endpoint_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccVPCEndpointAssociationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCEndpointAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointAssociationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointAssociationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "vpc_endpoint_association_arn",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "vpc_endpoint_association_arn"),
			},
			{
				Config: testAccVPCEndpointAssociationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointAssociationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccVPCEndpointAssociationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCEndpointAssociationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckVPCEndpointAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_vpc_endpoint_association" {
				continue
			}

			_, err := tfnetworkfirewall.FindVPCEndpointAssociationByARN(ctx, conn, rs.Primary.Attributes["vpc_endpoint_association_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("NetworkFirewall VPC Endpoint Association %s still exists", rs.Primary.Attributes["vpc_endpoint_association_arn"])
		}

		return nil
	}
}

func testAccCheckVPCEndpointAssociationExists(ctx context.Context, t *testing.T, n string, v *networkfirewall.DescribeVpcEndpointAssociationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

		output, err := tfnetworkfirewall.FindVPCEndpointAssociationByARN(ctx, conn, rs.Primary.Attributes["vpc_endpoint_association_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCEndpointAssociationPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

	input := &networkfirewall.ListVpcEndpointAssociationsInput{}

	_, err := conn.ListVpcEndpointAssociations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccVPCEndpointAssociationConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_basic(rName), fmt.Sprintf(`
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
`, rName))
}

func testAccVPCEndpointAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointAssociationConfig_base(rName), `
resource "aws_networkfirewall_vpc_endpoint_association" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn
  vpc_id       = aws_vpc.target.id

  subnet_mapping {
    subnet_id = aws_subnet.target.id
  }
}
`)
}

func testAccVPCEndpointAssociationConfig_full(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointAssociationConfig_base(rName), `
resource "aws_networkfirewall_vpc_endpoint_association" "test" {
  description  = "testing"
  firewall_arn = aws_networkfirewall_firewall.test.arn
  vpc_id       = aws_vpc.target.id

  subnet_mapping {
    ip_address_type = "IPV4"
    subnet_id       = aws_subnet.target.id
  }
}
`)
}

func testAccVPCEndpointAssociationConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccVPCEndpointAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_networkfirewall_vpc_endpoint_association" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn
  vpc_id       = aws_vpc.target.id

  subnet_mapping {
    subnet_id = aws_subnet.target.id
  }

  tags = {
    %[1]q = %[2]q
  }
}
`, tag1Key, tag1Value))
}

func testAccVPCEndpointAssociationConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccVPCEndpointAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_networkfirewall_vpc_endpoint_association" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn
  vpc_id       = aws_vpc.target.id

  subnet_mapping {
    subnet_id = aws_subnet.target.id
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tag1Key, tag1Value, tag2Key, tag2Value))
}
