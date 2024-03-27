// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourceexplorer2_test

import (
	"fmt"
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
					resource.TestCheckResourceAttrPair(dataSourceName, "view_arn", viewResourceName, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "resource_count.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "resources.#"),
				),
			},
		},
	})
}

func testAccSearchDataSource_IndexType(t *testing.T) {
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
				Config: testAccSearchDataSourceConfig_basic(rName, "AGGREGATOR"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "view_arn", viewResourceName, "arn"),
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
  name         = %[1]q
  default_view = true

  depends_on = [aws_resourceexplorer2_index.test]
}

data "aws_resourceexplorer2_search" "test" {
  query_string = "region:global"
  view_arn     = aws_resourceexplorer2_view.test.arn

  depends_on = [aws_resourceexplorer2_view.test]
}
`, rName, indexType)
}
