// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMPolicy_List_Basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_iam_policy.test[0]"
	resourceName2 := "aws_iam_policy.test[1]"
	resourceName3 := "aws_iam_policy.test[2]"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	arn1 := tfstatecheck.StateValue()
	arn2 := tfstatecheck.StateValue()
	arn3 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.IAMServiceID),
		CheckDestroy: testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Policy/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					arn1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrARN)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNExact("iam", "policy/"+rName+"-0")),

					arn2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrARN)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNExact("iam", "policy/"+rName+"-1")),

					arn3.GetStateValue(resourceName3, tfjsonpath.New(names.AttrARN)),
					statecheck.ExpectKnownValue(resourceName3, tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNExact("iam", "policy/"+rName+"-2")),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Policy/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_iam_policy.test", map[string]knownvalue.Check{
						names.AttrARN: arn1.Value(),
					}),

					querycheck.ExpectIdentity("aws_iam_policy.test", map[string]knownvalue.Check{
						names.AttrARN: arn2.Value(),
					}),

					querycheck.ExpectIdentity("aws_iam_policy.test", map[string]knownvalue.Check{
						names.AttrARN: arn3.Value(),
					}),
				},
			},
		},
	})
}

func TestAccIAMPolicy_List_PathPrefix(t *testing.T) {
	ctx := acctest.Context(t)

	resourceNameExpected1 := "aws_iam_policy.expected[0]"
	resourceNameExpected2 := "aws_iam_policy.expected[1]"
	resourceNameNotExpected1 := "aws_iam_policy.not_expected[0]"
	resourceNameNotExpected2 := "aws_iam_policy.not_expected[1]"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rPathName := "/" + sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) + "/"
	rOtherPathName := "/" + sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) + "/"

	expected1 := tfstatecheck.StateValue()
	expected2 := tfstatecheck.StateValue()
	notExpected1 := tfstatecheck.StateValue()
	notExpected2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.IAMServiceID),
		CheckDestroy: testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Policy/list_path_prefix/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:      config.StringVariable(rName),
					"expected_path_name": config.StringVariable(rPathName),
					"other_path_name":    config.StringVariable(rOtherPathName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					expected1.GetStateValue(resourceNameExpected1, tfjsonpath.New(names.AttrARN)),
					expected2.GetStateValue(resourceNameExpected2, tfjsonpath.New(names.AttrARN)),
					notExpected1.GetStateValue(resourceNameNotExpected1, tfjsonpath.New(names.AttrARN)),
					notExpected2.GetStateValue(resourceNameNotExpected2, tfjsonpath.New(names.AttrARN)),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Policy/list_path_prefix/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:      config.StringVariable(rName),
					"expected_path_name": config.StringVariable(rPathName),
					"other_path_name":    config.StringVariable(rOtherPathName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_iam_policy.expected", map[string]knownvalue.Check{
						names.AttrARN: expected1.Value(),
					}),
					querycheck.ExpectIdentity("aws_iam_policy.expected", map[string]knownvalue.Check{
						names.AttrARN: expected2.Value(),
					}),
					querycheck.ExpectNoIdentity("aws_iam_policy.expected", map[string]knownvalue.Check{
						names.AttrARN: notExpected1.Value(),
					}),
					querycheck.ExpectNoIdentity("aws_iam_policy.expected", map[string]knownvalue.Check{
						names.AttrARN: notExpected2.Value(),
					}),
				},
			},
		},
	})
}
