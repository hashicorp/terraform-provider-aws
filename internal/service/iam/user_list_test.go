// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

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

func TestAccIAMUser_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_iam_user.test[0]"
	resourceName2 := "aws_iam_user.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/User/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-1")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/User/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_iam_user.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_iam_user.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_iam_user.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_iam_user.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_iam_user.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"-1")),
					tfquerycheck.ExpectNoResourceObject("aws_iam_user.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccIAMUser_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_iam_user.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/User/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/User/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_iam_user.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_iam_user.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					querycheck.ExpectResourceKnownValues("aws_iam_user.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNExact("iam", "user/"+rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrForceDestroy), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPath), knownvalue.StringExact("/")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("permissions_boundary"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("unique_id"), knownvalue.NotNull()),
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

func TestAccIAMUser_List_pathPrefix(t *testing.T) {
	ctx := acctest.Context(t)

	resourceNameExpected1 := "aws_iam_user.expected[0]"
	resourceNameExpected2 := "aws_iam_user.expected[1]"
	resourceNameNotExpected1 := "aws_iam_user.not_expected[0]"
	resourceNameNotExpected2 := "aws_iam_user.not_expected[1]"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rPathName := "/" + acctest.RandomWithPrefix(t, acctest.ResourcePrefix) + "/"
	rOtherPathName := "/" + acctest.RandomWithPrefix(t, acctest.ResourcePrefix) + "/"

	identityExpected1 := tfstatecheck.Identity()
	identityExpected2 := tfstatecheck.Identity()
	identityNotExpected1 := tfstatecheck.Identity()
	identityNotExpected2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/User/list_path_prefix/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:      config.StringVariable(rName),
					"resource_count":     config.IntegerVariable(2),
					"expected_path_name": config.StringVariable(rPathName),
					"other_path_name":    config.StringVariable(rOtherPathName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identityExpected1.GetIdentity(resourceNameExpected1),
					statecheck.ExpectKnownValue(resourceNameExpected1, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),

					identityExpected2.GetIdentity(resourceNameExpected2),
					statecheck.ExpectKnownValue(resourceNameExpected2, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-1")),

					identityNotExpected1.GetIdentity(resourceNameNotExpected1),
					statecheck.ExpectKnownValue(resourceNameNotExpected1, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-other-0")),

					identityNotExpected2.GetIdentity(resourceNameNotExpected2),
					statecheck.ExpectKnownValue(resourceNameNotExpected2, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-other-1")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/User/list_path_prefix/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:      config.StringVariable(rName),
					"resource_count":     config.IntegerVariable(2),
					"expected_path_name": config.StringVariable(rPathName),
					"other_path_name":    config.StringVariable(rOtherPathName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("aws_iam_user.test", 2),
					tfquerycheck.ExpectIdentityFunc("aws_iam_user.test", identityExpected1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_iam_user.test", identityExpected2.Checks()),
				},
			},
		},
	})
}
