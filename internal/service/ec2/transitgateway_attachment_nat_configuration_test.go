// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestTransitGatewayAttachmentNATConfigurationValidateRules(t *testing.T) {
	t.Parallel()

	rule := func(n int32, source, nat string) *tfec2.TransitGatewayNATRuleModel {
		return &tfec2.TransitGatewayNATRuleModel{
			MatchSourceCIDR: types.StringValue(source),
			NATCIDR:         types.StringValue(nat),
			RuleNumber:      types.Int32Value(n),
		}
	}

	testCases := []struct {
		testName  string
		rules     []*tfec2.TransitGatewayNATRuleModel
		expectErr bool
	}{
		{
			testName: "empty",
		},
		{
			testName: "one valid rule",
			rules:    []*tfec2.TransitGatewayNATRuleModel{rule(100, "10.0.0.0/16", "100.64.0.0/16")},
		},
		{
			testName: "two valid rules",
			rules: []*tfec2.TransitGatewayNATRuleModel{
				rule(100, "10.0.0.0/24", "100.64.0.0/24"),
				rule(200, "10.0.1.0/24", "100.64.1.0/24"),
			},
		},
		{
			testName:  "prefix length mismatch",
			rules:     []*tfec2.TransitGatewayNATRuleModel{rule(100, "10.0.0.0/16", "100.64.0.0/24")},
			expectErr: true,
		},
		{
			testName: "duplicate rule number",
			rules: []*tfec2.TransitGatewayNATRuleModel{
				rule(100, "10.0.0.0/24", "100.64.0.0/24"),
				rule(100, "10.0.1.0/24", "100.64.1.0/24"),
			},
			expectErr: true,
		},
		{
			testName: "unknown values",
			rules: []*tfec2.TransitGatewayNATRuleModel{
				{
					MatchSourceCIDR: types.StringUnknown(),
					NATCIDR:         types.StringValue("100.64.0.0/24"),
					RuleNumber:      types.Int32Unknown(),
				},
				{
					MatchSourceCIDR: types.StringValue("10.0.0.0/16"),
					NATCIDR:         types.StringUnknown(),
					RuleNumber:      types.Int32Unknown(),
				},
			},
		},
		{
			testName: "null values",
			rules: []*tfec2.TransitGatewayNATRuleModel{
				{
					MatchSourceCIDR: types.StringNull(),
					NATCIDR:         types.StringNull(),
					RuleNumber:      types.Int32Null(),
				},
				{
					MatchSourceCIDR: types.StringNull(),
					NATCIDR:         types.StringNull(),
					RuleNumber:      types.Int32Null(),
				},
			},
		},
		{
			testName: "unparsable CIDR",
			rules:    []*tfec2.TransitGatewayNATRuleModel{rule(100, "not-a-cidr", "100.64.0.0/24")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			err := tfec2.ValidateTransitGatewayNATRules(testCase.rules)

			if got, want := err != nil, testCase.expectErr; got != want {
				t.Errorf("ValidateTransitGatewayNATRules() error = %v, expectErr = %t", err, want)
			}
		})
	}
}

func TestTransitGatewayAttachmentNATConfigurationRulesMatch(t *testing.T) {
	t.Parallel()

	rule := func(n int32, source, nat string, state awstypes.TransitGatewayNatRuleState) awstypes.TransitGatewayNatRule {
		return awstypes.TransitGatewayNatRule{
			MatchSourceCidr: aws.String(source),
			NatCidr:         aws.String(nat),
			RuleNumber:      aws.Int32(n),
			State:           state,
		}
	}
	request := func(n int32, source, nat string) awstypes.TransitGatewayNatRuleRequest {
		return awstypes.TransitGatewayNatRuleRequest{
			MatchSourceCidr: aws.String(source),
			NatCidr:         aws.String(nat),
			RuleNumber:      aws.Int32(n),
		}
	}
	want := []awstypes.TransitGatewayNatRuleRequest{
		request(100, "10.0.0.0/24", "100.64.0.0/24"),
		request(200, "10.0.1.0/24", "100.64.1.0/24"),
	}

	testCases := []struct {
		testName string
		got      []awstypes.TransitGatewayNatRule
		want     []awstypes.TransitGatewayNatRuleRequest
		expected bool
	}{
		{
			testName: "both empty",
			expected: true,
		},
		{
			testName: "same rules different order",
			got: []awstypes.TransitGatewayNatRule{
				rule(200, "10.0.1.0/24", "100.64.1.0/24", awstypes.TransitGatewayNatRuleStateActivated),
				rule(100, "10.0.0.0/24", "100.64.0.0/24", awstypes.TransitGatewayNatRuleStateActivated),
			},
			want:     want,
			expected: true,
		},
		{
			testName: "state differs",
			got: []awstypes.TransitGatewayNatRule{
				rule(100, "10.0.0.0/24", "100.64.0.0/24", awstypes.TransitGatewayNatRuleStatePending),
				rule(200, "10.0.1.0/24", "100.64.1.0/24", awstypes.TransitGatewayNatRuleStateDeactivated),
			},
			want:     want,
			expected: true,
		},
		{
			testName: "missing rule",
			got: []awstypes.TransitGatewayNatRule{
				rule(100, "10.0.0.0/24", "100.64.0.0/24", awstypes.TransitGatewayNatRuleStateActivated),
			},
			want: want,
		},
		{
			testName: "extra rule",
			got: []awstypes.TransitGatewayNatRule{
				rule(100, "10.0.0.0/24", "100.64.0.0/24", awstypes.TransitGatewayNatRuleStateActivated),
				rule(200, "10.0.1.0/24", "100.64.1.0/24", awstypes.TransitGatewayNatRuleStateActivated),
				rule(300, "10.0.2.0/24", "100.64.2.0/24", awstypes.TransitGatewayNatRuleStateDeleting),
			},
			want: want,
		},
		{
			testName: "changed NAT CIDR",
			got: []awstypes.TransitGatewayNatRule{
				rule(100, "10.0.0.0/24", "100.64.0.0/24", awstypes.TransitGatewayNatRuleStateActivated),
				rule(200, "10.0.1.0/24", "100.64.2.0/24", awstypes.TransitGatewayNatRuleStateActivated),
			},
			want: want,
		},
		{
			testName: "changed match source CIDR",
			got: []awstypes.TransitGatewayNatRule{
				rule(100, "10.0.0.0/24", "100.64.0.0/24", awstypes.TransitGatewayNatRuleStateActivated),
				rule(200, "10.0.2.0/24", "100.64.1.0/24", awstypes.TransitGatewayNatRuleStateActivated),
			},
			want: want,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			if got, want := tfec2.TransitGatewayNATRulesMatch(testCase.got, testCase.want), testCase.expected; got != want {
				t.Errorf("TransitGatewayNATRulesMatch() = %t, want %t", got, want)
			}
		})
	}
}

func testAccTransitGatewayAttachmentNATConfiguration_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v ec2.GetTransitGatewayAttachmentNatConfigurationOutput
	resourceName := "aws_ec2_transit_gateway_attachment_nat_configuration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayAttachmentNATConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayAttachmentNATConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, "aws_ec2_transit_gateway_vpc_attachment.test", names.AttrID),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_rule"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"match_source_cidr": knownvalue.StringExact("10.0.0.0/16"),
							"nat_cidr":          knownvalue.StringExact("100.64.0.0/16"),
							"rule_number":       knownvalue.Int32Exact(100),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_support"), tfknownvalue.StringExact(awstypes.TransitGatewayAttachmentNatGatewaySupportValueEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_system_route_support"), tfknownvalue.StringExact(awstypes.TransitGatewayAttachmentNatSystemRouteSupportValueEnable)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrTransitGatewayAttachmentID),
				ImportStateVerifyIdentifierAttribute: names.AttrTransitGatewayAttachmentID,
			},
		},
	})
}

func testAccTransitGatewayAttachmentNATConfiguration_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v ec2.GetTransitGatewayAttachmentNatConfigurationOutput
	resourceName := "aws_ec2_transit_gateway_attachment_nat_configuration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayAttachmentNATConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayAttachmentNATConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceTransitGatewayAttachmentNATConfiguration, resourceName),
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

func testAccTransitGatewayAttachmentNATConfiguration_updateRules(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v ec2.GetTransitGatewayAttachmentNatConfigurationOutput
	resourceName := "aws_ec2_transit_gateway_attachment_nat_configuration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayAttachmentNATConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayAttachmentNATConfigurationConfig_rules1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_rule"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"match_source_cidr": knownvalue.StringExact("10.0.0.0/24"),
							"nat_cidr":          knownvalue.StringExact("100.64.0.0/24"),
							"rule_number":       knownvalue.Int32Exact(100),
						}),
					})),
				},
			},
			{
				Config: testAccTransitGatewayAttachmentNATConfigurationConfig_rules2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_rule"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"match_source_cidr": knownvalue.StringExact("10.0.0.0/24"),
							"nat_cidr":          knownvalue.StringExact("100.64.0.0/24"),
							"rule_number":       knownvalue.Int32Exact(100),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"match_source_cidr": knownvalue.StringExact("10.0.1.0/24"),
							"nat_cidr":          knownvalue.StringExact("100.64.1.0/24"),
							"rule_number":       knownvalue.Int32Exact(200),
						}),
					})),
				},
			},
			{
				Config: testAccTransitGatewayAttachmentNATConfigurationConfig_rules3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_rule"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"match_source_cidr": knownvalue.StringExact("10.0.1.0/24"),
							"nat_cidr":          knownvalue.StringExact("100.64.2.0/24"),
							"rule_number":       knownvalue.Int32Exact(200),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrTransitGatewayAttachmentID),
				ImportStateVerifyIdentifierAttribute: names.AttrTransitGatewayAttachmentID,
			},
		},
	})
}

func testAccTransitGatewayAttachmentNATConfiguration_natGatewaySupport(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v ec2.GetTransitGatewayAttachmentNatConfigurationOutput
	resourceName := "aws_ec2_transit_gateway_attachment_nat_configuration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	ruleCheck := knownvalue.SetExact([]knownvalue.Check{
		knownvalue.ObjectExact(map[string]knownvalue.Check{
			"match_source_cidr": knownvalue.StringExact("10.0.0.0/16"),
			"nat_cidr":          knownvalue.StringExact("100.64.0.0/16"),
			"rule_number":       knownvalue.Int32Exact(100),
		}),
	})

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayAttachmentNATConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayAttachmentNATConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
					testAccCheckTransitGatewayAttachmentNATConfigurationStates(&v, awstypes.TransitGatewayAttachmentNatGatewaySupportStateEnabled, awstypes.TransitGatewayNatRuleStateActivated),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_rule"), ruleCheck),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_support"), tfknownvalue.StringExact(awstypes.TransitGatewayAttachmentNatGatewaySupportValueEnable)),
				},
			},
			{
				Config: testAccTransitGatewayAttachmentNATConfigurationConfig_natGatewaySupport(rName, string(awstypes.TransitGatewayAttachmentNatGatewaySupportValueDisable)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
					testAccCheckTransitGatewayAttachmentNATConfigurationStates(&v, awstypes.TransitGatewayAttachmentNatGatewaySupportStateDisabled, awstypes.TransitGatewayNatRuleStateDeactivated),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_rule"), ruleCheck),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_support"), tfknownvalue.StringExact(awstypes.TransitGatewayAttachmentNatGatewaySupportValueDisable)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrTransitGatewayAttachmentID),
				ImportStateVerifyIdentifierAttribute: names.AttrTransitGatewayAttachmentID,
			},
			{
				Config: testAccTransitGatewayAttachmentNATConfigurationConfig_natGatewaySupport(rName, string(awstypes.TransitGatewayAttachmentNatGatewaySupportValueEnable)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
					testAccCheckTransitGatewayAttachmentNATConfigurationStates(&v, awstypes.TransitGatewayAttachmentNatGatewaySupportStateEnabled, awstypes.TransitGatewayNatRuleStateActivated),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_rule"), ruleCheck),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_support"), tfknownvalue.StringExact(awstypes.TransitGatewayAttachmentNatGatewaySupportValueEnable)),
				},
			},
		},
	})
}

func testAccTransitGatewayAttachmentNATConfiguration_natSystemRouteSupport(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v ec2.GetTransitGatewayAttachmentNatConfigurationOutput
	resourceName := "aws_ec2_transit_gateway_attachment_nat_configuration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayAttachmentNATConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// VPC attachments always install NAT routes.
				Config: testAccTransitGatewayAttachmentNATConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_system_route_support"), tfknownvalue.StringExact(awstypes.TransitGatewayAttachmentNatSystemRouteSupportValueEnable)),
				},
			},
			{
				Config: testAccTransitGatewayAttachmentNATConfigurationConfig_natSystemRouteSupport(rName, string(awstypes.TransitGatewayAttachmentNatSystemRouteSupportValueEnable)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_system_route_support"), tfknownvalue.StringExact(awstypes.TransitGatewayAttachmentNatSystemRouteSupportValueEnable)),
				},
			},
			{
				// Removing the argument keeps the current value.
				Config: testAccTransitGatewayAttachmentNATConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_system_route_support"), tfknownvalue.StringExact(awstypes.TransitGatewayAttachmentNatSystemRouteSupportValueEnable)),
				},
			},
		},
	})
}

func testAccTransitGatewayAttachmentNATConfiguration_identity(t *testing.T, _ tfsync.Semaphore) {
	testAccTransitGatewayTransitGatewayAttachmentNATConfiguration_identitySerial(t)
}

func testAccCheckTransitGatewayAttachmentNATConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_attachment_nat_configuration" {
				continue
			}

			id := rs.Primary.Attributes[names.AttrTransitGatewayAttachmentID]
			_, err := tfec2.FindTransitGatewayAttachmentNATConfigurationByID(ctx, conn, id)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, "Transit Gateway Attachment NAT Configuration", id, err)
			}

			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, "Transit Gateway Attachment NAT Configuration", id, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTransitGatewayAttachmentNATConfigurationExists(ctx context.Context, t *testing.T, n string, v *ec2.GetTransitGatewayAttachmentNatConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, "Transit Gateway Attachment NAT Configuration", n, errors.New("not found"))
		}

		id := rs.Primary.Attributes[names.AttrTransitGatewayAttachmentID]
		if id == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, "Transit Gateway Attachment NAT Configuration", n, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		output, err := tfec2.FindTransitGatewayAttachmentNATConfigurationByID(ctx, conn, id)

		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, "Transit Gateway Attachment NAT Configuration", id, err)
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayAttachmentNATConfigurationStates(v *ec2.GetTransitGatewayAttachmentNatConfigurationOutput, support awstypes.TransitGatewayAttachmentNatGatewaySupportState, ruleState awstypes.TransitGatewayNatRuleState) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if got, want := v.NatGatewaySupport, support; got != want {
			return fmt.Errorf("NAT gateway support = %s, want %s", got, want)
		}

		for _, rule := range v.NatGatewayRules {
			if got, want := rule.State, ruleState; got != want {
				return fmt.Errorf("NAT gateway rule %d state = %s, want %s", aws.ToInt32(rule.RuleNumber), got, want)
			}
		}

		return nil
	}
}

func testAccTransitGatewayAttachmentNATConfigurationConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = aws_subnet.test[*].id
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTransitGatewayAttachmentNATConfigurationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayAttachmentNATConfigurationConfig_base(rName),
		`
resource "aws_ec2_transit_gateway_attachment_nat_configuration" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
  nat_gateway_support           = "enable"

  nat_gateway_rule {
    rule_number       = 100
    match_source_cidr = "10.0.0.0/16"
    nat_cidr          = "100.64.0.0/16"
  }
}
`)
}

func testAccTransitGatewayAttachmentNATConfigurationConfig_rules1(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayAttachmentNATConfigurationConfig_base(rName),
		`
resource "aws_ec2_transit_gateway_attachment_nat_configuration" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
  nat_gateway_support           = "enable"

  nat_gateway_rule {
    rule_number       = 100
    match_source_cidr = "10.0.0.0/24"
    nat_cidr          = "100.64.0.0/24"
  }
}
`)
}

func testAccTransitGatewayAttachmentNATConfigurationConfig_rules2(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayAttachmentNATConfigurationConfig_base(rName),
		`
resource "aws_ec2_transit_gateway_attachment_nat_configuration" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
  nat_gateway_support           = "enable"

  nat_gateway_rule {
    rule_number       = 100
    match_source_cidr = "10.0.0.0/24"
    nat_cidr          = "100.64.0.0/24"
  }

  nat_gateway_rule {
    rule_number       = 200
    match_source_cidr = "10.0.1.0/24"
    nat_cidr          = "100.64.1.0/24"
  }
}
`)
}

func testAccTransitGatewayAttachmentNATConfigurationConfig_rules3(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayAttachmentNATConfigurationConfig_base(rName),
		`
resource "aws_ec2_transit_gateway_attachment_nat_configuration" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
  nat_gateway_support           = "enable"

  nat_gateway_rule {
    rule_number       = 200
    match_source_cidr = "10.0.1.0/24"
    nat_cidr          = "100.64.2.0/24"
  }
}
`)
}

func testAccTransitGatewayAttachmentNATConfigurationConfig_natGatewaySupport(rName, support string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayAttachmentNATConfigurationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_attachment_nat_configuration" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
  nat_gateway_support           = %[1]q

  nat_gateway_rule {
    rule_number       = 100
    match_source_cidr = "10.0.0.0/16"
    nat_cidr          = "100.64.0.0/16"
  }
}
`, support))
}

func testAccTransitGatewayAttachmentNATConfigurationConfig_natSystemRouteSupport(rName, value string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayAttachmentNATConfigurationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_attachment_nat_configuration" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
  nat_gateway_support           = "enable"
  nat_system_route_support      = %[1]q

  nat_gateway_rule {
    rule_number       = 100
    match_source_cidr = "10.0.0.0/16"
    nat_cidr          = "100.64.0.0/16"
  }
}
`, value))
}
