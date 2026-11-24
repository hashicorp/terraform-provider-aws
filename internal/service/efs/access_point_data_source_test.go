// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package efs_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEFSAccessPointDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_efs_access_point.test"
	resourceName := "aws_efs_access_point.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessPoint/data.basic/"),
				ConfigVariables: config.Variables{},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrARN), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrOwnerID), resourceName, tfjsonpath.New(names.AttrOwnerID), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("posix_user"), resourceName, tfjsonpath.New("posix_user"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("root_directory"), resourceName, tfjsonpath.New("root_directory"), compare.ValuesSame()),
				},
			},
		},
	})
}
