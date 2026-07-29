// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccSSOAdminRegion_listBasic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ssoadmin_region.test[0]"
	resourceName2 := "aws_ssoadmin_region.test[1]"

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckMultipleRegion(t, 3)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		CheckDestroy:             testAccCheckRegionDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Region/list_basic/"),
				ConfigVariables: config.Variables{
					"region_names": config.ListVariable(
						config.StringVariable(acctest.AlternateRegion()),
						config.StringVariable(acctest.ThirdRegion()),
					),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("region_name"), knownvalue.StringExact(acctest.AlternateRegion())),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("region_name"), knownvalue.StringExact(acctest.ThirdRegion())),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Region/list_basic/"),
				ConfigVariables: config.Variables{
					"region_names": config.ListVariable(
						config.StringVariable(acctest.AlternateRegion()),
						config.StringVariable(acctest.ThirdRegion()),
					),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_region.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_ssoadmin_region.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(acctest.AlternateRegion())),
					tfquerycheck.ExpectNoResourceObject("aws_ssoadmin_region.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_region.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_ssoadmin_region.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(acctest.ThirdRegion())),
					tfquerycheck.ExpectNoResourceObject("aws_ssoadmin_region.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func testAccSSOAdminRegion_listIncludeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ssoadmin_region.test[0]"

	identity1 := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		CheckDestroy:             testAccCheckRegionDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Region/list_include_resource/"),
				ConfigVariables: config.Variables{
					"region_names": config.ListVariable(
						config.StringVariable(acctest.AlternateRegion()),
					),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("region_name"), knownvalue.StringExact(acctest.AlternateRegion())),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Region/list_include_resource/"),
				ConfigVariables: config.Variables{
					"region_names": config.ListVariable(
						config.StringVariable(acctest.AlternateRegion()),
					),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_region.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_ssoadmin_region.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(acctest.AlternateRegion())),
					querycheck.ExpectResourceKnownValues("aws_ssoadmin_region.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("instance_arn"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("region_name"), knownvalue.StringExact(acctest.AlternateRegion())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrStatus), knownvalue.StringExact(string(types.RegionStatusActive))),
					}),
				},
			},
		},
	})
}

func testAccSSOAdminRegion_listRegionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ssoadmin_region.test[0]"

	identity1 := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckSSOAdminInstancesWithRegion(ctx, t, acctest.AlternateRegion())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		CheckDestroy:             testAccCheckRegionDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Region/list_region_override/"),
				ConfigVariables: config.Variables{
					"region_names": config.ListVariable(
						config.StringVariable(acctest.Region()),
					),
					"resource_count": config.IntegerVariable(1),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Region/list_region_override/"),
				ConfigVariables: config.Variables{
					"region_names": config.ListVariable(
						config.StringVariable(acctest.Region()),
					),
					"resource_count": config.IntegerVariable(1),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_region.test", identity1.Checks()),
				},
			},
		},
	})
}
