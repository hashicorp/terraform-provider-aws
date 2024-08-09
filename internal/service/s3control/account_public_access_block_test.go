// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// S3 account-level settings must run serialized
// for TeamCity environment
func TestAccS3ControlAccountPublicAccessBlock_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"PublicAccessBlock": {
			acctest.CtBasic:         testAccAccountPublicAccessBlock_basic,
			acctest.CtDisappears:    testAccAccountPublicAccessBlock_disappears,
			"AccountId":             testAccAccountPublicAccessBlock_AccountID,
			"BlockPublicAcls":       testAccAccountPublicAccessBlock_BlockPublicACLs,
			"BlockPublicPolicy":     testAccAccountPublicAccessBlock_BlockPublicPolicy,
			"IgnorePublicAcls":      testAccAccountPublicAccessBlock_IgnorePublicACLs,
			"RestrictPublicBuckets": testAccAccountPublicAccessBlock_RestrictPublicBuckets,
			"DataSourceBasic":       testAccAccountPublicAccessBlockDataSource_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 5*time.Second)
}

func testAccAccountPublicAccessBlock_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPublicAccessBlockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", acctest.CtFalse),
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

func testAccAccountPublicAccessBlock_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPublicAccessBlockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceAccountPublicAccessBlock(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAccountPublicAccessBlock_AccountID(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPublicAccessBlockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
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

func testAccAccountPublicAccessBlock_BlockPublicACLs(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPublicAccessBlockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_acls(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_acls(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", acctest.CtFalse),
				),
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_acls(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccAccountPublicAccessBlock_BlockPublicPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPublicAccessBlockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_policy(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_policy(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", acctest.CtFalse),
				),
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_policy(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccAccountPublicAccessBlock_IgnorePublicACLs(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPublicAccessBlockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_ignoreACLs(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_ignoreACLs(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", acctest.CtFalse),
				),
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_ignoreACLs(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccAccountPublicAccessBlock_RestrictPublicBuckets(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPublicAccessBlockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_restrictBuckets(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_restrictBuckets(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", acctest.CtFalse),
				),
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_restrictBuckets(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckAccountPublicAccessBlockExists(ctx context.Context, n string, v *types.PublicAccessBlockConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		output, err := tfs3control.FindPublicAccessBlockByAccountID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAccountPublicAccessBlockDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_account_public_access_block" {
				continue
			}

			_, err := tfs3control.FindPublicAccessBlockByAccountID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Account Public Access Block %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAccountPublicAccessBlockConfig_basic() string {
	return `resource "aws_s3_account_public_access_block" "test" {}`
}

func testAccAccountPublicAccessBlockConfig_id() string {
	return `
data "aws_caller_identity" "test" {}

resource "aws_s3_account_public_access_block" "test" {
  account_id = data.aws_caller_identity.test.account_id
}
`
}

func testAccAccountPublicAccessBlockConfig_acls(blockPublicAcls bool) string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  block_public_acls = %[1]t
}
`, blockPublicAcls)
}

func testAccAccountPublicAccessBlockConfig_policy(blockPublicPolicy bool) string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  block_public_policy = %[1]t
}
`, blockPublicPolicy)
}

func testAccAccountPublicAccessBlockConfig_ignoreACLs(ignorePublicAcls bool) string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  ignore_public_acls = %[1]t
}
`, ignorePublicAcls)
}

func testAccAccountPublicAccessBlockConfig_restrictBuckets(restrictPublicBuckets bool) string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  restrict_public_buckets = %[1]t
}
`, restrictPublicBuckets)
}
