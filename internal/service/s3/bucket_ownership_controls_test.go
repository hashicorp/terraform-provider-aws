// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketOwnershipControls_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_ownership_controls.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketOwnershipControlsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketOwnershipControlsConfig_ruleObject(rName, string(types.ObjectOwnershipBucketOwnerPreferred)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketOwnershipControlsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.object_ownership", string(types.ObjectOwnershipBucketOwnerPreferred)),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketOwnershipControlsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketOwnershipControlsConfig_ruleObject(rName, string(types.ObjectOwnershipBucketOwnerPreferred)),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketOwnershipControlsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketOwnershipControlsConfig_ruleObject(rName, string(types.ObjectOwnershipBucketOwnerPreferred)),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketOwnershipControlsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketOwnershipControlsConfig_ruleObject(rName, string(types.ObjectOwnershipObjectWriter)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketOwnershipControlsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.object_ownership", string(types.ObjectOwnershipObjectWriter)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketOwnershipControlsConfig_ruleObject(rName, string(types.ObjectOwnershipBucketOwnerPreferred)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketOwnershipControlsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.object_ownership", string(types.ObjectOwnershipBucketOwnerPreferred)),
				),
			},
		},
	})
}

func TestAccS3BucketOwnershipControls_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketOwnershipControlsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketOwnershipControlsConfig_directoryBucket(rName, string(types.ObjectOwnershipBucketOwnerPreferred)),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketOwnershipControlsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_ownership_controls" {
				continue
			}

			_, err := tfs3.FindOwnershipControls(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Ownership Controls %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketOwnershipControlsExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		_, err := tfs3.FindOwnershipControls(ctx, conn, rs.Primary.ID)

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

func testAccBucketOwnershipControlsConfig_directoryBucket(rName, objectOwnership string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_directory_bucket.test.bucket

  rule {
    object_ownership = %[1]q
  }
}
`, objectOwnership))
}
