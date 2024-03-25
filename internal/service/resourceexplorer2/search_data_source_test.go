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

func TestAccResourceExplorer2SearchDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_resourceexplorer2_search.test"
	viewResourceName := "aws_resourceexplorer2_view.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ResourceExplorer2EndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceExplorer2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSearchDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					// testAccCheckSearchExists(ctx, dataSourceName, &search),
					resource.TestCheckResourceAttrPair(dataSourceName, "view_arn", viewResourceName, "arn"),
					// resource.TestCheckResourceAttrSet(dataSourceName, "count.#"),
				),
			},
		},
	})
}

func testAccSearchDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_resourceexplorer2_index" "test" {
  type = "AGGREGATOR"

  tags = {
    Name = %[1]q
  }
}

resource "aws_resourceexplorer2_view" "test" {
  name         = %[1]q

  depends_on = [aws_resourceexplorer2_index.test]
}

data "aws_resourceexplorer2_search" "test" {
  query_string = "region:us-west-2"
  view_arn = aws_resourceexplorer2_index.test.arn
}
`, rName)
}
