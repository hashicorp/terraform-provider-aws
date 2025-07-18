// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketMetadataConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.MetadataConfigurationResult
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_metadata_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetadataConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetadataConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketMetadataConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.0.journal_table_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "metadata_configuration.0.journal_table_configuration.0.table_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "metadata_configuration.0.journal_table_configuration.0.table_name"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrBucket),
				ImportStateVerifyIdentifierAttribute: names.AttrBucket,
			},
		},
	})
}

func TestAccS3BucketMetadataConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.MetadataConfigurationResult
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_metadata_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketMetadataConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketMetadataConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketMetadataConfigurationExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucketMetadataConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBucketMetadataConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_metadata_configuration" {
				continue
			}

			_, err := tfs3.FindBucketMetadataConfigurationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrBucket], rs.Primary.Attributes[names.AttrExpectedBucketOwner])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Metadata Configuration %s still exists", rs.Primary.Attributes[names.AttrBucket])
		}

		return nil
	}
}

func testAccCheckBucketMetadataConfigurationExists(ctx context.Context, n string, v *awstypes.MetadataConfigurationResult) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		output, err := tfs3.FindBucketMetadataConfigurationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrBucket], rs.Primary.Attributes[names.AttrExpectedBucketOwner])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccBucketMetadataConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_metadata_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  metadata_configuration {
    journal_table_configuration {
      record_expiration {
        days       = 7
        expiration = "ENABLED"
      }
    }
  }
}
`, rName)
}
