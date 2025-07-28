// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package signer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/signer"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsigner "github.com/hashicorp/terraform-provider-aws/internal/service/signer"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSignerSigningJob_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_signer_signing_job.test"
	var job signer.DescribeSigningJobOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningJobConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "platform_id", "AWSLambda-SHA384-ECDSA"),
					resource.TestCheckResourceAttr(resourceName, "platform_display_name", "AWS Lambda"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Succeeded"),
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

func testAccCheckSigningJobExists(ctx context.Context, n string, v *signer.DescribeSigningJobOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SignerClient(ctx)

		output, err := tfsigner.FindSigningJobByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
