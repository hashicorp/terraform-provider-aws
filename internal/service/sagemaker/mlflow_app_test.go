// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerMlflowApp_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_mlflow_app.test"
	roleResourceName := "aws_iam_role.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMlflowAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMlflowAppConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "artifact_store_uri", fmt.Sprintf("s3://%s/", rName)),
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

func TestAccSageMakerMlflowApp_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_mlflow_app.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMlflowAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMlflowAppConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsagemaker.ResourceMlflowApp, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccSageMakerMlflowApp_update(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	artifactStoreURI := fmt.Sprintf("s3://%s/", rName)
	artifactStoreURIUpdated := fmt.Sprintf("s3://%s/updated/", rName)
	resourceName := "aws_sagemaker_mlflow_app.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMlflowAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMlflowAppConfig_artifactStoreURI(rName, artifactStoreURI),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "artifact_store_uri", artifactStoreURI),
				),
			},
			{
				Config: testAccMlflowAppConfig_artifactStoreURI(rName, artifactStoreURIUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "artifact_store_uri", artifactStoreURIUpdated),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccMlflowAppConfig_artifactStoreURI(rName, artifactStoreURI),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "artifact_store_uri", artifactStoreURI),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccSageMakerMlflowApp_tags(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_mlflow_app.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMlflowAppDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMlflowAppConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, t, resourceName),
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
					testAccCheckMlflowAppExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccMlflowAppConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMlflowAppExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
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
			_, err := tfsagemaker.FindMlflowAppByARN(ctx, conn, arn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("reading SageMaker Mlflow App (%s): %w", arn, err)
			}

			return fmt.Errorf("SageMaker Mlflow App %s still exists", arn)
		}

		return nil
	}
}

func testAccCheckMlflowAppExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)
		_, err := tfsagemaker.FindMlflowAppByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		return err
	}
}

func testAccMlflowAppConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_trust.json
}

data "aws_iam_policy_document" "test_trust" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

data "aws_iam_policy_document" "test_perms" {
  statement {
    effect = "Allow"

    actions = [
      "s3:Get*",
      "s3:Put*",
      "s3:List*",
    ]

    resources = [
      "arn:aws:s3:::${aws_s3_bucket.test.bucket}",
      "arn:aws:s3:::${aws_s3_bucket.test.bucket}/*"
    ]
  }

  statement {
    effect = "Allow"

    actions = [
      "sagemaker:AddTags",
      "sagemaker:CreateModelPackageGroup",
      "sagemaker:CreateModelPackage",
      "sagemaker:UpdateModelPackage",
      "sagemaker:DescribeModelPackageGroup",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_policy" "test" {
  name   = %[1]q
  policy = data.aws_iam_policy_document.test_perms.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, rName)
}

func testAccMlflowAppConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccMlflowAppConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_mlflow_app" "test" {
  name               = %[1]q
  artifact_store_uri = "s3://${aws_s3_bucket.test.bucket}/"
  role_arn           = aws_iam_role.test.arn

  depends_on = [ aws_iam_role_policy_attachment.test ]
}
`, rName))
}

func testAccMlflowAppConfig_artifactStoreURI(rName, artifactStoreURI string) string {
	return acctest.ConfigCompose(testAccMlflowAppConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_mlflow_app" "test" {
  name               = %[1]q
  artifact_store_uri = %[2]q
  role_arn           = aws_iam_role.test.arn
}
`, rName, artifactStoreURI))
}

func testAccMlflowAppConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccMlflowAppConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_mlflow_app" "test" {
  name               = %[1]q
  artifact_store_uri = "s3://${aws_s3_bucket.test.bucket}/"
  role_arn           = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccMlflowAppConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccMlflowAppConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_mlflow_app" "test" {
  name               = %[1]q
  artifact_store_uri = "s3://${aws_s3_bucket.test.bucket}/"
  role_arn           = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
