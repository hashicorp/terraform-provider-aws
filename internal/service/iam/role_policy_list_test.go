// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMRolePolicy_List_Basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_iam_role_policy.test[0]"
	resourceName2 := "aws_iam_role_policy.test[1]"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.IAMServiceID),
		CheckDestroy: testAccCheckRolePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/RolePolicy/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-1")),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/RolePolicy/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_iam_role_policy.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrName:      knownvalue.StringExact(rName + "-0"),
						names.AttrRole:      knownvalue.StringExact(rName),
					}),
					querycheck.ExpectIdentity("aws_iam_role_policy.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrName:      knownvalue.StringExact(rName + "-1"),
						names.AttrRole:      knownvalue.StringExact(rName),
					}),
				},
			},
		},
	})
}
