// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	resourceName := "aws_s3control_directory_access_point_scope.test_scope"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	apName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointForDirectoryBucketScopeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointScopeConfig_basic(apName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketScopeExists(ctx, resourceName),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					name := rs.Primary.Attributes["name"]
					accountID := rs.Primary.Attributes["account_id"]

					if name == "" || accountID == "" {
						return "", fmt.Errorf("missing name or account_id in state")
					}

					return fmt.Sprintf("%s:%s", name, accountID), nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3ControlAccessPointScope_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_directory_access_point_scope.test_scope"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	apName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointForDirectoryBucketScopeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointScopeConfig_basic(apName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketScopeExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceAccessPointForDirectoryBucketScope(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlAccessPointScope_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_directory_access_point_scope.test_scope"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	apName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointForDirectoryBucketScopeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointScopeConfig_basic(apName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketScopeExists(ctx, resourceName),
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
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}

					name := rs.Primary.Attributes["name"]
					accountID := rs.Primary.Attributes["account_id"]

					if name == "" || accountID == "" {
						return "", fmt.Errorf("missing name or account_id in state")
					}

					return fmt.Sprintf("%s:%s", name, accountID), nil
				},

				ImportStateVerify: true,
			},
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointScopeConfig_updated(apName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketScopeExists(ctx, resourceName),
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

			return fmt.Errorf("S3 Access Point for Directory Bucket Scope still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAccessPointScopeConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAccessPointForDirectoryBucketConfig_basic(rName), `
resource "aws_s3control_directory_access_point_scope" "test_scope" {
  name       = aws_s3_directory_access_point.test_ap.name
  account_id = data.aws_caller_identity.current.account_id

  scope {
    permissions = ["GetObject", "PutObject"]
    prefixes    = ["prefix1/", "prefix2-*-*"]
  }
}
`)
}

func testAccAccessPointScopeConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccAccessPointForDirectoryBucketConfig_basic(rName), `
resource "aws_s3control_directory_access_point_scope" "test_scope" {
	  name       = aws_s3_directory_access_point.test_ap.name
	  account_id = data.aws_caller_identity.current.account_id
	
	  scope {
		permissions = ["GetObject"]
		prefixes    = ["prefix3/", "*prefix4*"]
	  }
	}
`)
}
