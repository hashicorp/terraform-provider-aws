// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMRole_List_Basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_iam_role.test[0]"
	resourceName2 := "aws_iam_role.test[1]"
	resourceName3 := "aws_iam_role.test[2]"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
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
				ConfigDirectory:          config.StaticDirectory("testdata/Role/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectGlobalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "iam", "role/{name}"),
					tfstatecheck.ExpectGlobalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "iam", "role/{name}"),
					tfstatecheck.ExpectGlobalARNFormat(resourceName3, tfjsonpath.New(names.AttrARN), "iam", "role/{name}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Role/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigQueryResultChecks: []querycheck.QueryResultCheck{
					// TODO
				},
			},
		},
	})
}
