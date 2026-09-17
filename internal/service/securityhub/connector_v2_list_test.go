// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccConnectorV2_List_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName1 := "aws_securityhub_connector_v2.test[0]"
	resourceName2 := "aws_securityhub_connector_v2.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		CheckDestroy:             testAccCheckConnectorV2Destroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/ConnectorV2/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ConnectorV2/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_securityhub_connector_v2.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_securityhub_connector_v2.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.NotNull()),
					tfquerycheck.ExpectNoResourceObject("aws_securityhub_connector_v2.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_securityhub_connector_v2.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_securityhub_connector_v2.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.NotNull()),
					tfquerycheck.ExpectNoResourceObject("aws_securityhub_connector_v2.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func testAccConnectorV2_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName1 := "aws_securityhub_connector_v2.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity1 := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		CheckDestroy:             testAccCheckConnectorV2Destroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/ConnectorV2/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ConnectorV2/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_securityhub_connector_v2.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_securityhub_connector_v2.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					querycheck.ExpectResourceKnownValues("aws_securityhub_connector_v2.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("securityhub", regexache.MustCompile(`connectorv2/.+`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("connector_id"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("health"), knownvalue.ListSizeExact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
					}),
				},
			},
		},
	})
}

func testAccConnectorV2_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName1 := "aws_securityhub_connector_v2.test[0]"
	resourceName2 := "aws_securityhub_connector_v2.test[1]"
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
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		CheckDestroy:             testAccCheckConnectorV2Destroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/ConnectorV2/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ConnectorV2/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_securityhub_connector_v2.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_securityhub_connector_v2.test", identity2.Checks()),
				},
			},
		},
	})
}
