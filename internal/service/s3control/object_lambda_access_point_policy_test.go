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

func TestAccS3ControlObjectLambdaAccessPointPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_object_lambda_access_point_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectLambdaAccessPointPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointPolicyExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
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

func TestAccS3ControlObjectLambdaAccessPointPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_object_lambda_access_point_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectLambdaAccessPointPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceObjectLambdaAccessPointPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlObjectLambdaAccessPointPolicy_Disappears_accessPoint(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_object_lambda_access_point_policy.test"
	accessPointResourceName := "aws_s3control_object_lambda_access_point.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectLambdaAccessPointPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceObjectLambdaAccessPoint(), accessPointResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlObjectLambdaAccessPointPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_object_lambda_access_point_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectLambdaAccessPointPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointPolicyExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccObjectLambdaAccessPointPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointPolicyExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
		},
	})
}

func testAccCheckObjectLambdaAccessPointPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_object_lambda_access_point_policy" {
				continue
			}

			accountID, name, err := tfs3control.ObjectLambdaAccessPointParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, _, err = tfs3control.FindObjectLambdaAccessPointPolicyAndStatusByTwoPartKey(ctx, conn, accountID, name)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Object Lambda Access Point Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckObjectLambdaAccessPointPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		accountID, name, err := tfs3control.ObjectLambdaAccessPointParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		_, _, err = tfs3control.FindObjectLambdaAccessPointPolicyAndStatusByTwoPartKey(ctx, conn, accountID, name)

		return err
	}
}

func testAccObjectLambdaAccessPointPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccObjectLambdaAccessPointBaseConfig(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.id
  name   = %[1]q
}

resource "aws_s3control_object_lambda_access_point" "test" {
  name = %[1]q

  configuration {
    supporting_access_point = aws_s3_access_point.test.arn

    transformation_configuration {
      actions = ["GetObject"]

      content_transformation {
        aws_lambda {
          function_arn = aws_lambda_function.test.arn
        }
      }
    }
  }
}

resource "aws_s3control_object_lambda_access_point_policy" "test" {
  name = aws_s3control_object_lambda_access_point.test.name

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect = "Allow"
      Action = "s3-object-lambda:GetObject"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Resource = aws_s3control_object_lambda_access_point.test.arn
    }]
  })
}
`, rName))
}

func testAccObjectLambdaAccessPointPolicyConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccObjectLambdaAccessPointBaseConfig(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.id
  name   = %[1]q
}

resource "aws_s3control_object_lambda_access_point" "test" {
  name = %[1]q

  configuration {
    supporting_access_point = aws_s3_access_point.test.arn

    transformation_configuration {
      actions = ["GetObject"]

      content_transformation {
        aws_lambda {
          function_arn = aws_lambda_function.test.arn
        }
      }
    }
  }
}

resource "aws_s3control_object_lambda_access_point_policy" "test" {
  name = aws_s3control_object_lambda_access_point.test.name

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect = "Allow"
      Action = "s3-object-lambda:*"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Resource = aws_s3control_object_lambda_access_point.test.arn
    }]
  })
}
`, rName))
}
