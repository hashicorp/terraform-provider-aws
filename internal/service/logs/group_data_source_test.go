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

func TestAccLogsGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudwatch_log_group.test"
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "retention_in_days", resourceName, "retention_in_days"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccGroupDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource aws_cloudwatch_log_group "test" {
  name = %[1]q
}

data aws_cloudwatch_log_group "test" {
  name = aws_cloudwatch_log_group.test.name
}
`, rName)
}
