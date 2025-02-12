// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this resource's maintainability by following this
// outline.
//
// 1. Package declaration (add "_test" since this is a test file)
// 2. Imports
// 3. Unit tests
// 4. Basic test
// 5. Disappears test
// 6. All the other tests
// 7. Helper functions (exists, destroy, check, etc.)
// 8. Functions that return Terraform configurations

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance test access AWS and cost money to run.
func TestAccS3BucketMetadataTableConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var bucketMetadataTableConfigurationResult awstypes.GetBucketMetadataTableConfigurationResult
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_metadata_table_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetadataTableConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetadataTableConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketMetadataTableConfigurationExists(ctx, resourceName, &bucketMetadataTableConfigurationResult),
					resource.TestCheckResourceAttrSet(resourceName, "metadata_table_configuration.0.s3_tables_destination.0.table_name"),
					resource.TestCheckResourceAttr(resourceName, "metadata_table_configuration.0.s3_tables_destination.0.table_name", "s3metadata_test_uniq"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrBucket),
				ImportStateVerifyIdentifierAttribute: names.AttrBucket,
				ImportStateVerifyIgnore:              []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccS3BucketMetadataTableConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var bucketMetadataTableConfigurationResult awstypes.GetBucketMetadataTableConfigurationResult
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_metadata_table_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.S3ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetadataTableConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetadataTableConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketMetadataTableConfigurationExists(ctx, resourceName, &bucketMetadataTableConfigurationResult),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucketMetadataTableConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBucketMetadataTableConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_metadata_table_configuration" {
				continue
			}

			bucket := rs.Primary.Attributes[names.AttrBucket]
			if bucket == "" {
				return create.Error(names.S3, create.ErrActionCheckingExistence, tfs3.ResNameBucketMetadataTableConfiguration, rs.Primary.ID, errors.New("no bucket is set"))
			}

			_, err := tfs3.FindBucketMetadataTableConfigurationByBucket(ctx, conn, bucket)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.S3, create.ErrActionCheckingDestroyed, tfs3.ResNameBucketMetadataTableConfiguration, bucket, err)
			}

			return create.Error(names.S3, create.ErrActionCheckingDestroyed, tfs3.ResNameBucketMetadataTableConfiguration, bucket, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckBucketMetadataTableConfigurationExists(ctx context.Context, name string, bucketmetadatatableconfiguration *awstypes.GetBucketMetadataTableConfigurationResult) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.S3, create.ErrActionCheckingExistence, tfs3.ResNameBucketMetadataTableConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.S3, create.ErrActionCheckingExistence, tfs3.ResNameBucketMetadataTableConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		bucket := rs.Primary.Attributes[names.AttrBucket]
		if bucket == "" {
			return create.Error(names.S3, create.ErrActionCheckingExistence, tfs3.ResNameBucketMetadataTableConfiguration, rs.Primary.ID, errors.New("no bucket is set"))
		}

		resp, err := tfs3.FindBucketMetadataTableConfigurationByBucket(ctx, conn, bucket)
		if err != nil {
			return create.Error(names.S3, create.ErrActionCheckingExistence, tfs3.ResNameBucketMetadataTableConfiguration, rs.Primary.ID, err)
		}

		*bucketmetadatatableconfiguration = *resp

		return nil
	}
}

func testAccBucketMetadataTableConfigurationConfig_base(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  force_destroy = true
  bucket        = %[1]q
}

resource "aws_s3tables_table_bucket" "test_destination" {
  name          = "%s-destination"
}
`, bucketName, bucketName)
}

func testAccBucketMetadataTableConfigurationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccBucketMetadataTableConfigurationConfig_base(rName), `
resource "aws_s3_bucket_metadata_table_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  metadata_table_configuration {
	s3_tables_destination {
	  table_bucket_arn = aws_s3tables_table_bucket.test_destination.arn
      table_name       = "s3metadata_test_uniq"
	}
  }
}
`)
}
