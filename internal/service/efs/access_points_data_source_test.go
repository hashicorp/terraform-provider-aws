// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEFSAccessPointsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_efs_access_points.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccEFSAccessPointsDataSource_empty(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_efs_access_points.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointsDataSourceConfig_empty(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccAccessPointsDataSourceConfig_basic() string {
	return `
resource "aws_efs_file_system" "test" {}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
}

data "aws_efs_access_points" "test" {
  file_system_id = aws_efs_access_point.test.file_system_id
}
`
}

func testAccAccessPointsDataSourceConfig_empty() string {
	return `
resource "aws_efs_file_system" "test" {}

data "aws_efs_access_points" "test" {
  file_system_id = aws_efs_file_system.test.id
}
`
}
