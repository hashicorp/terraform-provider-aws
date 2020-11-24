package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSSignerSigningJob_basic(t *testing.T) {
	rString := acctest.RandString(48)
	profileName := fmt.Sprintf("tf_acc_sp_basic_%s", rString)
	dataSourceName := "data.aws_signer_signing_job.test"
	resourceName := "aws_signer_signing_job.job_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSSignerSigningJobConfigBasic(profileName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "job_owner", resourceName, "job_owner"),
					resource.TestCheckResourceAttrPair(dataSourceName, "job_invoker", resourceName, "job_invoker"),
					resource.TestCheckResourceAttrPair(dataSourceName, "profile_name", resourceName, "profile_name"),
				),
			},
		},
	})
}

func testAccDataSourceAWSSignerSigningJobConfigBasic(profileName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name        = "%s"
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-signer-signing-bucket"

  versioning {
    enabled = true
  }

  force_destroy = true
}

resource "aws_s3_bucket" "dest_bucket" {
  bucket        = "tf-signer-signing-dest-bucket"
  force_destroy = true
}

resource "aws_s3_bucket_object" "lambda_signing_code" {
  bucket = aws_s3_bucket.bucket.bucket
  key    = "lambdatest.zip"
  source = "test-fixtures/lambdatest.zip"
}

resource "aws_signer_signing_job" "job_test" {
  profile_name = aws_signer_signing_profile.test_sp.name

  source {
    s3 {
      bucket  = aws_s3_bucket.bucket.bucket
      key     = aws_s3_bucket_object.lambda_signing_code.key
      version = aws_s3_bucket_object.lambda_signing_code.version_id
    }
  }

  destination {
    s3 {
      bucket = aws_s3_bucket.dest_bucket.bucket
    }
  }
}

data "aws_signer_signing_job" "test" {
  job_id = aws_signer_signing_job.job_test.job_id
}`, profileName)
}
