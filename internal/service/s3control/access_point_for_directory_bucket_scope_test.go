// // Copyright (c) HashiCorp, Inc.
// // SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlAccessPointForDirectoryBucketScope_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_directory_access_point_scope.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointForDirectoryBucketScopeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointForDirectoryBucketScopeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketScopeExists(ctx, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAccessPointForDirectoryBucketScopeExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		name, accountID, err := tfs3control.AccessPointForDirectoryBucketParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		_, err = tfs3control.FindAccessPointScopeByTwoPartKey(ctx, conn, accountID, name)
		if err != nil {
			return fmt.Errorf("error finding S3 Access Point for Directory Bucket Scope (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckAccessPointForDirectoryBucketScopeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_directory_access_point_scope" {
				continue
			}

			name, accountID, err := tfs3control.AccessPointForDirectoryBucketParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3control.FindAccessPointScopeByTwoPartKey(ctx, conn, accountID, name)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Directory Access Point Scope still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAccessPointForDirectoryBucketScopeConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_directory_bucket" "test_bucket" {
  bucket = "terraformbucket-revv--apse1-sggov-sin2-az1--x-s3"

  location {
    name = "apse1-sggov-sin2-az1"
    type = "LocalZone"
  }
}

resource "aws_s3_directory_access_point" "test_ap" {
  bucket     = aws_s3_directory_bucket.test_bucket.id
  name       = "%[1]s--apse1-sggov-sin2-az1--xa-s3"
  account_id = data.aws_caller_identity.current.account_id
}

resource "aws_s3control_directory_access_point_scope" "test_scope" {
  name       = aws_s3_directory_access_point.test_ap.name
  account_id = data.aws_caller_identity.current.account_id

  scope {
    permissions = ["s3:GetObject", "s3:PutObject"]
    prefixes    = ["prefix1/", "prefix2-*-*"]
  }
}
`, rName)
}
