// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAthenaNamedQueryDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_athena_named_query.test"
	dataSourceName := "data.aws_athena_named_query.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNamedQueryDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "database", resourceName, "database"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "query", resourceName, "querystring"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workgroup", resourceName, "workgroup"),
				),
			},
		},
	})
}

func testAccNamedQueryDataSourceConfig_basic() string {
	return acctest.ConfigCompose(testAccNamedQueryConfig_basic(sdkacctest.RandInt(), sdkacctest.RandString(5)), `
data "aws_athena_named_query" "test" {
  name = aws_athena_named_query.test.name
}
`)
}
