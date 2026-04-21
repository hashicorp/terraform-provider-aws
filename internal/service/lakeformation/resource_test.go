// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLakeFormationResource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	roleName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource.test"
	bucketResourceName := "aws_s3_bucket.test"
	roleResourceName := "aws_iam_role.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_basic(bucketName, roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, bucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "hybrid_access_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "with_federation", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "with_privileged_access", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccLakeFormationResource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflakeformation.ResourceResource(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLakeFormationResource_serviceLinkedRole(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource.test"
	bucketResourceName := "aws_s3_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/lakeformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_serviceLinkedRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, bucketResourceName, names.AttrARN),
					acctest.CheckResourceAttrGlobalARN(ctx, resourceName, names.AttrRoleARN, "iam", "role/aws-service-role/lakeformation.amazonaws.com/AWSServiceRoleForLakeFormationDataAccess"),
				),
			},
		},
	})
}

func TestAccLakeFormationResource_updateRoleToRole(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	roleName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	roleName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource.test"
	bucketResourceName := "aws_s3_bucket.test"
	roleResourceName := "aws_iam_role.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_basic(bucketName, roleName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, bucketResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccResourceConfig_basic(bucketName, roleName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, bucketResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccLakeFormationResource_updateSLRToRole(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	roleName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource.test"
	bucketResourceName := "aws_s3_bucket.test"
	roleResourceName := "aws_iam_role.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/lakeformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_serviceLinkedRole(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, bucketResourceName, names.AttrARN),
					acctest.CheckResourceAttrGlobalARN(ctx, resourceName, names.AttrRoleARN, "iam", "role/aws-service-role/lakeformation.amazonaws.com/AWSServiceRoleForLakeFormationDataAccess"),
				),
			},
			{
				Config: testAccResourceConfig_basic(bucketName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, bucketResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccLakeFormationResource_hybridAccessEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource.test"
	bucketResourceName := "aws_s3_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_hybridAccessEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, bucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "hybrid_access_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccLakeFormationResource_withPrivilegedAccessEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource.test"
	bucketResourceName := "aws_s3_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_withPrivilegedAccessEnabled(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, bucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "with_privileged_access", acctest.CtTrue),
				),
			},
		},
	})
}

// AWS does not support changing from an IAM role to an SLR. No error is thrown
// but the registration is not changed (the IAM role continues in the registration).
//
// func TestAccLakeFormationResource_updateRoleToSLR(t *testing.T) {

func testAccCheckResourceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LakeFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_resource" {
				continue
			}

			_, err := tflakeformation.FindResourceByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lake Formation Resource (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckResourceExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LakeFormationClient(ctx)

		_, err := tflakeformation.FindResourceByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccResourceConfig_basic(bucket, role string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[2]q
  path = "/test/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "test" {
  name = %[2]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets",
        "s3:GetObjectVersion",
        "s3:GetBucketAcl",
        "s3:GetObject",
        "s3:GetObjectACL",
        "s3:PutObject",
        "s3:PutObjectAcl"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*",
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
      ]
    }
  ]
}
EOF
}

resource "aws_lakeformation_resource" "test" {
  arn      = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn
}
`, bucket, role)
}

func testAccResourceConfig_serviceLinkedRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_lakeformation_resource" "test" {
  arn                     = aws_s3_bucket.test.arn
  use_service_linked_role = true
}
`, rName)
}

func testAccResourceConfig_hybridAccessEnabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_lakeformation_resource" "test" {
  arn                   = aws_s3_bucket.test.arn
  hybrid_access_enabled = true
}
`, rName)
}

func testAccResourceConfig_withPrivilegedAccessEnabled(bucket, role string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[2]q
  path = "/test/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "test" {
  name = %[2]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets",
        "s3:GetObjectVersion",
        "s3:GetBucketAcl",
        "s3:GetObject",
        "s3:GetObjectACL",
        "s3:PutObject",
        "s3:PutObjectAcl"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*",
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
      ]
    }
  ]
}
EOF
}

resource "aws_lakeformation_resource" "test" {
  arn                    = aws_s3_bucket.test.arn
  role_arn               = aws_iam_role.test.arn
  with_privileged_access = true
}
`, bucket, role)
}
