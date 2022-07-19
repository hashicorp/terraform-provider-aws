package signer_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/signer"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccSignerSigningJob_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_signer_signing_job.test"
	profileResourceName := "aws_signer_signing_profile.test"

	var job signer.DescribeSigningJobOutput
	var conf signer.GetSigningProfileOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		ErrorCheck:        acctest.ErrorCheck(t, signer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningJobConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningProfileExists(profileResourceName, &conf),
					testAccCheckSigningJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "platform_id", "AWSLambda-SHA384-ECDSA"),
					resource.TestCheckResourceAttr(resourceName, "platform_display_name", "AWS Lambda"),
					resource.TestCheckResourceAttr(resourceName, "status", "Succeeded"),
				),
			},
		},
	})

}

func testAccSigningJobConfig_basic(rName string) string {
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
  bucket        = "%[1]s"
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
`, rName)
}

func testAccCheckSigningJobExists(res string, job *signer.DescribeSigningJobOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Signing job not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Signing job with that ID does not exist")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SignerConn

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
