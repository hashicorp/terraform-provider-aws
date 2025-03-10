// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsDestinationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudwatch_log_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "destination_name", rName),
					acctest.MatchResourceAttrRegionalARN(ctx, dataSourceName, names.AttrARN, "logs", regexache.MustCompile(`destination:.+`)),
				),
			},
		},
	})
}

func testAccDestinationDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDestinationConfig_basic(rName), `
data "aws_cloudwatch_log_destination" "test" {
  destination_name = aws_cloudwatch_log_destination.test.name
  depends_on       = [aws_cloudwatch_log_destination.test]
}
`)
}
