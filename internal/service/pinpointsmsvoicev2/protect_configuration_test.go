// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"context"
	"fmt"
	"maps"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfpinpointsmsvoicev2 "github.com/hashicorp/terraform-provider-aws/internal/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestProtectConfigurationCountryRuleSetUpdates(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName      string
		current       map[string]string
		defaultStatus string
		rules         map[string]string
		want          map[string]awstypes.ProtectStatus
	}{
		{
			testName:      "no changes",
			current:       map[string]string{"CA": "ALLOW", "US": "BLOCK"},
			defaultStatus: "ALLOW",
			rules:         map[string]string{"US": "BLOCK"},
			want:          map[string]awstypes.ProtectStatus{},
		},
		{
			testName:      "removed country resets to default",
			current:       map[string]string{"CA": "BLOCK", "US": "BLOCK"},
			defaultStatus: "ALLOW",
			rules:         map[string]string{"US": "BLOCK"},
			want:          map[string]awstypes.ProtectStatus{"CA": "ALLOW"},
		},
		{
			testName:      "new default",
			current:       map[string]string{"CA": "ALLOW", "DE": "ALLOW", "US": "ALLOW"},
			defaultStatus: "BLOCK",
			rules:         map[string]string{"US": "ALLOW"},
			want:          map[string]awstypes.ProtectStatus{"CA": "BLOCK", "DE": "BLOCK"},
		},
		{
			testName:      "unsupported country is sent",
			current:       map[string]string{"US": "ALLOW"},
			defaultStatus: "ALLOW",
			rules:         map[string]string{"ZZ": "BLOCK"},
			want:          map[string]awstypes.ProtectStatus{"ZZ": "BLOCK"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			got := make(map[string]awstypes.ProtectStatus)
			for country, v := range tfpinpointsmsvoicev2.ProtectConfigurationCountryRuleSetUpdates(testCase.current, testCase.defaultStatus, testCase.rules) {
				got[country] = v.ProtectStatus
			}

			if !maps.Equal(got, testCase.want) {
				t.Errorf("got %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestAccPinpointSMSVoiceV2ProtectConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ProtectConfigurationInformation
	resourceName := "aws_pinpointsmsvoicev2_protect_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckProtectConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectConfigurationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "created_timestamp"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("account_default"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("sms-voice", regexache.MustCompile(`protect-configuration/protect-[0-9a-f]{32}$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.StringRegexp(regexache.MustCompile(`^protect-[0-9a-f]{32}$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("mms_country_rules"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rcs_country_rules"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sms_country_rules"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("voice_country_rules"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("mms_default_status"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rcs_default_status"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sms_default_status"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("voice_default_status"), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ProtectConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ProtectConfigurationInformation
	resourceName := "aws_pinpointsmsvoicev2_protect_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckProtectConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectConfigurationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectConfigurationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfpinpointsmsvoicev2.ResourceProtectConfiguration, resourceName),
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

func TestAccPinpointSMSVoiceV2ProtectConfiguration_countryRules(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ProtectConfigurationInformation
	resourceName := "aws_pinpointsmsvoicev2_protect_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckProtectConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectConfigurationConfig_countryRules(`
  sms_country_rules = {
    US = "BLOCK"
    CA = "FILTER"
    DE = "MONITOR"
  }
  voice_country_rules = {
    US = "BLOCK"
  }
  mms_country_rules = {
    CA = "BLOCK"
  }
  rcs_country_rules = {
    DE = "BLOCK"
  }
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectConfigurationExists(ctx, t, resourceName, &v),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilitySms, awstypes.ProtectStatusAllow, map[string]string{"US": "BLOCK", "CA": "FILTER", "DE": "MONITOR"}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityVoice, awstypes.ProtectStatusAllow, map[string]string{"US": "BLOCK"}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityMms, awstypes.ProtectStatusAllow, map[string]string{"CA": "BLOCK"}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityRcs, awstypes.ProtectStatusAllow, map[string]string{"DE": "BLOCK"}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sms_country_rules"), knownvalue.MapExact(map[string]knownvalue.Check{
						"CA": knownvalue.StringExact("FILTER"),
						"DE": knownvalue.StringExact("MONITOR"),
						"US": knownvalue.StringExact("BLOCK"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("voice_country_rules"), knownvalue.MapExact(map[string]knownvalue.Check{
						"US": knownvalue.StringExact("BLOCK"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("mms_country_rules"), knownvalue.MapExact(map[string]knownvalue.Check{
						"CA": knownvalue.StringExact("BLOCK"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rcs_country_rules"), knownvalue.MapExact(map[string]knownvalue.Check{
						"DE": knownvalue.StringExact("BLOCK"),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Change a status, add a country, remove a country, and remove all VOICE rules.
				Config: testAccProtectConfigurationConfig_countryRules(`
  sms_country_rules = {
    US = "MONITOR"
    DE = "MONITOR"
    GB = "BLOCK"
  }
  mms_country_rules = {
    CA = "BLOCK"
  }
  rcs_country_rules = {
    DE = "BLOCK"
    FR = "BLOCK"
  }
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectConfigurationExists(ctx, t, resourceName, &v),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilitySms, awstypes.ProtectStatusAllow, map[string]string{"US": "MONITOR", "DE": "MONITOR", "GB": "BLOCK"}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityVoice, awstypes.ProtectStatusAllow, map[string]string{}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityMms, awstypes.ProtectStatusAllow, map[string]string{"CA": "BLOCK"}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityRcs, awstypes.ProtectStatusAllow, map[string]string{"DE": "BLOCK", "FR": "BLOCK"}),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sms_country_rules"), knownvalue.MapExact(map[string]knownvalue.Check{
						"DE": knownvalue.StringExact("MONITOR"),
						"GB": knownvalue.StringExact("BLOCK"),
						"US": knownvalue.StringExact("MONITOR"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("voice_country_rules"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("mms_country_rules"), knownvalue.MapExact(map[string]knownvalue.Check{
						"CA": knownvalue.StringExact("BLOCK"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rcs_country_rules"), knownvalue.MapExact(map[string]knownvalue.Check{
						"DE": knownvalue.StringExact("BLOCK"),
						"FR": knownvalue.StringExact("BLOCK"),
					})),
				},
			},
			{
				Config: testAccProtectConfigurationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectConfigurationExists(ctx, t, resourceName, &v),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilitySms, awstypes.ProtectStatusAllow, map[string]string{}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityVoice, awstypes.ProtectStatusAllow, map[string]string{}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityMms, awstypes.ProtectStatusAllow, map[string]string{}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityRcs, awstypes.ProtectStatusAllow, map[string]string{}),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("mms_country_rules"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rcs_country_rules"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sms_country_rules"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("voice_country_rules"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ProtectConfiguration_defaultStatus(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ProtectConfigurationInformation
	resourceName := "aws_pinpointsmsvoicev2_protect_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckProtectConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectConfigurationConfig_countryRules(`
  sms_default_status = "BLOCK"
  sms_country_rules = {
    US = "ALLOW"
    CA = "FILTER"
  }
  voice_default_status = "BLOCK"
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectConfigurationExists(ctx, t, resourceName, &v),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilitySms, awstypes.ProtectStatusBlock, map[string]string{"US": "ALLOW", "CA": "FILTER"}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityVoice, awstypes.ProtectStatusBlock, map[string]string{}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityMms, awstypes.ProtectStatusAllow, map[string]string{}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityRcs, awstypes.ProtectStatusAllow, map[string]string{}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sms_country_rules"), knownvalue.MapExact(map[string]knownvalue.Check{
						"CA": knownvalue.StringExact("FILTER"),
						"US": knownvalue.StringExact("ALLOW"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sms_default_status"), knownvalue.StringExact("BLOCK")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("voice_country_rules"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("voice_default_status"), knownvalue.StringExact("BLOCK")),
				},
			},
			{
				// AWS doesn't store a default status, so import reads every country that isn't ALLOW as a country rule.
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"sms_country_rules", "sms_default_status", "voice_country_rules", "voice_default_status"},
			},
			{
				// Change the default, remove the default, and add a default.
				Config: testAccProtectConfigurationConfig_countryRules(`
  sms_default_status = "MONITOR"
  sms_country_rules = {
    US = "ALLOW"
  }
  mms_default_status = "BLOCK"
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectConfigurationExists(ctx, t, resourceName, &v),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilitySms, awstypes.ProtectStatusMonitor, map[string]string{"US": "ALLOW"}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityVoice, awstypes.ProtectStatusAllow, map[string]string{}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityMms, awstypes.ProtectStatusBlock, map[string]string{}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityRcs, awstypes.ProtectStatusAllow, map[string]string{}),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("mms_default_status"), knownvalue.StringExact("BLOCK")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sms_country_rules"), knownvalue.MapExact(map[string]knownvalue.Check{
						"US": knownvalue.StringExact("ALLOW"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sms_default_status"), knownvalue.StringExact("MONITOR")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("voice_default_status"), knownvalue.Null()),
				},
			},
			{
				Config: testAccProtectConfigurationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectConfigurationExists(ctx, t, resourceName, &v),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilitySms, awstypes.ProtectStatusAllow, map[string]string{}),
					testAccCheckProtectConfigurationCountryRules(ctx, t, resourceName, awstypes.NumberCapabilityMms, awstypes.ProtectStatusAllow, map[string]string{}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("mms_default_status"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sms_country_rules"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sms_default_status"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ProtectConfiguration_deletionProtectionEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ProtectConfigurationInformation
	resourceName := "aws_pinpointsmsvoicev2_protect_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckProtectConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectConfigurationConfig_deletionProtectionEnabled(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectConfigurationExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(true)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProtectConfigurationConfig_deletionProtectionEnabled(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProtectConfigurationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ProtectConfiguration_invalidCountryRules(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				// ALLOW is the default when no default status is set, so it is expressed by omitting the country.
				Config: testAccProtectConfigurationConfig_countryRules(`
  sms_country_rules = {
    US = "ALLOW"
  }
`),
				ExpectError: regexache.MustCompile(`Invalid Country Rule`),
			},
			{
				Config: testAccProtectConfigurationConfig_countryRules(`
  voice_default_status = "BLOCK"
  voice_country_rules = {
    US = "BLOCK"
  }
`),
				ExpectError: regexache.MustCompile(`Invalid Country Rule`),
			},
			{
				// VOICE, MMS, and RCS only support ALLOW and BLOCK.
				Config: testAccProtectConfigurationConfig_countryRules(`
  mms_default_status = "MONITOR"
`),
				ExpectError: regexache.MustCompile(`value must be one of`),
			},
			{
				Config: testAccProtectConfigurationConfig_countryRules(`
  voice_country_rules = {
    US = "FILTER"
  }
`),
				ExpectError: regexache.MustCompile(`value must be one of`),
			},
			{
				Config: testAccProtectConfigurationConfig_countryRules(`
  sms_country_rules = {
    us = "BLOCK"
  }
`),
				ExpectError: regexache.MustCompile(`ISO 3166-1 alpha-2`),
			},
			{
				Config: testAccProtectConfigurationConfig_countryRules(`
  sms_country_rules = {}
`),
				ExpectError: regexache.MustCompile(`map must contain at least 1 elements`),
			},
			{
				Config: testAccProtectConfigurationConfig_countryRules(`
  sms_country_rules = {
    US = null
  }
`),
				ExpectError: regexache.MustCompile(`null value`),
			},
		},
	})
}

func testAccCheckProtectConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpointsmsvoicev2_protect_configuration" {
				continue
			}

			ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)
			_, err := tfpinpointsmsvoicev2.FindProtectConfigurationByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("End User Messaging SMS Protect Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProtectConfigurationExists(ctx context.Context, t *testing.T, n string, v *awstypes.ProtectConfigurationInformation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		output, err := tfpinpointsmsvoicev2.FindProtectConfigurationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// testAccCheckProtectConfigurationCountryRules checks the rules stored by AWS, independent of state.
// want lists the countries whose status differs from defaultStatus.
func testAccCheckProtectConfigurationCountryRules(ctx context.Context, t *testing.T, n string, capability awstypes.NumberCapability, defaultStatus awstypes.ProtectStatus, want map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		ruleSet, err := tfpinpointsmsvoicev2.FindProtectConfigurationCountryRuleSet(ctx, conn, rs.Primary.ID, capability)
		if err != nil {
			return err
		}

		got := make(map[string]string)
		for country, status := range ruleSet {
			if status != string(defaultStatus) {
				got[country] = status
			}
		}

		if !maps.Equal(got, want) {
			return fmt.Errorf("End User Messaging SMS Protect Configuration %s %s country rules: got %v, want %v", rs.Primary.ID, capability, got, want)
		}

		return nil
	}
}

func testAccPreCheckProtectConfiguration(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

	var input pinpointsmsvoicev2.DescribeProtectConfigurationsInput
	_, err := conn.DescribeProtectConfigurations(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccProtectConfigurationConfig_basic() string {
	return `
resource "aws_pinpointsmsvoicev2_protect_configuration" "test" {}
`
}

func testAccProtectConfigurationConfig_countryRules(rules string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_protect_configuration" "test" {
%[1]s
}
`, rules)
}

func testAccProtectConfigurationConfig_deletionProtectionEnabled(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_protect_configuration" "test" {
  deletion_protection_enabled = %[1]t
}
`, enabled)
}
