package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/signer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSignerSigningJob_basic(t *testing.T) {
	resourceName := "aws_signer_signing_job.test_job"
	profileResourceName := "aws_signer_signing_profile.test_sp"
	rString := acctest.RandString(48)
	profileName := fmt.Sprintf("tf_acc_sp_basic_%s", rString)

	var job signer.DescribeSigningJobOutput
	var conf signer.GetSigningProfileOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSignerSigningJobConfig(profileName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSignerSigningProfileExists(profileResourceName, profileName, &conf),
					testAccCheckAWSSignerSigningJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "platform_id", "AWSLambda-SHA384-ECDSA"),
					resource.TestCheckResourceAttr(resourceName, "platform_display_name", "AWS Lambda"),
					resource.TestCheckResourceAttr(resourceName, "status", "Succeeded"),
				),
			},
		},
	})

}

func testAccAWSSignerSigningJobConfig(profileName string) string {
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

resource "aws_signer_signing_job" "test_job" {
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
}`, profileName)
}

func testAccCheckAWSSignerSigningJobExists(res string, job *signer.DescribeSigningJobOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Signing job not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Signing job with that ID does not exist")
		}

		conn := testAccProvider.Meta().(*AWSClient).signerconn

		params := &signer.DescribeSigningJobInput{
			JobId: aws.String(rs.Primary.ID),
		}

		getJob, err := conn.DescribeSigningJob(params)
		if err != nil {
			return err
		}

		*job = *getJob

		return nil
	}
}
