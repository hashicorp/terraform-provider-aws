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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3files "github.com/hashicorp/terraform-provider-aws/internal/service/s3files"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3FilesSynchronizationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3files.GetSynchronizationConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_synchronization_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSynchronizationConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSynchronizationConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSynchronizationConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "expiration_days_after_last_access", "30"),
					resource.TestCheckResourceAttr(resourceName, "import_size_less_than", "131072"),
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

func testAccCheckSynchronizationConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3files_synchronization_configuration" {
				continue
			}
			_, err := tfs3files.FindSyncConfigByID(ctx, conn, rs.Primary.Attributes["file_system_id"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("S3 Files Synchronization Configuration %s still exists", rs.Primary.Attributes["file_system_id"])
		}
		return nil
	}
}

func testAccCheckSynchronizationConfigurationExists(ctx context.Context, t *testing.T, n string, v *s3files.GetSynchronizationConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		conn := acctest.ProviderMeta(ctx, t).S3FilesClient(ctx)
		output, err := tfs3files.FindSyncConfigByID(ctx, conn, rs.Primary.Attributes["file_system_id"])
		if err != nil {
			return err
		}
		*v = *output
		return nil
	}
}

func testAccSynchronizationConfigurationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFileSystemConfig_basic(rName), `
resource "aws_s3files_synchronization_configuration" "test" {
  file_system_id                    = aws_s3files_file_system.test.file_system_id
  expiration_days_after_last_access = 30
  import_size_less_than             = 131072
}
`)
}

func TestAccS3FilesSynchronizationConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3files.GetSynchronizationConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3files_synchronization_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3FilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSynchronizationConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSynchronizationConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSynchronizationConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "expiration_days_after_last_access", "30"),
					resource.TestCheckResourceAttr(resourceName, "import_size_less_than", "131072"),
				),
			},
			{
				Config: testAccSynchronizationConfigurationConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSynchronizationConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "expiration_days_after_last_access", "7"),
					resource.TestCheckResourceAttr(resourceName, "import_size_less_than", "262144"),
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

func testAccSynchronizationConfigurationConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccFileSystemConfig_basic(rName), `
resource "aws_s3files_synchronization_configuration" "test" {
  file_system_id                    = aws_s3files_file_system.test.file_system_id
  expiration_days_after_last_access = 7
  import_size_less_than             = 262144
}
`)
}
