package signer_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/signer"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSignerSigningJobDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_signer_signing_job.test"
	resourceName := "aws_signer_signing_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		ErrorCheck:        acctest.ErrorCheck(t, signer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningJobBasicDataSourceConfig(rName),
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

func testAccSigningJobBasicDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile" "test" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_s3_bucket" "source" {
  bucket        = "%[1]s-source"
  force_destroy = true
}

resource "aws_s3_bucket_versioning" "source" {
  bucket = aws_s3_bucket.source.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket" "destination" {
  bucket        = "%[1]s-destination"
  force_destroy = true
}

resource "aws_s3_object" "source" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.source]

  bucket = aws_s3_bucket.source.bucket
  key    = "lambdatest.zip"
  source = "test-fixtures/lambdatest.zip"
}

resource "aws_signer_signing_job" "test" {
  profile_name = aws_signer_signing_profile.test.name

  source {
    s3 {
      bucket  = aws_s3_object.source.bucket
      key     = aws_s3_object.source.key
      version = aws_s3_object.source.version_id
    }
  }

  destination {
    s3 {
      bucket = aws_s3_bucket.destination.bucket
    }
  }
}

data "aws_signer_signing_job" "test" {
  job_id = aws_signer_signing_job.test.job_id
}
`, rName)
}
