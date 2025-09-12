// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

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

func TestAccLogsLogGroup_List_Basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_cloudwatch_log_group.test[0]"
	resourceName2 := "aws_cloudwatch_log_group.test[1]"
	resourceName3 := "aws_cloudwatch_log_group.test[2]"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.LogsServiceID),
		CheckDestroy: testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/LogGroup/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				// Check: resource.ComposeAggregateTestCheckFunc(
				// 	testAccCheckLogGroupExists(ctx, t, resourceName, &v),
				// ),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "logs", "log-group:{name}"),
					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "logs", "log-group:{name}"),
					tfstatecheck.ExpectRegionalARNFormat(resourceName3, tfjsonpath.New(names.AttrARN), "logs", "log-group:{name}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/LogGroup/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigQueryChecks: []querycheck.QueryCheck{
					// TODO
				},
			},
		},
	})
}
