// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworksecuritymanager "github.com/hashicorp/terraform-provider-aws/internal/service/networksecuritymanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var checkRuleARN knownvalue.Check = tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`rule:[0-9a-z]+$`))

func TestRuleConfigurationEquivalent(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		a, b string
		want bool
	}{
		"identical": {
			a:    `{"DefaultAction":{"Allow":{}}}`,
			b:    `{"DefaultAction":{"Allow":{}}}`,
			want: true,
		},
		"whitespace and key order": {
			a:    `{"VisibilityConfig":{"SampledRequestsEnabled":true,"MetricName":"m"}}`,
			b:    "{ \"VisibilityConfig\": {\n  \"MetricName\": \"m\", \"SampledRequestsEnabled\": true } }",
			want: true,
		},
		"number returned as string": {
			a:    `{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":300}}}`,
			b:    `{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":"300"}}}`,
			want: true,
		},
		"string compared to number": {
			a:    `{"n":"300"}`,
			b:    `{"n":300}`,
			want: true,
		},
		"number in array returned as string": {
			a:    `{"l":[1,2.5]}`,
			b:    `{"l":["1","2.50"]}`,
			want: true,
		},
		"different number": {
			a:    `{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":300}}}`,
			b:    `{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":"301"}}}`,
			want: false,
		},
		"non-numeric string": {
			a:    `{"n":300}`,
			b:    `{"n":"three hundred"}`,
			want: false,
		},
		"bool is not a string": {
			a:    `{"b":true}`,
			b:    `{"b":"true"}`,
			want: false,
		},
		"different value": {
			a:    `{"DefaultAction":{"Allow":{}}}`,
			b:    `{"DefaultAction":{"Block":{}}}`,
			want: false,
		},
		"extra key": {
			a:    `{"DefaultAction":{"Allow":{}}}`,
			b:    `{"DefaultAction":{"Allow":{}},"x":null}`,
			want: false,
		},
		"different array length": {
			a:    `{"l":[1]}`,
			b:    `{"l":[1,1]}`,
			want: false,
		},
		"invalid JSON": {
			a:    `{"DefaultAction":{"Allow":{}}}`,
			b:    `not json`,
			want: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tfnetworksecuritymanager.RuleConfigurationEquivalent(testCase.a, testCase.b); got != testCase.want {
				t.Errorf("RuleConfigurationEquivalent(%q, %q) = %t, want %t", testCase.a, testCase.b, got, testCase.want)
			}
		})
	}
}

func TestAccNetworkSecurityManagerRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Rule/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkRuleARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.StringExact(`{"DefaultAction":{"Allow":{}}}`)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("firewall_type"), tfknownvalue.StringExact(awstypes.RuleFirewallTypeWaf)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.RuleTypeConfiguration)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Rule/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccNetworkSecurityManagerRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Rule/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworksecuritymanager.ResourceRule, resourceName),
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

// Updating description and configuration publishes a new version in place.
// The second configuration contains a number, which the service returns as a
// string; the post-apply empty plan verifies that this is not reported as drift.
func TestAccNetworkSecurityManagerRule_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 networksecuritymanager.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_rule.test"
	const (
		configuration1 = `{"DefaultAction":{"Allow":{}}}`
		configuration2 = `{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":300}}}`
	)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Rule/update/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("initial"),
					"configuration": config.StringVariable(configuration1),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.StringExact(configuration1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("initial")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Rule/update/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("updated"),
					"configuration": config.StringVariable(configuration2),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v2),
					testAccCheckRuleNotRecreated(&v1, &v2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.StringExact(configuration2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("updated")),
				},
			},
		},
	})
}

func testAccCheckRuleExists(ctx context.Context, t *testing.T, n string, v *networksecuritymanager.GetRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		output, err := tfnetworksecuritymanager.FindRuleByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRuleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networksecuritymanager_rule" {
				continue
			}

			ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)
			_, err := tfnetworksecuritymanager.FindRuleByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Network Security Manager Rule %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckRuleNotRecreated(before, after *networksecuritymanager.GetRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.RuleId), aws.ToString(after.RuleId); before != after {
			return fmt.Errorf("Network Security Manager Rule (%s) recreated (%s)", before, after)
		}

		return nil
	}
}

func testAccPreCheckRule(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

	var input networksecuritymanager.ListRulesInput
	_, err := conn.ListRules(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
