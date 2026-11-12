// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityir_test

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

func testAccMembership_List_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityir_membership.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityIRServiceID),
		CheckDestroy:             testAccCheckMembershipDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Membership/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkMembershipARN),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Membership/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(resourceName, identity.Checks()),
					querycheck.ExpectResourceDisplayName(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.StringExact(rName)),
					tfquerycheck.ExpectNoResourceObject(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks())),
				},
			},
		},
	})
}

func testAccMembership_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityir_membership.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityIRServiceID),
		CheckDestroy:             testAccCheckMembershipDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Membership/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkMembershipARN),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Membership/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(resourceName, identity.Checks()),
					querycheck.ExpectResourceDisplayName(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.StringExact(rName)),
					querycheck.ExpectResourceKnownValues(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAccountID), tfknownvalue.AccountID()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), checkMembershipARN),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("cover_entire_organization"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("customer_type"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("incident_response_team"), knownvalue.ListSizeExact(2)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("membership_activation_timestamp"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("membership_name"), knownvalue.StringExact(rName)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("membership_status"), knownvalue.StringExact("Active")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("number_of_accounts_covered"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("organizational_units"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
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

func testAccMembership_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityir_membership.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityIRServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Membership/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Membership/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(resourceName, identity.Checks()),
				},
			},
		},
	})
}
