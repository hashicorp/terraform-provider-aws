// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLogsGroupsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudwatch_log_groups.test"
	resource1Name := "aws_cloudwatch_log_group.test.0"
	resource2Name := "aws_cloudwatch_log_group.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "arns.*", resource1Name, "arn"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "arns.*", resource2Name, "arn"),
					resource.TestCheckResourceAttr(dataSourceName, "log_group_names.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "log_group_names.*", resource1Name, "name"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "log_group_names.*", resource2Name, "name"),
				),
			},
		},
	})
}

func TestAccLogsGroupsDataSource_noPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudwatch_log_groups.test"
	resource1Name := "aws_cloudwatch_log_group.test.0"
	resource2Name := "aws_cloudwatch_log_group.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupsDataSourceConfig_noPrefix(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "arns.#", 1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "arns.*", resource1Name, "arn"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "arns.*", resource2Name, "arn"),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "log_group_names.#", 1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "log_group_names.*", resource1Name, "name"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "log_group_names.*", resource2Name, "name"),
				),
			},
		},
	})
}

func testAccGroupsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource aws_cloudwatch_log_group "test" {
  count = 2

  name = "%[1]s-${count.index}"
}

data aws_cloudwatch_log_groups "test" {
  log_group_name_prefix = %[1]q

  depends_on = [aws_cloudwatch_log_group.test[0], aws_cloudwatch_log_group.test[1]]
}
`, rName)
}

func testAccGroupsDataSourceConfig_noPrefix(rName string) string {
	return fmt.Sprintf(`
resource aws_cloudwatch_log_group "test" {
  count = 2

  name = "%[1]s-${count.index}"
}

data aws_cloudwatch_log_groups "test" {
  depends_on = [aws_cloudwatch_log_group.test[0], aws_cloudwatch_log_group.test[1]]
}
`, rName)
}
