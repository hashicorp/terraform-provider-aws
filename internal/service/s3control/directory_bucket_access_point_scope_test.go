// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlDirectoryBucketAccessPointScope_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_directory_bucket_access_point_scope.test"
	accessPointName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketAccessPointScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointScopeConfig_basic(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketScopeExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "scope.0.permissions.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "scope.0.permissions.0"),
					resource.TestCheckTypeSetElemAttr(resourceName, "scope.0.permissions.*", "GetObject"),
					resource.TestCheckResourceAttrSet(resourceName, "scope.0.permissions.1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "scope.0.permissions.*", "PutObject"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.prefixes.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "scope.0.prefixes.0"),
					resource.TestCheckTypeSetElemAttr(resourceName, "scope.0.prefixes.*", "prefix1/"),
					resource.TestCheckResourceAttrSet(resourceName, "scope.0.prefixes.1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "scope.0.prefixes.*", "prefix2-*-*"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    directoryBucketAccessPointScopeStateImportFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerify:                    true,
			},
		},
	})
}

func TestAccS3ControlDirectoryBucketAccessPointScope_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_directory_bucket_access_point_scope.test"
	accessPointName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketAccessPointScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointScopeConfig_basic(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketScopeExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfs3control.ResourceDirectoryBucketAccessPointScope, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlDirectoryBucketAccessPointScope_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_directory_bucket_access_point_scope.test"
	accessPointName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryBucketAccessPointScopeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointScopeConfig_basic(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketScopeExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "scope.0.permissions.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "scope.0.permissions.0"),
					resource.TestCheckTypeSetElemAttr(resourceName, "scope.0.permissions.*", "GetObject"),
					resource.TestCheckResourceAttrSet(resourceName, "scope.0.permissions.1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "scope.0.permissions.*", "PutObject"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.prefixes.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "scope.0.prefixes.0"),
					resource.TestCheckTypeSetElemAttr(resourceName, "scope.0.prefixes.*", "prefix1/"),
					resource.TestCheckResourceAttrSet(resourceName, "scope.0.prefixes.1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "scope.0.prefixes.*", "prefix2-*-*"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    directoryBucketAccessPointScopeStateImportFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerify:                    true,
			},
			{
				Config: testAccAccessPointScopeConfig_updated(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketScopeExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "scope.0.permissions.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "scope.0.permissions.0"),
					resource.TestCheckTypeSetElemAttr(resourceName, "scope.0.permissions.*", "GetObject"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.prefixes.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "scope.0.prefixes.0"),
					resource.TestCheckTypeSetElemAttr(resourceName, "scope.0.prefixes.*", "prefix3/"),
					resource.TestCheckResourceAttrSet(resourceName, "scope.0.prefixes.1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "scope.0.prefixes.*", "*prefix4*"),
				),
			},
		},
	})
}

func directoryBucketAccessPointScopeStateImportFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrAccountID]), nil
	}
}

func testAccCheckAccessPointForDirectoryBucketScopeExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3ControlClient(ctx)
		_, err := tfs3control.FindDirectoryAccessPointScopeByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAccountID], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return create.Error(names.S3Control, create.ErrActionCheckingExistence, tfs3control.ResNameDirectoryBucketAccessPointScope, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckDirectoryBucketAccessPointScopeDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_directory_bucket_access_point_scope" {
				continue
			}

			_, err := tfs3control.FindDirectoryAccessPointScopeByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAccountID], rs.Primary.Attributes[names.AttrName])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.S3Control, create.ErrActionCheckingDestroyed, tfs3control.ResNameDirectoryBucketAccessPointScope, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAccessPointScopeConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAccessPointConfig_directoryBucketBasic(rName), `
data "aws_caller_identity" "current" {}

resource "aws_s3control_directory_bucket_access_point_scope" "test" {
  name       = aws_s3_access_point.test.name
  account_id = data.aws_caller_identity.current.account_id

  scope {
    permissions = ["GetObject", "PutObject"]
    prefixes    = ["prefix1/", "prefix2-*-*"]
  }
}
`)
}

func testAccAccessPointScopeConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccAccessPointConfig_directoryBucketBasic(rName), `
data "aws_caller_identity" "current" {}

resource "aws_s3control_directory_bucket_access_point_scope" "test" {
  name       = aws_s3_access_point.test.name
  account_id = data.aws_caller_identity.current.account_id

  scope {
    permissions = ["GetObject"]
    prefixes    = ["prefix3/", "*prefix4*"]
  }
}
`)
}
