// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTransitGatewayPolicyTableEntry_identity(t *testing.T, _ tfsync.Semaphore) {
	testAccTransitGatewayTransitGatewayPolicyTableEntry_identitySerial(t)
}

func testAccTransitGatewayPolicyTableEntry_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayPolicyTableEntry
	resourceName := "aws_ec2_transit_gateway_policy_table_entry.test"
	policyTableResourceName := "aws_ec2_transit_gateway_policy_table.test"
	routeTableResourceName := "aws_ec2_transit_gateway_route_table.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayPolicyTableEntryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPolicyTableEntryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPolicyTableEntryExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_policy_table_id", policyTableResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "target_route_table_id", routeTableResourceName, names.AttrID),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule_number"), knownvalue.StringExact("100")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrState), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "transit_gateway_policy_table_id",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "transit_gateway_policy_table_id", "policy_rule_number"),
			},
		},
	})
}

func testAccTransitGatewayPolicyTableEntry_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayPolicyTableEntry
	resourceName := "aws_ec2_transit_gateway_policy_table_entry.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayPolicyTableEntryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPolicyTableEntryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPolicyTableEntryExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappearsWithStateFunc(ctx, t, tfec2.ResourceTransitGatewayPolicyTableEntry, resourceName, func(ctx context.Context, state *tfsdk.State, is *terraform.InstanceState) error {
						v, ok := is.Attributes["transit_gateway_policy_table_id"]
						if !ok {
							return errors.New(`Identifying attribute "transit_gateway_policy_table_id" not defined`)
						}
						state.SetAttribute(ctx, path.Root("transit_gateway_policy_table_id"), v)

						v, ok = is.Attributes["policy_rule_number"]
						if !ok {
							return errors.New(`Identifying attribute "policy_rule_number" not defined`)
						}
						state.SetAttribute(ctx, path.Root("policy_rule_number"), v)

						return nil
					}),
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

func testAccTransitGatewayPolicyTableEntry_fullRule(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayPolicyTableEntry
	resourceName := "aws_ec2_transit_gateway_policy_table_entry.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayPolicyTableEntryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPolicyTableEntryConfig_fullRule(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPolicyTableEntryExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule_number"), knownvalue.StringExact("200")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey("source_cidr_block"), knownvalue.StringExact("10.0.1.0/24")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey("destination_cidr_block"), knownvalue.StringExact("10.0.2.0/24")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey(names.AttrProtocol), knownvalue.StringExact("6")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey("source_port_range"), knownvalue.StringExact("1024-65535")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey("destination_port_range"), knownvalue.StringExact("443")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey("metadata").AtSliceIndex(0).AtMapKey(names.AttrKey), knownvalue.StringExact("test")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey("metadata").AtSliceIndex(0).AtMapKey(names.AttrValue), knownvalue.StringExact("test")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "transit_gateway_policy_table_id",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "transit_gateway_policy_table_id", "policy_rule_number"),
			},
		},
	})
}

func testAccTransitGatewayPolicyTableEntry_update(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayPolicyTableEntry
	resourceName := "aws_ec2_transit_gateway_policy_table_entry.test"
	routeTableResourceName := "aws_ec2_transit_gateway_route_table.test"
	routeTable2ResourceName := "aws_ec2_transit_gateway_route_table.test2"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayPolicyTableEntryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPolicyTableEntryConfig_update(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPolicyTableEntryExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "target_route_table_id", routeTableResourceName, names.AttrID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey(names.AttrProtocol), knownvalue.StringExact("6")),
				},
			},
			{
				Config: testAccTransitGatewayPolicyTableEntryConfig_update(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPolicyTableEntryExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "target_route_table_id", routeTable2ResourceName, names.AttrID),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey(names.AttrProtocol), knownvalue.StringExact("17")),
				},
			},
		},
	})
}

func testAccTransitGatewayPolicyTableEntry_protocolAny(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayPolicyTableEntry
	resourceName := "aws_ec2_transit_gateway_policy_table_entry.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayPolicyTableEntryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPolicyTableEntryConfig_protocolAny(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPolicyTableEntryExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey(names.AttrProtocol), knownvalue.StringExact("*")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey("source_port_range"), knownvalue.StringExact("*")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey("destination_port_range"), knownvalue.StringExact("*")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "transit_gateway_policy_table_id",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "transit_gateway_policy_table_id", "policy_rule_number"),
			},
		},
	})
}

func testAccTransitGatewayPolicyTableEntryImportState(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "transit_gateway_policy_table_id", "policy_rule_number")
}

func testAccCheckTransitGatewayPolicyTableEntryDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_policy_table_entry" {
				continue
			}

			_, err := tfec2.FindTransitGatewayPolicyTableEntryByTwoPartKey(ctx, conn, rs.Primary.Attributes["transit_gateway_policy_table_id"], rs.Primary.Attributes["policy_rule_number"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Transit Gateway Policy Table %s Entry %s still exists", rs.Primary.Attributes["transit_gateway_policy_table_id"], rs.Primary.Attributes["policy_rule_number"])
		}

		return nil
	}
}

func testAccCheckTransitGatewayPolicyTableEntryExists(ctx context.Context, t *testing.T, n string, v *awstypes.TransitGatewayPolicyTableEntry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		output, err := tfec2.FindTransitGatewayPolicyTableEntryByTwoPartKey(ctx, conn, rs.Primary.Attributes["transit_gateway_policy_table_id"], rs.Primary.Attributes["policy_rule_number"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTransitGatewayPolicyTableEntryConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_policy_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTransitGatewayPolicyTableEntryConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayPolicyTableEntryConfig_base(rName),
		`
resource "aws_ec2_transit_gateway_policy_table_entry" "test" {
  transit_gateway_policy_table_id = aws_ec2_transit_gateway_policy_table.test.id
  policy_rule_number              = 100
  target_route_table_id           = aws_ec2_transit_gateway_route_table.test.id
}
`,
	)
}

func testAccTransitGatewayPolicyTableEntryConfig_fullRule(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayPolicyTableEntryConfig_base(rName),
		`
resource "aws_ec2_transit_gateway_policy_table_entry" "test" {
  transit_gateway_policy_table_id = aws_ec2_transit_gateway_policy_table.test.id
  policy_rule_number              = 200
  target_route_table_id           = aws_ec2_transit_gateway_route_table.test.id

  policy_rule {
    source_cidr_block       = "10.0.1.0/24"
    source_port_range       = "1024-65535"
    destination_cidr_block  = "10.0.2.0/24"
    destination_port_range  = "443"
    protocol                = "6"

    metadata {
      key   = "test"
      value = "test"
    }
  }
}
`,
	)
}

func testAccTransitGatewayPolicyTableEntryConfig_update(rName string, useSecondRouteTable bool) string {
	targetRouteTableID := "aws_ec2_transit_gateway_route_table.test.id"
	protocol := "6"
	if useSecondRouteTable {
		targetRouteTableID = "aws_ec2_transit_gateway_route_table.test2.id"
		protocol = "17"
	}

	return acctest.ConfigCompose(
		testAccTransitGatewayPolicyTableEntryConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_route_table" "test2" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_policy_table_entry" "test" {
  transit_gateway_policy_table_id = aws_ec2_transit_gateway_policy_table.test.id
  policy_rule_number              = 300
  target_route_table_id           = %[2]s

  policy_rule {
    source_cidr_block      = "10.0.1.0/24"
    destination_cidr_block = "10.0.2.0/24"
    protocol                = %[3]q
  }
}
`, rName, targetRouteTableID, protocol),
	)
}

func testAccTransitGatewayPolicyTableEntryConfig_protocolAny(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayPolicyTableEntryConfig_base(rName),
		`
resource "aws_ec2_transit_gateway_policy_table_entry" "test" {
  transit_gateway_policy_table_id = aws_ec2_transit_gateway_policy_table.test.id
  policy_rule_number              = 400
  target_route_table_id           = aws_ec2_transit_gateway_route_table.test.id

  policy_rule {
    source_cidr_block      = "10.0.1.0/24"
    destination_cidr_block = "10.0.2.0/24"
    source_port_range      = "*"
    destination_port_range = "*"
    protocol               = "*"
  }
}
`,
	)
}
