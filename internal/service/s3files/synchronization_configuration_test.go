// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3files"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfs3files "github.com/hashicorp/terraform-provider-aws/internal/service/s3files"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3FilesSynchronizationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var synchronization s3files.GetSynchronizationConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_synchronization_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSynchronizationConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSynchronizationConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSynchronizationConfigurationExists(ctx, t, resourceName, &synchronization),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrFileSystemID),
					resource.TestCheckResourceAttr(resourceName, "import_data_rule.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "expiration_data_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "expiration_data_rule.0.days_after_last_access", "30"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrFileSystemID),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrFileSystemID,
			},
		},
	})
}

func TestAccS3FilesSynchronizationConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var synchronization s3files.GetSynchronizationConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_synchronization_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSynchronizationConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSynchronizationConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSynchronizationConfigurationExists(ctx, t, resourceName, &synchronization),
					resource.TestCheckResourceAttr(resourceName, "expiration_data_rule.0.days_after_last_access", "30"),
				),
			},
			{
				Config: testAccSynchronizationConfigurationConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSynchronizationConfigurationExists(ctx, t, resourceName, &synchronization),
					resource.TestCheckResourceAttr(resourceName, "expiration_data_rule.0.days_after_last_access", "60"),
				),
			},
		},
	})
}

func testAccCheckSynchronizationConfigurationExists(ctx context.Context, t *testing.T, n string, v *s3files.GetSynchronizationConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)

		output, err := tfs3files.FindSynchronizationConfigurationByFileSystemID(ctx, conn, rs.Primary.Attributes[names.AttrFileSystemID])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSynchronizationConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3files_synchronization_configuration" {
				continue
			}

			_, err := tfs3files.FindSynchronizationConfigurationByFileSystemID(ctx, conn, rs.Primary.Attributes[names.AttrFileSystemID])

			if err == nil {
				return fmt.Errorf("S3 Files Synchronization %s still exists", rs.Primary.Attributes[names.AttrFileSystemID])
			}
		}

		return nil
	}
}

func testAccSynchronizationConfigurationConfig_base(rName string) string {
	return acctest.ConfigCompose(
		testAccFileSystemConfig_base(rName),
		`
resource "aws_s3files_file_system" "test" {
  bucket   = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn

  depends_on = [aws_s3_bucket_versioning.test]
}
`)
}

func testAccSynchronizationConfigurationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccSynchronizationConfigurationConfig_base(rName),
		`
resource "aws_s3files_synchronization_configuration" "test" {
  file_system_id = aws_s3files_file_system.test.id

  import_data_rule {
    prefix         = ""
    size_less_than = 52673613135872
    trigger        = "ON_FILE_ACCESS"
  }

  import_data_rule {
    prefix         = "data/"
    size_less_than = 1048576
    trigger        = "ON_FILE_ACCESS"
  }

  expiration_data_rule {
    days_after_last_access = 30
  }
}
`)
}

func testAccSynchronizationConfigurationConfig_updated(rName string) string {
	return acctest.ConfigCompose(
		testAccSynchronizationConfigurationConfig_base(rName),
		`
resource "aws_s3files_synchronization_configuration" "test" {
  file_system_id = aws_s3files_file_system.test.id

  import_data_rule {
    prefix         = ""
    size_less_than = 52673613135872
    trigger        = "ON_FILE_ACCESS"
  }

  import_data_rule {
    prefix         = "data/"
    size_less_than = 1048576
    trigger        = "ON_FILE_ACCESS"
  }

  expiration_data_rule {
    days_after_last_access = 60
  }
}
`)
}

func TestAccS3FilesSynchronizationConfiguration_prefix(t *testing.T) {
	ctx := acctest.Context(t)
	var synchronization s3files.GetSynchronizationConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_synchronization_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSynchronizationConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSynchronizationConfigurationConfig_prefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSynchronizationConfigurationExists(ctx, t, resourceName, &synchronization),
					resource.TestCheckResourceAttr(resourceName, "import_data_rule.0.prefix", "data/"),
				),
			},
		},
	})
}

func testAccSynchronizationConfigurationConfig_prefix(rName string) string {
	return acctest.ConfigCompose(
		testAccFileSystemConfig_base(rName),
		`
resource "aws_s3files_file_system" "test" {
  bucket   = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn
  prefix   = "data/"

  depends_on = [aws_s3_bucket_versioning.test]
}

resource "aws_s3files_synchronization_configuration" "test" {
  file_system_id = aws_s3files_file_system.test.id

  import_data_rule {
    prefix         = "data/"
    size_less_than = 131072
    trigger        = "ON_DIRECTORY_FIRST_ACCESS"
  }

  expiration_data_rule {
    days_after_last_access = 30
  }
}
`)
}
