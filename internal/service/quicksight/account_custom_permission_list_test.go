// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

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

func testAccAccountCustomPermission_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_quicksight_account_custom_permission.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		CheckDestroy:             testAccCheckAccountCustomPermissionDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccountCustomPermission/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAWSAccountID), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_permissions_name"), knownvalue.StringExact(rName)),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/AccountCustomPermission/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_quicksight_account_custom_permission.test", identity.Checks()),
					querycheck.ExpectResourceDisplayName("aws_quicksight_account_custom_permission.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.StringExact(rName)),
					tfquerycheck.ExpectNoResourceObject("aws_quicksight_account_custom_permission.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks())),
				},
			},
		},
	})
}

func testAccAccountCustomPermission_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_quicksight_account_custom_permission.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		CheckDestroy:             testAccCheckAccountCustomPermissionDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccountCustomPermission/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAWSAccountID), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_permissions_name"), knownvalue.StringExact(rName)),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/AccountCustomPermission/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_quicksight_account_custom_permission.test", identity.Checks()),
					querycheck.ExpectResourceDisplayName("aws_quicksight_account_custom_permission.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.StringExact(rName)),
					querycheck.ExpectResourceKnownValues("aws_quicksight_account_custom_permission.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAWSAccountID), tfknownvalue.AccountID()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("custom_permissions_name"), knownvalue.StringExact(rName)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					}),
				},
			},
		},
	})
}

func testAccAccountCustomPermission_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_quicksight_account_custom_permission.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		CheckDestroy:             testAccCheckAccountCustomPermissionDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccountCustomPermission/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_permissions_name"), knownvalue.StringExact(rName)),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/AccountCustomPermission/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_quicksight_account_custom_permission.test", identity.Checks()),
				},
			},
		},
	})
}
