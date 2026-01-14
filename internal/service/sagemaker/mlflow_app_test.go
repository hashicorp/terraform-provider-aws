// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerMlflowApp_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var app sagemaker.DescribeMlflowAppOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_mlflow_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMlflowAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMlflowAppConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRoleARN, os.Getenv("SAGEMAKER_IAM_ROLE_ARN")),
					resource.TestCheckResourceAttr(resourceName, "artifact_store_uri", fmt.Sprintf("s3://%s/", os.Getenv("SAGEMAKER_S3_BUCKET"))),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`mlflow-app/app-.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccSageMakerMlflowApp_update(t *testing.T) {
	ctx := acctest.Context(t)
	var app sagemaker.DescribeMlflowAppOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_mlflow_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMlflowAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMlflowAppConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "artifact_store_uri", fmt.Sprintf("s3://%s/updated/", os.Getenv("SAGEMAKER_S3_BUCKET"))),
				),
			},
			{
				Config: testAccMlflowAppConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func TestAccSageMakerMlflowApp_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var app sagemaker.DescribeMlflowAppOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_mlflow_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMlflowAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMlflowAppConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMlflowAppConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccMlflowAppConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSageMakerMlflowApp_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var app sagemaker.DescribeMlflowAppOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_mlflow_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMlflowAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMlflowAppConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, resourceName, &app),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceMlflowApp(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMlflowAppDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_mlflow_app" {
				continue
			}

			output, err := tfsagemaker.FindMlflowAppByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("reading SageMaker Mlflow App (%s): %w", rs.Primary.ID, err)
			}

			// If status is Deleted, consider it destroyed
			if output.Status == "Deleted" {
				continue
			}

			return fmt.Errorf("SageMaker Mlflow App %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMlflowAppExists(ctx context.Context, n string, app *sagemaker.DescribeMlflowAppOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SageMaker Mlflow App ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerClient(ctx)
		resp, err := tfsagemaker.FindMlflowAppByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*app = *resp

		return nil
	}
}

func testAccMlflowAppConfig_base() string {
	return fmt.Sprintf(`
data "aws_s3_bucket" "test" {
  bucket = %[1]q
}

locals {
  role_arn = %[2]q
}
`,
		os.Getenv("SAGEMAKER_S3_BUCKET"),
		os.Getenv("SAGEMAKER_IAM_ROLE_ARN"))
}

func testAccMlflowAppConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccMlflowAppConfig_base(), fmt.Sprintf(`
resource "aws_sagemaker_mlflow_app" "test" {
  name               = %[1]q
  artifact_store_uri = "s3://${data.aws_s3_bucket.test.bucket}/"
  role_arn           = local.role_arn
}
`, rName))
}

func testAccMlflowAppConfig_update(rName string) string {
	return acctest.ConfigCompose(testAccMlflowAppConfig_base(), fmt.Sprintf(`
resource "aws_sagemaker_mlflow_app" "test" {
  name               = %[1]q
  artifact_store_uri = "s3://${data.aws_s3_bucket.test.bucket}/updated/"
  role_arn           = local.role_arn
}
`, rName))
}

func testAccMlflowAppConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccMlflowAppConfig_base(), fmt.Sprintf(`
resource "aws_sagemaker_mlflow_app" "test" {
  name               = %[1]q
  artifact_store_uri = "s3://${data.aws_s3_bucket.test.bucket}/"
  role_arn           = local.role_arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccMlflowAppConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccMlflowAppConfig_base(), fmt.Sprintf(`
resource "aws_sagemaker_mlflow_app" "test" {
  name               = %[1]q
  artifact_store_uri = "s3://${data.aws_s3_bucket.test.bucket}/"
  role_arn           = local.role_arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
