// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAttachmentNATConfigurationValidateRules(t *testing.T) {
	t.Parallel()

	rule := func(n int32, source, nat string) *tfnetworkmanager.AttachmentNATRuleModel {
		return &tfnetworkmanager.AttachmentNATRuleModel{
			MatchSourceCIDR: types.StringValue(source),
			NATCIDR:         types.StringValue(nat),
			RuleNumber:      types.Int32Value(n),
		}
	}

	testCases := []struct {
		testName  string
		rules     []*tfnetworkmanager.AttachmentNATRuleModel
		expectErr bool
	}{
		{
			testName: "empty",
		},
		{
			testName: "one valid rule",
			rules:    []*tfnetworkmanager.AttachmentNATRuleModel{rule(100, "10.0.0.0/16", "100.64.0.0/16")},
		},
		{
			testName: "two valid rules",
			rules: []*tfnetworkmanager.AttachmentNATRuleModel{
				rule(100, "10.0.0.0/24", "100.64.0.0/24"),
				rule(200, "10.0.1.0/24", "100.64.1.0/24"),
			},
		},
		{
			testName:  "prefix length mismatch",
			rules:     []*tfnetworkmanager.AttachmentNATRuleModel{rule(100, "10.0.0.0/16", "100.64.0.0/24")},
			expectErr: true,
		},
		{
			testName: "duplicate rule number",
			rules: []*tfnetworkmanager.AttachmentNATRuleModel{
				rule(100, "10.0.0.0/24", "100.64.0.0/24"),
				rule(100, "10.0.1.0/24", "100.64.1.0/24"),
			},
			expectErr: true,
		},
		{
			testName: "unknown values",
			rules: []*tfnetworkmanager.AttachmentNATRuleModel{
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
			rules: []*tfnetworkmanager.AttachmentNATRuleModel{
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
			rules:    []*tfnetworkmanager.AttachmentNATRuleModel{rule(100, "not-a-cidr", "100.64.0.0/24")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			err := tfnetworkmanager.ValidateAttachmentNATRules(testCase.rules)

			if got, want := err != nil, testCase.expectErr; got != want {
				t.Errorf("ValidateAttachmentNATRules() error = %v, expectErr = %t", err, want)
			}
		})
	}
}

func TestAttachmentNATConfigurationRulesMatch(t *testing.T) {
	t.Parallel()

	rule := func(n int32, source, nat string, state awstypes.NatRuleState) awstypes.NatRule {
		return awstypes.NatRule{
			MatchSourceCidr: aws.String(source),
			NatCidr:         aws.String(nat),
			RuleNumber:      aws.Int32(n),
			State:           state,
		}
	}
	request := func(n int32, source, nat string) awstypes.NatRuleRequest {
		return awstypes.NatRuleRequest{
			MatchSourceCidr: aws.String(source),
			NatCidr:         aws.String(nat),
			RuleNumber:      aws.Int32(n),
		}
	}
	want := []awstypes.NatRuleRequest{
		request(100, "10.0.0.0/24", "100.64.0.0/24"),
		request(200, "10.0.1.0/24", "100.64.1.0/24"),
	}

	testCases := []struct {
		testName string
		got      []awstypes.NatRule
		want     []awstypes.NatRuleRequest
		expected bool
	}{
		{
			testName: "both empty",
			expected: true,
		},
		{
			testName: "same rules different order",
			got: []awstypes.NatRule{
				rule(200, "10.0.1.0/24", "100.64.1.0/24", awstypes.NatRuleStateActivated),
				rule(100, "10.0.0.0/24", "100.64.0.0/24", awstypes.NatRuleStateActivated),
			},
			want:     want,
			expected: true,
		},
		{
			testName: "state differs",
			got: []awstypes.NatRule{
				rule(100, "10.0.0.0/24", "100.64.0.0/24", awstypes.NatRuleStateCreating),
				rule(200, "10.0.1.0/24", "100.64.1.0/24", awstypes.NatRuleStateDeactivated),
			},
			want:     want,
			expected: true,
		},
		{
			testName: "missing rule",
			got: []awstypes.NatRule{
				rule(100, "10.0.0.0/24", "100.64.0.0/24", awstypes.NatRuleStateActivated),
			},
			want: want,
		},
		{
			testName: "extra rule",
			got: []awstypes.NatRule{
				rule(100, "10.0.0.0/24", "100.64.0.0/24", awstypes.NatRuleStateActivated),
				rule(200, "10.0.1.0/24", "100.64.1.0/24", awstypes.NatRuleStateActivated),
				rule(300, "10.0.2.0/24", "100.64.2.0/24", awstypes.NatRuleStateDeleting),
			},
			want: want,
		},
		{
			testName: "changed NAT CIDR",
			got: []awstypes.NatRule{
				rule(100, "10.0.0.0/24", "100.64.0.0/24", awstypes.NatRuleStateActivated),
				rule(200, "10.0.1.0/24", "100.64.2.0/24", awstypes.NatRuleStateActivated),
			},
			want: want,
		},
		{
			testName: "changed match source CIDR",
			got: []awstypes.NatRule{
				rule(100, "10.0.0.0/24", "100.64.0.0/24", awstypes.NatRuleStateActivated),
				rule(200, "10.0.2.0/24", "100.64.1.0/24", awstypes.NatRuleStateActivated),
			},
			want: want,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			if got, want := tfnetworkmanager.AttachmentNATRulesMatch(testCase.got, testCase.want), testCase.expected; got != want {
				t.Errorf("AttachmentNATRulesMatch() = %t, want %t", got, want)
			}
		})
	}
}

func TestAccNetworkManagerAttachmentNATConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v awstypes.AttachmentNatConfiguration
	resourceName := "aws_networkmanager_attachment_nat_configuration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentNATConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentNATConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "attachment_id", "aws_networkmanager_vpc_attachment.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_support"), tfknownvalue.StringExact(awstypes.NatGatewaySupportValueEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_system_route_support"), knownvalue.Bool(true)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccAttachmentNATConfigurationImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "attachment_id",
			},
		},
	})
}

func TestAccNetworkManagerAttachmentNATConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v awstypes.AttachmentNatConfiguration
	resourceName := "aws_networkmanager_attachment_nat_configuration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentNATConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentNATConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworkmanager.ResourceAttachmentNATConfiguration, resourceName),
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

func TestAccNetworkManagerAttachmentNATConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v awstypes.AttachmentNatConfiguration
	resourceName := "aws_networkmanager_attachment_nat_configuration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	twoRulesCheck := knownvalue.SetExact([]knownvalue.Check{
		knownvalue.ObjectExact(map[string]knownvalue.Check{
			"match_source_cidr": knownvalue.StringExact("10.0.0.0/24"),
			"nat_cidr":          knownvalue.StringExact("100.64.10.0/24"),
			"rule_number":       knownvalue.Int32Exact(100),
		}),
		knownvalue.ObjectExact(map[string]knownvalue.Check{
			"match_source_cidr": knownvalue.StringExact("10.0.1.0/24"),
			"nat_cidr":          knownvalue.StringExact("100.64.11.0/24"),
			"rule_number":       knownvalue.Int32Exact(200),
		}),
	})

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttachmentNATConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentNATConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_support"), tfknownvalue.StringExact(awstypes.NatGatewaySupportValueEnable)),
				},
			},
			{
				Config: testAccAttachmentNATConfigurationConfig_update(rName, string(awstypes.NatGatewaySupportValueEnable)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_rule"), twoRulesCheck),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_support"), tfknownvalue.StringExact(awstypes.NatGatewaySupportValueEnable)),
				},
			},
			{
				Config: testAccAttachmentNATConfigurationConfig_update(rName, string(awstypes.NatGatewaySupportValueDisable)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAttachmentNATConfigurationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_rule"), twoRulesCheck),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("nat_gateway_support"), tfknownvalue.StringExact(awstypes.NatGatewaySupportValueDisable)),
				},
			},
		},
	})
}

func testAccCheckAttachmentNATConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_attachment_nat_configuration" {
				continue
			}

			id := rs.Primary.Attributes["attachment_id"]
			_, err := tfnetworkmanager.FindAttachmentNATConfigurationByID(ctx, conn, id)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.NetworkManager, create.ErrActionCheckingDestroyed, "Attachment NAT Configuration", id, err)
			}

			return create.Error(names.NetworkManager, create.ErrActionCheckingDestroyed, "Attachment NAT Configuration", id, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAttachmentNATConfigurationExists(ctx context.Context, t *testing.T, n string, v *awstypes.AttachmentNatConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return create.Error(names.NetworkManager, create.ErrActionCheckingExistence, "Attachment NAT Configuration", n, errors.New("not found"))
		}

		id := rs.Primary.Attributes["attachment_id"]
		if id == "" {
			return create.Error(names.NetworkManager, create.ErrActionCheckingExistence, "Attachment NAT Configuration", n, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		output, err := tfnetworkmanager.FindAttachmentNATConfigurationByID(ctx, conn, id)

		if err != nil {
			return create.Error(names.NetworkManager, create.ErrActionCheckingExistence, "Attachment NAT Configuration", id, err)
		}

		*v = *output

		return nil
	}
}

func testAccAttachmentNATConfigurationImportStateIdFunc(n string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(n, ",", "core_network_id", "attachment_id")
}

func testAccAttachmentNATConfigurationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCAttachmentConfig_basic(rName, false),
		`
resource "aws_networkmanager_attachment_nat_configuration" "test" {
  attachment_id       = aws_networkmanager_vpc_attachment.test.id
  core_network_id     = aws_networkmanager_vpc_attachment.test.core_network_id
  nat_gateway_support = "ENABLE"

  nat_gateway_rule {
    rule_number       = 100
    match_source_cidr = "10.0.0.0/16"
    nat_cidr          = "100.64.0.0/16"
  }
}
`)
}

func testAccAttachmentNATConfigurationConfig_update(rName, natGatewaySupport string) string {
	return acctest.ConfigCompose(
		testAccVPCAttachmentConfig_basic(rName, false),
		fmt.Sprintf(`
resource "aws_networkmanager_attachment_nat_configuration" "test" {
  attachment_id       = aws_networkmanager_vpc_attachment.test.id
  core_network_id     = aws_networkmanager_vpc_attachment.test.core_network_id
  nat_gateway_support = %[1]q

  nat_gateway_rule {
    rule_number       = 100
    match_source_cidr = "10.0.0.0/24"
    nat_cidr          = "100.64.10.0/24"
  }

  nat_gateway_rule {
    rule_number       = 200
    match_source_cidr = "10.0.1.0/24"
    nat_cidr          = "100.64.11.0/24"
  }
}
`, natGatewaySupport))
}
