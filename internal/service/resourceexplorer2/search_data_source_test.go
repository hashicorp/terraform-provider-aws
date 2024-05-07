// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourceexplorer2_test

import (
	"fmt"
	"os"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccSearchDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_resourceexplorer2_search.test"
	viewResourceName := "aws_resourceexplorer2_view.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ResourceExplorer2EndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSearchDataSourceConfig_basic(rName, "LOCAL"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "view_arn", viewResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, "resource_count.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "resources.#"),
				),
			},
		},
	})
}

// Can only be run once a day as changing the index type has a 24 hr cooldown
func testAccSearchDataSource_IndexType(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("RESOURCEEXPLORER_INDEX_TYPE") != "AGGREGATOR" {
		t.Skip("Environment variable RESOURCEEXPLORER_INDEX_TYPE is not set to AGGREGATOR. Before setting this environment variable and running this test, ensure no tests with a LOCAL index type also need to run. Changing the index type will trigger a cool down period of 24 hours.")
	}
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_resourceexplorer2_search.test"
	viewResourceName := "aws_resourceexplorer2_view.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ResourceExplorer2EndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSearchDataSourceConfig_basic(rName, "AGGREGATOR"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "view_arn", viewResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, "resource_count.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "resources.#"),
				),
			},
		},
	})
}

func testAccSearchDataSourceConfig_basic(rName, indexType string) string {
	return fmt.Sprintf(`
resource "aws_resourceexplorer2_index" "test" {
  type = %[2]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_resourceexplorer2_view" "test" {
  depends_on = [aws_resourceexplorer2_index.test]

  name         = %[1]q
  default_view = true
}

data "aws_resourceexplorer2_search" "test" {
  depends_on = [aws_resourceexplorer2_view.test]

  query_string = "region:global"
  view_arn     = aws_resourceexplorer2_view.test.arn
}
`, rName, indexType)
}
