// Copyright IBM Corp. 2014, 2026
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNamedQueryDataSourceConfig_basic(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDatabase, resourceName, names.AttrDatabase),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "query", resourceName, "querystring"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workgroup", resourceName, "workgroup"),
				),
			},
		},
	})
}

func testAccNamedQueryDataSourceConfig_basic(t *testing.T) string {
	return acctest.ConfigCompose(testAccNamedQueryConfig_basic(acctest.RandInt(t), sdkacctest.RandString(5)), `
data "aws_athena_named_query" "test" {
  name = aws_athena_named_query.test.name
}
`)
}
