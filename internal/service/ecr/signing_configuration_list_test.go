// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"testing"

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

// A registry has at most one signing configuration per Region, so each query returns one result.

func testAccECRSigningConfiguration_List_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_signing_configuration.test"
	rName := testAccSigningProfileName(t)
	identity := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSigningConfiguration(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/SigningConfiguration/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningConfigurationExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/SigningConfiguration/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength(resourceName, 1),
					tfquerycheck.ExpectIdentityFunc(resourceName, identity.Checks()),
					querycheck.ExpectResourceDisplayName(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), tfknownvalue.AccountID()),
					tfquerycheck.ExpectNoResourceObject(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks())),
				},
			},
		},
	})
}

func testAccECRSigningConfiguration_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_signing_configuration.test"
	rName := testAccSigningProfileName(t)
	identity := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSigningConfiguration(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/SigningConfiguration/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningConfigurationExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
				},
			},

			// Step 2: Query with include_resource
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/SigningConfiguration/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength(resourceName, 1),
					tfquerycheck.ExpectIdentityFunc(resourceName, identity.Checks()),
					querycheck.ExpectResourceDisplayName(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), tfknownvalue.AccountID()),
					querycheck.ExpectResourceKnownValues(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("registry_id"), tfknownvalue.AccountID()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRule), knownvalue.SetSizeExact(1)),
					}),
				},
			},
		},
	})
}

func testAccECRSigningConfiguration_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_signing_configuration.test"
	rName := testAccSigningProfileName(t)
	altRegion := acctest.AlternateRegion()
	identity := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckSigningConfiguration(ctx, t, altRegion)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/SigningConfiguration/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(altRegion),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningConfigurationExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(altRegion)),
				},
			},

			// Step 2: Query with config.region set to the alternate Region
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/SigningConfiguration/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(altRegion),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength(resourceName, 1),
					tfquerycheck.ExpectIdentityFunc(resourceName, identity.Checks()),
					querycheck.ExpectResourceDisplayName(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), tfknownvalue.AccountID()),
				},
			},
		},
	})
}
