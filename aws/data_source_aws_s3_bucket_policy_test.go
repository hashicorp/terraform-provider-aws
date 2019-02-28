package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceS3BucketPolicy(t *testing.T) {
	name := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	partition := testAccGetPartition()

	policy := fmt.Sprintf(`{
	"Version": "2012-10-17",
	"Statement": [{
		"Sid": "",
		"Effect": "Allow",
		"Principal": {"AWS":"*"},
		"Action": "s3:*",
		"Resource": ["arn:%s:s3:::%s/*","arn:%s:s3:::%s"]
	}]
}`, partition, name, partition, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3BucketPolicyConfig(name, policy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("data.aws_s3_bucket_policy.bucket"),
					testAccCheckAWSS3BucketHasPolicy("data.aws_s3_bucket_policy.bucket", policy),
				),
			},
		},
	})
}

func testAccAWSDataSourceS3BucketPolicyConfig(bucketName string, bucketPolicy string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "%s"
	tags = {
		TestName = "TestAccAWSDataSourceS3BucketPolicy"
	}
}

resource "aws_s3_bucket_policy" "bucket" {
	bucket = "${aws_s3_bucket.bucket.id}"
	policy = <<POLICY
%s
POLICY
}

data "aws_s3_bucket_policy" "bucket" {
	bucket = "${aws_s3_bucket_policy.bucket.bucket}"
}
`, bucketName, bucketPolicy)
}
