// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2Instance_List_Basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_instance.test[0]"
	resourceName2 := "aws_instance.test[1]"
	resourceName3 := "aws_instance.test[2]"

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy: testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Instance/list_basic/"),
				// Check: resource.ComposeAggregateTestCheckFunc(
				// 	testAccCheckLogGroupExists(ctx, t, resourceName, &v),
				// ),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "ec2", "instance/{id}"),
					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "ec2", "instance/{id}"),
					tfstatecheck.ExpectRegionalARNFormat(resourceName3, tfjsonpath.New(names.AttrARN), "ec2", "instance/{id}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Instance/list_basic/"),
				ConfigQueryChecks:        []querycheck.QueryCheck{
					// TODO
				},
			},
		},
	})
}
