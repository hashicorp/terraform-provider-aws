// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
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
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayPolicyTableEntry/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPolicyTableEntryExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule_number"), knownvalue.StringExact("100")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("target_route_table_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("transit_gateway_policy_table_id"), knownvalue.NotNull()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayPolicyTableEntry/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "transit_gateway_policy_table_id",
				ImportStateIdFunc:                    testAccTransitGatewayPolicyTableEntryImportStateIDFunc(resourceName),
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
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayPolicyTableEntry/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPolicyTableEntryExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceTransitGatewayPolicyTableEntry, resourceName),
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
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayPolicyTableEntry/fullRule/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"destination_cidr_block": knownvalue.StringExact("10.0.2.0/24"),
						"destination_port_range": knownvalue.StringExact("443"),
						"metadata": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"key":   knownvalue.StringExact("test"),
							"value": knownvalue.StringExact("test"),
						})}),
						"protocol":          knownvalue.StringExact("6"),
						"source_cidr_block": knownvalue.StringExact("10.0.1.0/24"),
						"source_port_range": knownvalue.StringExact("1024-65535"),
					})})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayPolicyTableEntry/fullRule/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "transit_gateway_policy_table_id",
				ImportStateIdFunc:                    testAccTransitGatewayPolicyTableEntryImportStateIDFunc(resourceName),
				// The EC2 API doesn't return policy rule metadata via
				// GetTransitGatewayPolicyTableEntries, so it can't be recovered on import.
				ImportStateVerifyIgnore: []string{
					"policy_rule.0.metadata",
				},
			},
		},
	})
}

func testAccTransitGatewayPolicyTableEntry_update(t *testing.T, semaphore tfsync.Semaphore) {
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
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayPolicyTableEntry/protocol/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"protocol":      config.StringVariable("6"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPolicyTableEntryExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey(names.AttrProtocol), knownvalue.StringExact("6")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayPolicyTableEntry/protocol/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"protocol":      config.StringVariable("17"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPolicyTableEntryExists(ctx, t, resourceName, &v),
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
			{
				ConfigDirectory: config.StaticDirectory("testdata/TransitGatewayPolicyTableEntry/protocol/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"protocol":      config.StringVariable("*"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPolicyTableEntryExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey("destination_port_range"), knownvalue.StringExact("*")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey(names.AttrProtocol), knownvalue.StringExact("*")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_rule").AtSliceIndex(0).AtMapKey("source_port_range"), knownvalue.StringExact("*")),
				},
			},
		},
	})
}

func testAccTransitGatewayPolicyTableEntryImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
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
