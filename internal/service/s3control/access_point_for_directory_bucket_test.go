// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlDirectoryAccessPoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accessPointName := fmt.Sprintf("%s--usw2-az2--xa-s3", sdkacctest.RandomWithPrefix("dap"))
	resourceName := "aws_s3_directory_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryAccessPointConfig_basic(bucketName, accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, bucketName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, accessPointName),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Internet"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", "true"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "0"),
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

func TestAccS3ControlAccessPointForDirectoryBucket_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) + "--usw2-az2--x-s3"
	accessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) + "--usw2-az2--xa-s3"
	resourceName := "aws_s3_directory_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointForDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryAccessPointConfig_basic(bucketName, accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceAccessPoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlDirectoryAccessPoint_policy(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	accessPointName := fmt.Sprintf("%s--usw2-az2--xa-s3", sdkacctest.RandomWithPrefix("dap"))
	resourceName := "aws_s3_directory_access_point.test"

	expectedPolicyText1 := func() string {
		return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:%[1]s:iam::%[3]s:root"
      },
      "Action": "s3:GetObject",
      "Resource": [
        "arn:%[1]s:s3express:%[2]s:%[3]s:accesspoint/%[4]s/object/*"
      ]
    }
  ]
}`, acctest.Partition(), acctest.Region(), acctest.AccountID(ctx), accessPointName)
	}

	expectedPolicyText2 := func() string {
		return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:%[1]s:iam::%[3]s:root"
      },
      "Action": [
        "s3:PutObject",
        "s3:GetObject"
      ],
      "Resource": [
        "arn:%[1]s:s3express:%[2]s:%[3]s:accesspoint/%[4]s/object/*"
      ]
    }
  ]
}`, acctest.Partition(), acctest.Region(), acctest.AccountID(ctx), accessPointName)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointForDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointForDirectoryBucketConfig_policy(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					testAccCheckAccessPointForDirectoryBucketHasPolicy(ctx, resourceName, expectedPolicyText1),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3express", fmt.Sprintf("accesspoint/%s", accessPointName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, accessPointName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, accessPointName),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Internet"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccessPointForDirectoryBucketConfig_policyUpdated(accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					testAccCheckAccessPointForDirectoryBucketHasPolicy(ctx, resourceName, expectedPolicyText2),
				),
			},
			{
				Config: testAccAccessPointForDirectoryBucketConfig_noPolicy(accessPointName),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointForDirectoryBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointForDirectoryBucketConfig_publicBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3express", fmt.Sprintf("accesspoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Internet"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "0"),
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

func TestAccS3ControlAccessPointForDirectoryBucket_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_access_point.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointForDirectoryBucketConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3express", fmt.Sprintf("accesspoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "VPC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_configuration.0.vpc_id", vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "0"),
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

func TestAccS3ControlAccessPointForDirectoryBucket_scope(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.GetAccessPointOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_directory_access_point.test"

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
				Config: testAccAccessPointForDirectoryBucketConfig_scope(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					testAccCheckAccessPointForDirectoryBucketHasScope(ctx, resourceName, expectedScope1),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "s3express", fmt.Sprintf("accesspoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "VPC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					// TODO: add more checks here for scope

				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccessPointForDirectoryBucketConfig_scopeUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					testAccCheckAccessPointForDirectoryBucketHasScope(ctx, resourceName, expectedScope2),
				),
			},
			{
				Config: testAccAccessPointForDirectoryBucketConfig_noScope(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointForDirectoryBucketExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "0"),
				),
			},
		},
	})
}

func testAccDirectoryAccessPointConfig_basic(bucketName, accessPointName string) string {
	return fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test_bucket" {
  bucket = "%s--usw2-az2--x-s3"
}

resource "aws_s3_directory_access_point" "test_ap" {
  bucket = aws_s3_directory_bucket.test_bucket.bucket
  name   = "%s--usw2-az2--xa-s3"
  account_id = data.aws_caller_identity.current.account_id
}
`, bucketName, accessPointName)
}

func testAccCheckAccessPointForDirectoryBucketDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_access_point" {
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
resource "aws_s3_directory_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_directory_access_point" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  name   = %[1]q
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
      "s3:GetObject",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3express:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:accesspoint/%[1]s/object/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}
`, rName)
}

func testAccCheckAccessPointForDirectoryBucketHasPolicy(ctx context.Context, n string, fn func() string) resource.TestCheckFunc {
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

		actualPolicyText, err := tfs3control.FindAccessPointForDirectoryBucketPolicyByTwoPartKey(ctx, conn, accountID, name)

		if err != nil {
			return err
		}

		expectedPolicyText := fn()

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccAccessPointForDirectoryBucketConfig_policyUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_directory_access_point" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  name   = %[1]q
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
      "s3:PutObject",
      "s3:GetObject"
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3express:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:accesspoint/%[1]s/object/*",
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
resource "aws_s3_directory_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_directory_access_point" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  name   = %[1]q

  public_access_block_configuration {
    block_public_acls       = true
    block_public_policy     = true
    ignore_public_acls      = true
    restrict_public_buckets = true
  }

  policy = "{}"
}
`, rName)
}

func testAccAccessPointForDirectoryBucketConfig_publicBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_directory_access_point" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  name   = %[1]q

  public_access_block_configuration {
    block_public_acls       = true
    block_public_policy     = true
    ignore_public_acls      = true
    restrict_public_buckets = true
  }
}
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

resource "aws_s3_directory_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_directory_access_point" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  name   = %[1]q

  vpc_configuration {
    vpc_id = aws_vpc.test.id
  }
}
`, rName)
}

func testAccAccessPointForDirectoryBucketConfig_scope(rName string) string {
	return fmt.Sprintf(`

resource "aws_s3_directory_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_directory_access_point" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  name   = %[1]q
  account_id = data.aws_caller_identity.current.account_id

  scope {
    permissions = ["GetObject", "PutObject"]
    prefixes    = ["prefix1/", "prefix2-*-*"]
  }
}
`, rName)
}

func testAccAccessPointForDirectoryBucketConfig_scopeUpdated(rName string) string {
	return fmt.Sprintf(`

resource "aws_s3_directory_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_directory_access_point" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  name   = %[1]q
  account_id = data.aws_caller_identity.current.account_id

  scope {
    permissions = ["GetObject"]
    prefixes    = ["*"]
  }
}
`, rName)
}

func testAccAccessPointForDirectoryBucketConfig_noScope(rName string) string {
	return fmt.Sprintf(`

resource "aws_s3_directory_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_directory_access_point" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  name   = %[1]q
  account_id = data.aws_caller_identity.current.account_id
}
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
