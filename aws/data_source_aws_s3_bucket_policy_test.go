package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAWSS3BucketPolicy_basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3BucketPolicyConfigBasic1(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
				),
			},
			{
				Config: testAccAWSDataSourceS3BucketPolicyConfigBasic2(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttrPair("aws_s3_bucket_policy.bucket", "policy", "data.aws_s3_bucket_policy.policy", "policy"),
				),
			},
		},
	})
}

func testAccAWSDataSourceS3BucketPolicyConfigBasic1(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_s3_bucket" "bucket" {
		bucket = "tf-test-bucket-%d"
		tags = {
			TestName = "TestAccAWSS3BucketPolicy_basic"
		}
	}
	
	resource "aws_s3_bucket_policy" "bucket" {
		bucket = "${aws_s3_bucket.bucket.bucket}"
		policy = "${data.aws_iam_policy_document.policy.json}"
	}
	
	data "aws_iam_policy_document" "policy" {
	  statement {
		effect = "Allow"
	
		actions = [
		  "s3:*",
		]
	
		resources = [
		  "${aws_s3_bucket.bucket.arn}",
		  "${aws_s3_bucket.bucket.arn}/*",
		]
	
		principals {
		  type        = "AWS"
		  identifiers = ["*"]
		}
	  }
	}
`, rInt)
}

func testAccAWSDataSourceS3BucketPolicyConfigBasic2(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_s3_bucket" "bucket" {
		bucket = "tf-test-bucket-%d"
		tags = {
			TestName = "TestAccAWSS3BucketPolicy_basic"
		}
	}
	
	resource "aws_s3_bucket_policy" "bucket" {
		bucket = "${aws_s3_bucket.bucket.bucket}"
		policy = "${data.aws_iam_policy_document.policy.json}"
	}
	
	data "aws_iam_policy_document" "policy" {
	  statement {
		effect = "Allow"
	
		actions = [
		  "s3:*",
		]
	
		resources = [
		  "${aws_s3_bucket.bucket.arn}",
		  "${aws_s3_bucket.bucket.arn}/*",
		]
	
		principals {
		  type        = "AWS"
		  identifiers = ["*"]
		}
	  }
	}

	data "aws_s3_bucket_policy" "policy" {
		bucket = "${aws_s3_bucket.bucket.bucket}"
	}
`, rInt)
}
