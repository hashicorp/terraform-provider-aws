// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLakeFormationResource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource.test"
	bucketResourceName := "aws_s3_bucket.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_basic(bucketName, roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, bucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "hybrid_access_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "with_federation", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccLakeFormationResource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflakeformation.ResourceResource(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLakeFormationResource_serviceLinkedRole(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/lakeformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_serviceLinkedRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, bucketResourceName, names.AttrARN),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrRoleARN, "iam", "role/aws-service-role/lakeformation.amazonaws.com/AWSServiceRoleForLakeFormationDataAccess"),
				),
			},
		},
	})
}

func TestAccLakeFormationResource_updateRoleToRole(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource.test"
	bucketResourceName := "aws_s3_bucket.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_basic(bucketName, roleName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, bucketResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccResourceConfig_basic(bucketName, roleName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, bucketResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccLakeFormationResource_updateSLRToRole(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource.test"
	bucketResourceName := "aws_s3_bucket.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/lakeformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_serviceLinkedRole(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, bucketResourceName, names.AttrARN),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrRoleARN, "iam", "role/aws-service-role/lakeformation.amazonaws.com/AWSServiceRoleForLakeFormationDataAccess"),
				),
			},
			{
				Config: testAccResourceConfig_basic(bucketName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, bucketResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccLakeFormationResource_hybridAccessEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_resource.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_hybridAccessEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, bucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "hybrid_access_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

// AWS does not support changing from an IAM role to an SLR. No error is thrown
// but the registration is not changed (the IAM role continues in the registration).
//
// func TestAccLakeFormationResource_updateRoleToSLR(t *testing.T) {

func testAccCheckResourceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_resource" {
				continue
			}

			_, err := tflakeformation.FindResourceByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccCheckResourceExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

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
