// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlObjectLambdaAccessPoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.ObjectLambdaConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_object_lambda_access_point.test"
	accessPointResourceName := "aws_s3_access_point.test"
	lambdaFunctionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectLambdaAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					resource.TestMatchResourceAttr(resourceName, names.AttrAlias, regexache.MustCompile("^.{1,20}-.*--ol-s3$")),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "s3-object-lambda", fmt.Sprintf("accesspoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.allowed_features.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cloud_watch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.supporting_access_point", accessPointResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.transformation_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration.0.transformation_configuration.*", map[string]string{
						"actions.#":                             acctest.Ct1,
						"content_transformation.#":              acctest.Ct1,
						"content_transformation.0.aws_lambda.#": acctest.Ct1,
						"content_transformation.0.aws_lambda.0.function_payload": "",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.transformation_configuration.*.actions.*", "GetObject"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "configuration.0.transformation_configuration.*.content_transformation.0.aws_lambda.0.function_arn", lambdaFunctionResourceName, names.AttrARN),
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

func TestAccS3ControlObjectLambdaAccessPoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.ObjectLambdaConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_object_lambda_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectLambdaAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceObjectLambdaAccessPoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlObjectLambdaAccessPoint_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.ObjectLambdaConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_object_lambda_access_point.test"
	accessPointResourceName := "aws_s3_access_point.test"
	lambdaFunctionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectLambdaAccessPointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointConfig_optionals(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "s3-object-lambda", fmt.Sprintf("accesspoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.allowed_features.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.allowed_features.*", "GetObject-PartNumber"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.allowed_features.*", "GetObject-Range"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cloud_watch_metrics_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.supporting_access_point", accessPointResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.transformation_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration.0.transformation_configuration.*", map[string]string{
						"actions.#":                             acctest.Ct1,
						"content_transformation.#":              acctest.Ct1,
						"content_transformation.0.aws_lambda.#": acctest.Ct1,
						"content_transformation.0.aws_lambda.0.function_payload": "{\"res-x\": \"100\",\"res-y\": \"100\"}",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.transformation_configuration.*.actions.*", "GetObject"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "configuration.0.transformation_configuration.*.content_transformation.0.aws_lambda.0.function_arn", lambdaFunctionResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccObjectLambdaAccessPointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "s3-object-lambda", fmt.Sprintf("accesspoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.allowed_features.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cloud_watch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.supporting_access_point", accessPointResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.transformation_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration.0.transformation_configuration.*", map[string]string{
						"actions.#":                             acctest.Ct1,
						"content_transformation.#":              acctest.Ct1,
						"content_transformation.0.aws_lambda.#": acctest.Ct1,
						"content_transformation.0.aws_lambda.0.function_payload": "",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.transformation_configuration.*.actions.*", "GetObject"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "configuration.0.transformation_configuration.*.content_transformation.0.aws_lambda.0.function_arn", lambdaFunctionResourceName, names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckObjectLambdaAccessPointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_object_lambda_access_point" {
				continue
			}

			accountID, name, err := tfs3control.ObjectLambdaAccessPointParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3control.FindObjectLambdaAccessPointConfigurationByTwoPartKey(ctx, conn, accountID, name)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Object Lambda Access Point %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckObjectLambdaAccessPointExists(ctx context.Context, n string, v *types.ObjectLambdaConfiguration) resource.TestCheckFunc {
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

		output, err := tfs3control.FindObjectLambdaAccessPointConfigurationByTwoPartKey(ctx, conn, accountID, name)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccObjectLambdaAccessPointBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLambdaBase(rName, rName, rName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs14.x"
}
`, rName))
}

func testAccObjectLambdaAccessPointConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccObjectLambdaAccessPointBaseConfig(rName), fmt.Sprintf(`
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
`, rName))
}

func testAccObjectLambdaAccessPointConfig_optionals(rName string) string {
	return acctest.ConfigCompose(testAccObjectLambdaAccessPointBaseConfig(rName), fmt.Sprintf(`
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
    allowed_features            = ["GetObject-Range", "GetObject-PartNumber"]
    cloud_watch_metrics_enabled = true
    supporting_access_point     = aws_s3_access_point.test.arn

    transformation_configuration {
      actions = ["GetObject"]

      content_transformation {
        aws_lambda {
          function_arn     = aws_lambda_function.test.arn
          function_payload = "{\"res-x\": \"100\",\"res-y\": \"100\"}"
        }
      }
    }
  }
}
`, rName))
}
