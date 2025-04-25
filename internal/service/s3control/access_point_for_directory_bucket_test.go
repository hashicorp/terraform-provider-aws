// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlDirectoryAccessPoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_access_point.test_ap"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointForDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointForDirectoryBucketConfig_basic(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Internet"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", "true"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.permissions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.prefixes.#", "0"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("s3express", regexache.MustCompile(`accesspoint/.+--xa-s3`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.StringRegexp(tfs3control.DirectoryBucketNameRegex)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringRegexp(tfs3control.AccessPointForDirectoryBucketNameRegex)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3ControlAccessPointForDirectoryBucket_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_access_point.test_ap"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointForDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointForDirectoryBucketConfig_basic(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceAccessPointForDirectoryBucket(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlDirectoryAccessPoint_policy(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_access_point.test_ap"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointForDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) +
					testAccAccessPointForDirectoryBucketConfig_policy(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					testAccCheckAccessPointForDirectoryBucketHasPolicy(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Internet"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.permissions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.prefixes.#", "0"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("s3express", regexache.MustCompile(`accesspoint/.+--xa-s3`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.StringRegexp(tfs3control.DirectoryBucketNameRegex)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringRegexp(tfs3control.AccessPointForDirectoryBucketNameRegex)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointForDirectoryBucketConfig_policyUpdated(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					testAccCheckAccessPointForDirectoryBucketHasPolicy(ctx, resourceName),
				),
			},
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointForDirectoryBucketConfig_noPolicy(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
				),
			},
		},
	})
}

func TestAccS3ControlAccessPointForDirectoryBucket_publicAccessBlock(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_access_point.test_ap"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointForDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointForDirectoryBucketConfig_publicBlock(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Internet"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.permissions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.prefixes.#", "0"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("s3express", regexache.MustCompile(`accesspoint/.+--xa-s3`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.StringRegexp(tfs3control.DirectoryBucketNameRegex)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringRegexp(tfs3control.AccessPointForDirectoryBucketNameRegex)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3ControlAccessPointForDirectoryBucket_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_access_point.test_ap"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointForDirectoryBucketConfig_vpc(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "VPC"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.%", "2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_configuration.0.vpc_id", vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.permissions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.prefixes.#", "0"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("s3express", regexache.MustCompile(`accesspoint/.+--xa-s3`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.StringRegexp(tfs3control.DirectoryBucketNameRegex)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringRegexp(tfs3control.AccessPointForDirectoryBucketNameRegex)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3ControlAccessPointForDirectoryBucket_scope(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_access_point.test_ap"

	expectedScope1 := func() *types.Scope {

		permissions := []types.ScopePermission{
			types.ScopePermission("GetObject"),
			types.ScopePermission("PutObject"),
		}

		prefixes := []string{"prefix1/", "prefix2-*-*"}

		return &types.Scope{
			Permissions: permissions,
			Prefixes:    prefixes,
		}
	}

	expectedScope2 := func() *types.Scope {

		permissions := []types.ScopePermission{
			types.ScopePermission("GetObject"),
		}
		prefixes := []string{"*"}

		return &types.Scope{
			Permissions: permissions,
			Prefixes:    prefixes,
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointForDirectoryBucketConfig_scope(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					testAccCheckAccessPointForDirectoryBucketHasScope(ctx, resourceName, expectedScope1),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Internet"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.permissions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.permissions.0", "GetObject"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.permissions.1", "PutObject"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.prefixes.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.prefixes.0", "prefix1/"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.prefixes.1", "prefix2-*-*"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("s3express", regexache.MustCompile(`accesspoint/.+--xa-s3`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrBucket), knownvalue.StringRegexp(tfs3control.DirectoryBucketNameRegex)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringRegexp(tfs3control.AccessPointForDirectoryBucketNameRegex)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointForDirectoryBucketConfig_scopeUpdated(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					testAccCheckAccessPointForDirectoryBucketHasScope(ctx, resourceName, expectedScope2),
					resource.TestCheckResourceAttr(resourceName, "scope.0.permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.permissions.0", "GetObject"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.prefixes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.prefixes.0", "*"),
				),
			},
			{
				Config: testAccDirectoryBucketConfig_basic(bucketName) + testAccAccessPointForDirectoryBucketConfig_noScope(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.permissions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope.0.prefixes.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAccessPointForDirectoryBucketDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_directory_access_point" {
				continue
			}

			name, accountID, err := tfs3control.AccessPointForDirectoryBucketParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3control.FindAccessPointByTwoPartKey(ctx, conn, accountID, name)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Access Point for directory bucket %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccessPointForDirectoryBucketExists(ctx context.Context, n string, v *s3control.GetAccessPointOutput) resource.TestCheckFunc {
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

		output, err := tfs3control.FindAccessPointByTwoPartKey(ctx, conn, accountID, name)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAccessPointForDirectoryBucketConfig_policy(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_directory_access_point" "test_ap" {
  bucket     = aws_s3_directory_bucket.test_bucket.bucket
  name       = "%[1]s--${local.location_name}--xa-s3"
  account_id = data.aws_caller_identity.current.account_id
  policy = data.aws_iam_policy_document.test.json

  public_access_block_configuration {
    block_public_acls       = true
    block_public_policy     = true
    ignore_public_acls      = true
    restrict_public_buckets = true
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    actions = [
      "s3express:CreateSession",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3express:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:accesspoint/%[1]s--${local.location_name}--xa-s3",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}
`, rName)
}

func testAccCheckAccessPointForDirectoryBucketHasPolicy(ctx context.Context, n string) resource.TestCheckFunc {
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

		_, err = tfs3control.FindAccessPointForDirectoryBucketPolicyByTwoPartKey(ctx, conn, accountID, name)
		return err
	}
}

func testAccAccessPointForDirectoryBucketConfig_policyUpdated(rName string) string {
	return fmt.Sprintf(`

resource "aws_s3_directory_access_point" "test_ap" {
  bucket     = aws_s3_directory_bucket.test_bucket.bucket
  name       = "%[1]s--${local.location_name}--xa-s3"
  account_id = data.aws_caller_identity.current.account_id

  public_access_block_configuration {
    block_public_acls       = true
    block_public_policy     = true
    ignore_public_acls      = true
    restrict_public_buckets = true
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    actions = [
      "s3express:CreateSession",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3express:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:accesspoint/%[1]s--${local.location_name}--xa-s3",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}
`, rName)
}

func testAccAccessPointForDirectoryBucketConfig_noPolicy(rName string) string {
	return fmt.Sprintf(`

resource "aws_s3_directory_access_point" "test_ap" {
  bucket     = aws_s3_directory_bucket.test_bucket.bucket
  name       = "%[1]s--${local.location_name}--xa-s3"
  account_id = data.aws_caller_identity.current.account_id

  public_access_block_configuration {
    block_public_acls       = true
    block_public_policy     = true
    ignore_public_acls      = true
    restrict_public_buckets = true
  }

  policy = "{}"
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

`, rName)
}

func testAccAccessPointForDirectoryBucketConfig_publicBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_directory_access_point" "test_ap" {
  bucket     = aws_s3_directory_bucket.test_bucket.bucket
  name       = "%[1]s--${local.location_name}--xa-s3"
  account_id = data.aws_caller_identity.current.account_id

  public_access_block_configuration {
    block_public_acls       = true
    block_public_policy     = true
    ignore_public_acls      = true
    restrict_public_buckets = true
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

`, rName)
}

func testAccAccessPointForDirectoryBucketConfig_vpc(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_s3_directory_access_point" "test_ap" {
  bucket     = aws_s3_directory_bucket.test_bucket.bucket
  name       = "%[1]s--${local.location_name}--xa-s3"
  account_id = data.aws_caller_identity.current.account_id

  vpc_configuration {
    vpc_id = aws_vpc.test.id
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}
`, rName)
}

func testAccAccessPointForDirectoryBucketConfig_scope(rName string) string {
	return fmt.Sprintf(`

resource "aws_s3_directory_access_point" "test_ap" {
  bucket     = aws_s3_directory_bucket.test_bucket.bucket
  name       = "%[1]s--${local.location_name}--xa-s3"
  account_id = data.aws_caller_identity.current.account_id

  scope {
    permissions = ["GetObject", "PutObject"]
    prefixes    = ["prefix1/", "prefix2-*-*"]
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}
`, rName)
}

func testAccAccessPointForDirectoryBucketConfig_scopeUpdated(rName string) string {
	return fmt.Sprintf(`

resource "aws_s3_directory_access_point" "test_ap" {
  bucket     = aws_s3_directory_bucket.test_bucket.bucket
  name       = "%[1]s--${local.location_name}--xa-s3"
  account_id = data.aws_caller_identity.current.account_id

  scope {
    permissions = ["GetObject"]
    prefixes    = ["*"]
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}
`, rName)
}

func testAccAccessPointForDirectoryBucketConfig_noScope(rName string) string {
	return fmt.Sprintf(`

resource "aws_s3_directory_access_point" "test_ap" {
  bucket     = aws_s3_directory_bucket.test_bucket.bucket
  name       = "%[1]s--${local.location_name}--xa-s3"
  account_id = data.aws_caller_identity.current.account_id
  scope {
    permissions = []
    prefixes    = []
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}
`, rName)
}

func testAccCheckAccessPointForDirectoryBucketHasScope(ctx context.Context, n string, fn func() *types.Scope) resource.TestCheckFunc {
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

		actualScope, err := tfs3control.FindAccessPointScopeByTwoPartKey(ctx, conn, accountID, name)
		if err != nil {
			return err
		}

		expectedScope := fn()

		sort.Slice(actualScope.Permissions, func(i, j int) bool {
			return actualScope.Permissions[i] < actualScope.Permissions[j]
		})

		sort.Slice(expectedScope.Permissions, func(i, j int) bool {
			return expectedScope.Permissions[i] < expectedScope.Permissions[j]
		})

		if len(actualScope.Permissions) != len(expectedScope.Permissions) {
			return fmt.Errorf("scope permissions count mismatch:\nexpected: %#v\ngot: %#v", expectedScope.Permissions, actualScope.Permissions)
		}

		for i := range actualScope.Permissions {
			if actualScope.Permissions[i] != expectedScope.Permissions[i] {
				return fmt.Errorf("scope permissions differ at index %d: expected %s, got %s", i, expectedScope.Permissions[i], actualScope.Permissions[i])
			}
		}

		sort.Strings(actualScope.Prefixes)
		sort.Strings(expectedScope.Prefixes)

		if len(actualScope.Prefixes) != len(expectedScope.Prefixes) {
			return fmt.Errorf("scope prefixes count mismatch:\nexpected: %#v\ngot: %#v", expectedScope.Prefixes, actualScope.Prefixes)
		}
		for i := range actualScope.Prefixes {
			if actualScope.Prefixes[i] != expectedScope.Prefixes[i] {
				return fmt.Errorf("scope prefixes differ at index %d: expected %s, got %s", i, expectedScope.Prefixes[i], actualScope.Prefixes[i])
			}
		}

		return nil
	}
}

func testAccConfigDirectoryBucket_availableAZs() string {
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide/directory-bucket-az-networking.html#s3-express-endpoints-az.
	return acctest.ConfigAvailableAZsNoOptInExclude("use2-az2", "use1-az1", "use1-az2", "use1-az3", "usw2-az2", "aps1-az3", "apne1-az2", "euw1-az2")
}

func testAccDirectoryBucketConfig_baseAZ(rName string) string {
	return acctest.ConfigCompose(testAccConfigDirectoryBucket_availableAZs(), fmt.Sprintf(`
locals {
  location_name = data.aws_availability_zones.available.zone_ids[0]
  bucket        = "%[1]s--${local.location_name}--x-s3"
}
`, rName))
}

func testAccDirectoryBucketConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_baseAZ(rName), `
resource "aws_s3_directory_bucket" "test_bucket" {
  bucket = local.bucket

  location {
    name = local.location_name
  }

  force_destroy = true
}
`)
}

func testAccAccessPointForDirectoryBucketConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_s3_directory_access_point" "test_ap" {
  bucket     = aws_s3_directory_bucket.test_bucket.bucket
  name       = "%[1]s--${local.location_name}--xa-s3"
  account_id = data.aws_caller_identity.current.account_id
}
`, rName)
}
