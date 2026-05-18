// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTelemetryRuleForOrganization_List_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName1 := "aws_observabilityadmin_telemetry_rule_for_organization.test[0]"
	resourceName2 := "aws_observabilityadmin_telemetry_rule_for_organization.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccTelemetryRuleForOrganizationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		CheckDestroy:             testAccCheckTelemetryRuleForOrganizationDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TelemetryRuleForOrganization/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("rule_arn"), tfknownvalue.RegionalARNExact("observabilityadmin", "organization-telemetry-rule/"+rName+"-0")),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("rule_arn"), tfknownvalue.RegionalARNExact("observabilityadmin", "organization-telemetry-rule/"+rName+"-1")),
				},
			},
		},
	})
}

func testAccTelemetryRuleForOrganization_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName1 := "aws_observabilityadmin_telemetry_rule_for_organization.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccTelemetryRuleForOrganizationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		CheckDestroy:             testAccCheckTelemetryRuleForOrganizationDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TelemetryRuleForOrganization/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("rule_arn"), tfknownvalue.RegionalARNExact("observabilityadmin", "organization-telemetry-rule/"+rName+"-0")),
				},
			},
		},
	})
}

func testAccTelemetryRuleForOrganization_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName1 := "aws_observabilityadmin_telemetry_rule.test[0]"
	resourceName2 := "aws_observabilityadmin_telemetry_rule.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		CheckDestroy:             testAccCheckTelemetryRuleDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/TelemetryRule/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("rule_arn"), tfknownvalue.RegionalARNAlternateRegionExact("observabilityadmin", "telemetry-rule/"+rName+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("rule_arn"), tfknownvalue.RegionalARNAlternateRegionExact("observabilityadmin", "telemetry-rule/"+rName+"-1")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/TelemetryRule/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_observabilityadmin_telemetry_rule.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_observabilityadmin_telemetry_rule.test", identity2.Checks()),
				},
			},
		},
	})
}
