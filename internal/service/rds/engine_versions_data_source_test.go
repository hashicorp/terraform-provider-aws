// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRDSEngineVersionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_engine_versions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "versions.0.engine", "mysql"),
					resource.TestCheckResourceAttrSet(dataSourceName, "versions.0.engine_version"),
					resource.TestCheckResourceAttrSet(dataSourceName, "versions.0.db_parameter_group_family"),
					resource.TestCheckResourceAttrSet(dataSourceName, "versions.0.status"),
				),
			},
		},
	})
}

func testAccEngineVersionsDataSourceConfig_basic() string {
	return `data "aws_rds_engine_versions" "test" {
  filter {
    name   = "engine"
    values = ["mysql"]
  }
}`
}
