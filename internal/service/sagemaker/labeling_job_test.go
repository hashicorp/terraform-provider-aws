// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccLabelingJob_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_labeling_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLabelingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLabelingJobConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLabelingJobExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("labeling_job_arn"), tfknownvalue.RegionalARNExact("sagemaker", fmt.Sprintf("labeling-job/%s", rName))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("labeling_job_status"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "labeling_job_name"),
				ImportStateVerifyIdentifierAttribute: "labeling_job_name",
			},
		},
	})
}

func testAccLabelingJob_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_labeling_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLabelingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLabelingJobConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLabelingJobExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsagemaker.ResourceLabelingJob, resourceName),
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

func testAccLabelingJob_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_labeling_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLabelingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLabelingJobConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLabelingJobExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "labeling_job_name"),
				ImportStateVerifyIdentifierAttribute: "labeling_job_name",
			},
		},
	})
}

func testAccCheckLabelingJobDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_labeling_job" {
				continue
			}

			_, err := tfsagemaker.FindLabelingJobByName(ctx, conn, rs.Primary.Attributes["labeling_job_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker AI Labeling Job (%s) still exists", rs.Primary.Attributes["labeling_job_name"])
		}
		return nil
	}
}

func testAccCheckLabelingJobExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		_, err := tfsagemaker.FindLabelingJobByName(ctx, conn, rs.Primary.Attributes["labeling_job_name"])

		return err
	}
}

func testAccLabelingJobConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccWorkteamConfig_oidc(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "categories" {
  bucket = aws_s3_bucket.test.bucket
  key    = "label-categories.json"
  source = "test-fixtures/label-categories.json"
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}

data "aws_iam_policy_document" "test_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "s3:*",
      "sns:*",
      "sqs:*"
    ]
    effect    = "Allow"
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}

data "aws_partition" "current" {}
data "aws_region" "current" {}
`, rName))
}

func testAccLabelingJobConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLabelingJobConfig_base(rName), fmt.Sprintf(`
# https://docs.aws.amazon.com/sagemaker/latest/dg/sms-named-entity-recg.html#sms-creating-ner-api.
resource "aws_sagemaker_labeling_job" "test" {
  label_attribute_name = "testlabel"
  labeling_job_name    = %[1]q
  role_arn             = aws_iam_role.test.arn

  label_category_config_s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.categories.key}"

  human_task_config {
    number_of_human_workers_per_data_object = 1
    task_description                        = "Apply the labels provided to specific words or phrases within the larger text block."
    task_title                              = "Named entity Recognition task"
    task_time_limit_in_seconds              = 28800
    workteam_arn                            = aws_sagemaker_workteam.test.arn

    ui_config {
      human_task_ui_arn = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.region}:394669845002:human-task-ui/NamedEntityRecognition"
    }

    pre_human_task_lambda_arn = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.region}:081040173940:function:PRE-NamedEntityRecognition"

    annotation_consolidation_config {
      annotation_consolidation_lambda_arn = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.region}:081040173940:function:ACS-NamedEntityRecognition"
    }
  }

  input_config {
    data_source {
      sns_data_source {
        sns_topic_arn = aws_sns_topic.test.arn
      }
    }
  }

  output_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccLabelingJobConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccLabelingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_labeling_job" "test" {
  label_attribute_name = "testlabel"
  labeling_job_name    = %[1]q
  role_arn             = aws_iam_role.test.arn

  label_category_config_s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.categories.key}"

  human_task_config {
    number_of_human_workers_per_data_object = 1
    task_description                        = "Apply the labels provided to specific words or phrases within the larger text block."
    task_title                              = "Named entity Recognition task"
    task_time_limit_in_seconds              = 28800
    workteam_arn                            = aws_sagemaker_workteam.test.arn

    ui_config {
      human_task_ui_arn = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.region}:394669845002:human-task-ui/NamedEntityRecognition"
    }

    pre_human_task_lambda_arn = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.region}:081040173940:function:PRE-NamedEntityRecognition"

    annotation_consolidation_config {
      annotation_consolidation_lambda_arn = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.region}:081040173940:function:ACS-NamedEntityRecognition"
    }
  }

  input_config {
    data_source {
      sns_data_source {
        sns_topic_arn = aws_sns_topic.test.arn
      }
    }
  }

  output_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, tagKey1, tagValue1))
}
