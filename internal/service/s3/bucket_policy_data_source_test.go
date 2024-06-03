// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"fmt"
	"testing"

	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketPolicyDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_bucket_policy.test"
	resourceName := "aws_s3_bucket_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPolicyDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketPolicyMatch(dataSourceName, names.AttrPolicy, resourceName, names.AttrPolicy),
				),
			},
		},
	})
}

func testAccCheckBucketPolicyMatch(nameFirst, keyFirst, nameSecond, keySecond string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[nameFirst]
		if !ok {
			return fmt.Errorf("Not found: %s", nameFirst)
		}

		policy1, ok := rs.Primary.Attributes[keyFirst]
		if !ok {
			return fmt.Errorf("attribute %q not found for %q", keyFirst, nameFirst)
		}

		rs, ok = s.RootModule().Resources[nameSecond]
		if !ok {
			return fmt.Errorf("Not found: %s", nameSecond)
		}

		policy2, ok := rs.Primary.Attributes[keySecond]
		if !ok {
			return fmt.Errorf("attribute %q not found for %q", keySecond, nameSecond)
		}

		areEquivalent, err := awspolicy.PoliciesAreEquivalent(policy1, policy2)
		if err != nil {
			return fmt.Errorf("comparing IAM Policies failed: %s", err)
		}

		if !areEquivalent {
			return fmt.Errorf("S3 bucket policies differ.\npolicy1: %s\npolicy2: %s", policy1, policy2)
		}

		return nil
	}
}

func testAccDataSourceBucketPolicyConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    actions = [
      "s3:GetObject",
      "s3:ListBucket",
    ]

    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*",
    ]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}
`, rName)
}

func testAccBucketPolicyDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDataSourceBucketPolicyConfig_base(rName), `
data "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id

  depends_on = [aws_s3_bucket_policy.test]
}
`)
}
