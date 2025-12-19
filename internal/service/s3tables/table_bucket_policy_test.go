// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package s3tables_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3tables "github.com/hashicorp/terraform-provider-aws/internal/service/s3tables"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3TablesTableBucketPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var tablebucketpolicy s3tables.GetTableBucketPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3tables_table_bucket_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableBucketPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketPolicyExists(ctx, resourceName, &tablebucketpolicy),
					resource.TestCheckResourceAttrSet(resourceName, "resource_policy"),
					resource.TestCheckResourceAttrPair(resourceName, "table_bucket_arn", "aws_s3tables_table_bucket.test", names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "table_bucket_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "table_bucket_arn",
				ImportStateVerifyIgnore:              []string{"resource_policy"},
			},
		},
	})
}

func TestAccS3TablesTableBucketPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var tablebucketpolicy s3tables.GetTableBucketPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3tables_table_bucket_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableBucketPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableBucketPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableBucketPolicyExists(ctx, resourceName, &tablebucketpolicy),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfs3tables.ResourceTableBucketPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTableBucketPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3TablesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3tables_table_bucket_policy" {
				continue
			}

			_, err := tfs3tables.FindTableBucketPolicyByARN(ctx, conn, rs.Primary.Attributes["table_bucket_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Tables Table Bucket Policy %s still exists", rs.Primary.Attributes["table_bucket_arn"])
		}

		return nil
	}
}

func testAccCheckTableBucketPolicyExists(ctx context.Context, n string, v *s3tables.GetTableBucketPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3TablesClient(ctx)

		output, err := tfs3tables.FindTableBucketPolicyByARN(ctx, conn, rs.Primary.Attributes["table_bucket_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTableBucketPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3tables_table_bucket_policy" "test" {
  resource_policy  = data.aws_iam_policy_document.test.json
  table_bucket_arn = aws_s3tables_table_bucket.test.arn
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["s3tables:*"]
    principals {
      type        = "AWS"
      identifiers = [data.aws_caller_identity.current.account_id]
    }
    resources = ["${aws_s3tables_table_bucket.test.arn}/*"]
  }
}

resource "aws_s3tables_table_bucket" "test" {
  name = %[1]q
}

data "aws_caller_identity" "current" {}
`, rName)
}
