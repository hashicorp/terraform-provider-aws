// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerMlflowTrackingServer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var mpg sagemaker.DescribeMlflowTrackingServerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_mlflow_tracking_server.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMlflowTrackingServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMlflowTrackingServerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowTrackingServerExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "tracking_server_name", rName),
					resource.TestCheckResourceAttr(resourceName, "automatic_model_registration", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "tracking_server_size", "Small"),
					resource.TestCheckResourceAttrSet(resourceName, "tracking_server_url"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("mlflow-tracking-server/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMlflowTrackingServerConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowTrackingServerExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "tracking_server_name", rName),
					resource.TestCheckResourceAttr(resourceName, "automatic_model_registration", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "tracking_server_size", "Medium"),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_window_start", "Sun:01:00"),
					resource.TestCheckResourceAttrSet(resourceName, "tracking_server_url"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("mlflow-tracking-server/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccSageMakerMlflowTrackingServer_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var mpg sagemaker.DescribeMlflowTrackingServerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_mlflow_tracking_server.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMlflowTrackingServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMlflowTrackingServerConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowTrackingServerExists(ctx, resourceName, &mpg),
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
				Config: testAccMlflowTrackingServerConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowTrackingServerExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccMlflowTrackingServerConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowTrackingServerExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSageMakerMlflowTrackingServer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var mpg sagemaker.DescribeMlflowTrackingServerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_mlflow_tracking_server.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMlflowTrackingServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMlflowTrackingServerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowTrackingServerExists(ctx, resourceName, &mpg),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceMlflowTrackingServer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMlflowTrackingServerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_mlflow_tracking_server" {
				continue
			}

			_, err := tfsagemaker.FindMlflowTrackingServerByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("reading SageMaker AI Mlflow Tracking Server (%s): %w", rs.Primary.ID, err)
			}

			return fmt.Errorf("sagemaker Mlflow Tracking Server %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMlflowTrackingServerExists(ctx context.Context, n string, mpg *sagemaker.DescribeMlflowTrackingServerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Mlflow Tracking Server ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerClient(ctx)
		resp, err := tfsagemaker.FindMlflowTrackingServerByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*mpg = *resp

		return nil
	}
}

func testAccMlflowTrackingServerConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test.json
  inline_policy {
    name   = "TrackingServerPolicy"
    policy = data.aws_iam_policy_document.tracking.json
  }
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

data "aws_iam_policy_document" "tracking" {
  statement {
    sid    = "Tracking"
    effect = "Allow"
    actions = [
      "s3:Get*",
      "s3:Put*",
      "s3:List*",
      "sagemaker:AddTags",
      "sagemaker:CreateModelPackageGroup",
      "sagemaker:CreateModelPackage",
      "sagemaker:UpdateModelPackage",
      "sagemaker:DescribeModelPackageGroup"
    ]
    resources = ["*"]
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, rName)
}

func testAccMlflowTrackingServerConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccMlflowTrackingServerConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_mlflow_tracking_server" "test" {
  tracking_server_name = %[1]q
  role_arn             = aws_iam_role.test.arn
  artifact_store_uri   = "s3://${aws_s3_bucket.test.bucket}/path"
}
`, rName))
}

func testAccMlflowTrackingServerConfig_update(rName string) string {
	return acctest.ConfigCompose(testAccMlflowTrackingServerConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_mlflow_tracking_server" "test" {
  tracking_server_name            = %[1]q
  role_arn                        = aws_iam_role.test.arn
  artifact_store_uri              = "s3://${aws_s3_bucket.test.bucket}/path"
  automatic_model_registration    = true
  tracking_server_size            = "Medium"
  weekly_maintenance_window_start = "Sun:01:00"
}
`, rName))
}

func testAccMlflowTrackingServerConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccMlflowTrackingServerConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_mlflow_tracking_server" "test" {
  tracking_server_name = %[1]q
  role_arn             = aws_iam_role.test.arn
  artifact_store_uri   = "s3://${aws_s3_bucket.test.bucket}/path"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccMlflowTrackingServerConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccMlflowTrackingServerConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_mlflow_tracking_server" "test" {
  tracking_server_name = %[1]q
  role_arn             = aws_iam_role.test.arn
  artifact_store_uri   = "s3://${aws_s3_bucket.test.bucket}/path"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
