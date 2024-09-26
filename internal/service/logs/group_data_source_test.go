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

func TestAccLogsGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudwatch_log_group.test"
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyID, resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttrPair(dataSourceName, "log_group_class", resourceName, "log_group_class"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "retention_in_days", resourceName, "retention_in_days"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccLogsGroupDataSource_retentionPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudwatch_log_group.test"
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_retentionPolicy(rName, 365),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "retention_in_days", resourceName, "retention_in_days"),
				),
			},
		},
	})
}

func testAccGroupDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data aws_cloudwatch_log_group "test" {
  name = aws_cloudwatch_log_group.test.name
}

resource aws_cloudwatch_log_group "test" {
  name = %[1]q
}
`, rName)
}

func testAccGroupDataSourceConfig_retentionPolicy(rName string, val int) string {
	return fmt.Sprintf(`
data aws_cloudwatch_log_group "test" {
  name = aws_cloudwatch_log_group.test.name
}

resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  retention_in_days = %[2]d
}
`, rName, val)
}
