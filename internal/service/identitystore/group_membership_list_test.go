// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package identitystore_test

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

func TestAccIdentityStoreGroupMembership_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_identitystore_group_membership.test[0]"
	resourceName2 := "aws_identitystore_group_membership.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstancesWithRegion(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		CheckDestroy:             testAccCheckGroupMembershipDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/GroupMembership/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/GroupMembership/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_identitystore_group_membership.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_identitystore_group_membership.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.NotNull()),
					tfquerycheck.ExpectNoResourceObject("aws_identitystore_group_membership.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_identitystore_group_membership.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_identitystore_group_membership.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.NotNull()),
					tfquerycheck.ExpectNoResourceObject("aws_identitystore_group_membership.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccIdentityStoreGroupMembership_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_identitystore_group_membership.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity := tfstatecheck.Identity()
	identityStoreID := tfstatecheck.StateValue()
	groupID := tfstatecheck.StateValue()
	memberID := tfstatecheck.StateValue()
	membershipID := tfstatecheck.StateValue()
	resourceID := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstancesWithRegion(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		CheckDestroy:             testAccCheckGroupMembershipDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/GroupMembership/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					identityStoreID.GetStateValue(resourceName, tfjsonpath.New("identity_store_id")),
					groupID.GetStateValue(resourceName, tfjsonpath.New("group_id")),
					memberID.GetStateValue(resourceName, tfjsonpath.New("member_id")),
					membershipID.GetStateValue(resourceName, tfjsonpath.New("membership_id")),
					resourceID.GetStateValue(resourceName, tfjsonpath.New(names.AttrID)),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/GroupMembership/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_identitystore_group_membership.test", identity.Checks()),
					querycheck.ExpectResourceDisplayName("aws_identitystore_group_membership.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), membershipID.ValueCheck()),
					querycheck.ExpectResourceKnownValues("aws_identitystore_group_membership.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), resourceID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("group_id"), groupID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("identity_store_id"), identityStoreID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("member_id"), memberID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("membership_id"), membershipID.ValueCheck()),
					}),
				},
			},
		},
	})
}

func TestAccIdentityStoreGroupMembership_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_identitystore_group_membership.test[0]"
	resourceName2 := "aws_identitystore_group_membership.test[1]"
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
			acctest.PreCheckSSOAdminInstancesWithRegion(ctx, t, acctest.AlternateRegion())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/GroupMembership/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/GroupMembership/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_identitystore_group_membership.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_identitystore_group_membership.test", identity2.Checks()),
				},
			},
		},
	})
}
