// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions_test

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
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedPermissionsPolicyStoreAlias_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_verifiedpermissions_policy_store_alias.test[0]"
	resourceName2 := "aws_verifiedpermissions_policy_store_alias.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	aliasName1 := testAccPolicyStoreAliasName(rName + "-0")
	aliasName2 := testAccPolicyStoreAliasName(rName + "-1")

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(
				t,
				names.VerifiedPermissionsEndpointID,
			)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory(
					"testdata/PolicyStoreAlias/list_basic/",
				),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(
						resourceName1,
						tfjsonpath.New("alias_name"),
						knownvalue.StringExact(aliasName1),
					),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(
						resourceName2,
						tfjsonpath.New("alias_name"),
						knownvalue.StringExact(aliasName2),
					),
				},
			},
			{
				Query: true,
				ConfigDirectory: config.StaticDirectory(
					"testdata/PolicyStoreAlias/list_basic/",
				),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(
						"aws_verifiedpermissions_policy_store_alias.test",
						identity1.Checks(),
					),
					querycheck.ExpectResourceDisplayName(
						"aws_verifiedpermissions_policy_store_alias.test",
						tfqueryfilter.ByResourceIdentityFunc(
							identity1.Checks(),
						),
						knownvalue.StringExact(aliasName1),
					),
					tfquerycheck.ExpectNoResourceObject(
						"aws_verifiedpermissions_policy_store_alias.test",
						tfqueryfilter.ByResourceIdentityFunc(
							identity1.Checks(),
						),
					),

					tfquerycheck.ExpectIdentityFunc(
						"aws_verifiedpermissions_policy_store_alias.test",
						identity2.Checks(),
					),
					querycheck.ExpectResourceDisplayName(
						"aws_verifiedpermissions_policy_store_alias.test",
						tfqueryfilter.ByResourceIdentityFunc(
							identity2.Checks(),
						),
						knownvalue.StringExact(aliasName2),
					),
					tfquerycheck.ExpectNoResourceObject(
						"aws_verifiedpermissions_policy_store_alias.test",
						tfqueryfilter.ByResourceIdentityFunc(
							identity2.Checks(),
						),
					),
				},
			},
		},
	})
}

func TestAccVerifiedPermissionsPolicyStoreAlias_List_includeResource(
	t *testing.T,
) {
	ctx := acctest.Context(t)

	resourceName := "aws_verifiedpermissions_policy_store_alias.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	aliasName := testAccPolicyStoreAliasName(rName + "-0")

	identity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(
				t,
				names.VerifiedPermissionsEndpointID,
			)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory(
					"testdata/PolicyStoreAlias/list_include_resource/",
				),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("alias_name"),
						knownvalue.StringExact(aliasName),
					),
				},
			},
			{
				Query: true,
				ConfigDirectory: config.StaticDirectory(
					"testdata/PolicyStoreAlias/list_include_resource/",
				),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(
						"aws_verifiedpermissions_policy_store_alias.test",
						identity.Checks(),
					),
					querycheck.ExpectResourceDisplayName(
						"aws_verifiedpermissions_policy_store_alias.test",
						tfqueryfilter.ByResourceIdentityFunc(
							identity.Checks(),
						),
						knownvalue.StringExact(aliasName),
					),
					querycheck.ExpectResourceKnownValues(
						"aws_verifiedpermissions_policy_store_alias.test",
						tfqueryfilter.ByResourceIdentityFunc(
							identity.Checks(),
						),
						[]querycheck.KnownValueCheck{
							tfquerycheck.KnownValueCheck(
								tfjsonpath.New(names.AttrARN),
								knownvalue.NotNull(),
							),
							tfquerycheck.KnownValueCheck(
								tfjsonpath.New("alias_name"),
								knownvalue.StringExact(aliasName),
							),
							tfquerycheck.KnownValueCheck(
								tfjsonpath.New("created_at"),
								knownvalue.NotNull(),
							),
							tfquerycheck.KnownValueCheck(
								tfjsonpath.New("policy_store_id"),
								knownvalue.NotNull(),
							),
							tfquerycheck.KnownValueCheck(
								tfjsonpath.New(names.AttrRegion),
								knownvalue.StringExact(acctest.Region()),
							),
							tfquerycheck.KnownValueCheck(
								tfjsonpath.New(names.AttrState),
								knownvalue.StringExact("Active"),
							),
						},
					),
				},
			},
		},
	})
}

func TestAccVerifiedPermissionsPolicyStoreAlias_List_regionOverride(
	t *testing.T,
) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_verifiedpermissions_policy_store_alias.test[0]"
	resourceName2 := "aws_verifiedpermissions_policy_store_alias.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionHasService(
				t,
				names.VerifiedPermissionsEndpointID,
			)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory(
					"testdata/PolicyStoreAlias/list_region_override/",
				),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region": config.StringVariable(
						acctest.AlternateRegion(),
					),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},
			{
				Query: true,
				ConfigDirectory: config.StaticDirectory(
					"testdata/PolicyStoreAlias/list_region_override/",
				),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region": config.StringVariable(
						acctest.AlternateRegion(),
					),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(
						"aws_verifiedpermissions_policy_store_alias.test",
						identity1.Checks(),
					),
					tfquerycheck.ExpectIdentityFunc(
						"aws_verifiedpermissions_policy_store_alias.test",
						identity2.Checks(),
					),
				},
			},
		},
	})
}
