// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerMlflowApp_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var app sagemaker.DescribeMlflowAppOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_mlflow_app.test"
	s3Bucket := os.Getenv("SAGEMAKER_S3_BUCKET")
	roleArn := os.Getenv("SAGEMAKER_IAM_ROLE_ARN")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.SkipIfEnvVarNotSet(t, "SAGEMAKER_IAM_ROLE_ARN")
			acctest.SkipIfEnvVarNotSet(t, "SAGEMAKER_S3_BUCKET")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMlflowAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMlflowAppConfig_basic(rName, s3Bucket, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, t, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRoleARN, roleArn),
					resource.TestCheckResourceAttr(resourceName, "artifact_store_uri", fmt.Sprintf("s3://%s/", s3Bucket)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`mlflow-app/app-.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccSageMakerMlflowApp_update(t *testing.T) {
	ctx := acctest.Context(t)
	var app sagemaker.DescribeMlflowAppOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_mlflow_app.test"
	s3Bucket := os.Getenv("SAGEMAKER_S3_BUCKET")
	roleArn := os.Getenv("SAGEMAKER_IAM_ROLE_ARN")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.SkipIfEnvVarNotSet(t, "SAGEMAKER_IAM_ROLE_ARN")
			acctest.SkipIfEnvVarNotSet(t, "SAGEMAKER_S3_BUCKET")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMlflowAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMlflowAppConfig_update(rName, s3Bucket, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, t, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "artifact_store_uri", fmt.Sprintf("s3://%s/updated/", s3Bucket)),
				),
			},
			{
				Config: testAccMlflowAppConfig_basic(rName, s3Bucket, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, t, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func TestAccSageMakerMlflowApp_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var app sagemaker.DescribeMlflowAppOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_mlflow_app.test"
	s3Bucket := os.Getenv("SAGEMAKER_S3_BUCKET")
	roleArn := os.Getenv("SAGEMAKER_IAM_ROLE_ARN")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.SkipIfEnvVarNotSet(t, "SAGEMAKER_IAM_ROLE_ARN")
			acctest.SkipIfEnvVarNotSet(t, "SAGEMAKER_S3_BUCKET")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMlflowAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMlflowAppConfig_tags1(rName, s3Bucket, roleArn, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, t, resourceName, &app),
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
				Config: testAccMlflowAppConfig_tags2(rName, s3Bucket, roleArn, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, t, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccMlflowAppConfig_tags1(rName, s3Bucket, roleArn, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, t, resourceName, &app),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_mlflow_app.test"
	s3Bucket := os.Getenv("SAGEMAKER_S3_BUCKET")
	roleArn := os.Getenv("SAGEMAKER_IAM_ROLE_ARN")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.SkipIfEnvVarNotSet(t, "SAGEMAKER_IAM_ROLE_ARN")
			acctest.SkipIfEnvVarNotSet(t, "SAGEMAKER_S3_BUCKET")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMlflowAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMlflowAppConfig_basic(rName, s3Bucket, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, t, resourceName, &app),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsagemaker.ResourceMlflowApp, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMlflowAppDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_mlflow_app" {
				continue
			}

			arn := rs.Primary.Attributes[names.AttrARN]
			output, err := tfsagemaker.FindMlflowAppByARN(ctx, conn, arn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("reading SageMaker Mlflow App (%s): %w", arn, err)
			}

			if output.Status == awstypes.MlflowAppStatusDeleted {
				continue
			}

			return fmt.Errorf("SageMaker Mlflow App %s still exists", arn)
		}

		return nil
	}
}

func testAccCheckMlflowAppExists(ctx context.Context, t *testing.T, n string, app *sagemaker.DescribeMlflowAppOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)
		resp, err := tfsagemaker.FindMlflowAppByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*app = *resp

		return nil
	}
}

func testAccMlflowAppConfig_base(s3Bucket, roleArn string) string {
	return fmt.Sprintf(`
data "aws_s3_bucket" "test" {
  bucket = %[1]q
}

locals {
  role_arn = %[2]q
}
`, s3Bucket, roleArn)
}

func testAccMlflowAppConfig_basic(rName, s3Bucket, roleArn string) string {
	return acctest.ConfigCompose(testAccMlflowAppConfig_base(s3Bucket, roleArn), fmt.Sprintf(`
resource "aws_sagemaker_mlflow_app" "test" {
  name               = %[1]q
  artifact_store_uri = "s3://${data.aws_s3_bucket.test.bucket}/"
  role_arn           = local.role_arn
}
`, rName))
}

func testAccMlflowAppConfig_update(rName, s3Bucket, roleArn string) string {
	return acctest.ConfigCompose(testAccMlflowAppConfig_base(s3Bucket, roleArn), fmt.Sprintf(`
resource "aws_sagemaker_mlflow_app" "test" {
  name               = %[1]q
  artifact_store_uri = "s3://${data.aws_s3_bucket.test.bucket}/updated/"
  role_arn           = local.role_arn
}
`, rName))
}

func testAccMlflowAppConfig_tags1(rName, s3Bucket, roleArn, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccMlflowAppConfig_base(s3Bucket, roleArn), fmt.Sprintf(`
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

func testAccMlflowAppConfig_tags2(rName, s3Bucket, roleArn, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccMlflowAppConfig_base(s3Bucket, roleArn), fmt.Sprintf(`
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
