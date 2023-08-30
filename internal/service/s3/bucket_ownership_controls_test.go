// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestAccS3BucketOwnershipControls_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_ownership_controls.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketOwnershipControlsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketOwnershipControlsConfig_ruleObject(rName, s3.ObjectOwnershipBucketOwnerPreferred),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketOwnershipControlsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.object_ownership", s3.ObjectOwnershipBucketOwnerPreferred),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketOwnershipControls_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_ownership_controls.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketOwnershipControlsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketOwnershipControlsConfig_ruleObject(rName, s3.ObjectOwnershipBucketOwnerPreferred),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketOwnershipControlsExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucketOwnershipControls(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketOwnershipControls_Disappears_bucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_ownership_controls.test"
	s3BucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketOwnershipControlsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketOwnershipControlsConfig_ruleObject(rName, s3.ObjectOwnershipBucketOwnerPreferred),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketOwnershipControlsExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucket(), s3BucketResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketOwnershipControls_Rule_objectOwnership(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_ownership_controls.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketOwnershipControlsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketOwnershipControlsConfig_ruleObject(rName, s3.ObjectOwnershipObjectWriter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketOwnershipControlsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.object_ownership", s3.ObjectOwnershipObjectWriter),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketOwnershipControlsConfig_ruleObject(rName, s3.ObjectOwnershipBucketOwnerPreferred),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketOwnershipControlsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.object_ownership", s3.ObjectOwnershipBucketOwnerPreferred),
				),
			},
		},
	})
}

func testAccCheckBucketOwnershipControlsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_ownership_controls" {
				continue
			}

			input := &s3.GetBucketOwnershipControlsInput{
				Bucket: aws.String(rs.Primary.ID),
			}

			_, err := conn.GetBucketOwnershipControlsWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
				continue
			}

			if tfawserr.ErrCodeEquals(err, "OwnershipControlsNotFoundError") {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Ownership Controls (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketOwnershipControlsExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no resource ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn(ctx)

		input := &s3.GetBucketOwnershipControlsInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetBucketOwnershipControlsWithContext(ctx, input)

		return err
	}
}

func testAccBucketOwnershipControlsConfig_ruleObject(rName, objectOwnership string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    object_ownership = %[2]q
  }
}
`, rName, objectOwnership)
}
