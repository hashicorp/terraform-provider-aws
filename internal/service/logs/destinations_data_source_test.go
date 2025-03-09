// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsDestinationsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudwatch_log_destinations.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "destinations.#", 1),
				),
			},
		},
	})
}

func TestAccLogsDestinationsDataSource_destinationNamePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudwatch_log_destinations.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestinationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationsDataSourceConfig_destinationNamePrefix(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "destinations.#", 1),
				),
			},
		},
	})
}

func testAccDestinationsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDestinationConfig_basic(rName), `
data "aws_cloudwatch_log_destinations" "test" {
  depends_on = [aws_cloudwatch_log_destination.test]
}
`)
}

func testAccDestinationsDataSourceConfig_destinationNamePrefix(rName string) string {
	return acctest.ConfigCompose(testAccDestinationConfig_basic(rName), fmt.Sprintf(`
data "aws_cloudwatch_log_destinations" "test" {
  destination_name_prefix = "%s"
  depends_on = [aws_cloudwatch_log_destination.test]
}
`, rName))
}
