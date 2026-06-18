// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketNotificationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_bucket_notification.test"
	lambdaResourceName := "aws_lambda_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketNotificationDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "eventbridge", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "lambda_function.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "lambda_function.0.id", "notification-lambda"),
					resource.TestCheckResourceAttrPair(dataSourceName, "lambda_function.0.lambda_function_arn", lambdaResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "lambda_function.0.events.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "lambda_function.0.filter_prefix", "tf-acc-test/"),
					resource.TestCheckResourceAttr(dataSourceName, "lambda_function.0.filter_suffix", ".png"),
					resource.TestCheckResourceAttr(dataSourceName, "queue.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "topic.#", "0"),
				),
			},
		},
	})
}

func testAccBucketNotificationDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccBucketNotificationConfig_lambdaFunction(rName),
		`
data "aws_s3_bucket_notification" "test" {
  bucket = aws_s3_bucket_notification.test.bucket
}
`,
	)
}
