// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Note: These acceptance tests use human-based evaluation jobs with precomputed
// inference response data. Unlike automated evaluation jobs, this combination
// doesn't require Amazon Bedrock to invoke any foundation model on the caller's
// behalf, so it can be exercised in accounts that restrict model invocation.
// Automated evaluation jobs and knowledge base (RAG) evaluation jobs are
// supported by this resource but not covered by these acceptance tests.
//
// Additionally, destroying the resource stops the evaluation job, but Amazon
// Bedrock does not support permanently deleting it. The resource will still
// appear in a "Stopped" state after the Terraform resource is destroyed.

func TestAccBedrockEvaluationJob_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_evaluation_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckEvaluationJob(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEvaluationJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEvaluationJobConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEvaluationJobExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "job_arn"),
					resource.TestCheckResourceAttr(resourceName, "job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "job_type", "Human"),
					resource.TestCheckResourceAttr(resourceName, "application_type", "ModelEvaluation"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, "evaluation_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_config.0.human.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_config.0.human.0.dataset_metric_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_config.0.human.0.dataset_metric_config.0.task_type", "Generation"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_config.0.human.0.dataset_metric_config.0.dataset.0.name", "custom-dataset"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_config.0.human.0.dataset_metric_config.0.metric_names.0", "overall"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_config.0.human.0.custom_metric.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_config.0.human.0.custom_metric.0.name", "overall"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_config.0.human.0.custom_metric.0.rating_method", "ThumbsUpDown"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_config.0.human.0.human_workflow_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inference_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inference_config.0.model.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inference_config.0.model.0.precomputed_inference_source.0.inference_source_identifier", "my-model-v1"),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "job_arn", "bedrock", regexache.MustCompile(`evaluation-job/.+`)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "job_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "job_arn",
			},
		},
	})
}

func TestAccBedrockEvaluationJob_optional(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_evaluation_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckEvaluationJob(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEvaluationJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEvaluationJobConfig_optional(rName, rName, "test description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEvaluationJobExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "job_description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "application_type", "ModelEvaluation"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "job_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "job_arn",
			},
			{
				Config: testAccEvaluationJobConfig_optional(rName, rName+"-2", "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEvaluationJobExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "job_name", rName+"-2"),
					resource.TestCheckResourceAttr(resourceName, "job_description", "updated description"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func TestAccBedrockEvaluationJob_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	acctest.SkipIfEnvVarNotSet(t, "TF_AWS_ALLOW_SKIP_DESTROY")

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_evaluation_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckEvaluationJob(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEvaluationJobNotStopped(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEvaluationJobConfig_skipDestroy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEvaluationJobExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "InProgress"),
				),
			},
		},
	})
}

func testAccCheckEvaluationJobExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return errors.New("resource not found: " + n)
		}

		arn := rs.Primary.Attributes["job_arn"]
		if arn == "" {
			return errors.New("job_arn not set")
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		_, err := tfbedrock.FindEvaluationJobByARN(ctx, conn, arn)
		return err
	}
}

func testAccCheckEvaluationJobDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_evaluation_job" {
				continue
			}

			arn := rs.Primary.Attributes["job_arn"]
			output, err := tfbedrock.FindEvaluationJobByARN(ctx, conn, arn)
			if err != nil {
				return err
			}

			if got, want := string(output.Status), "Stopped"; got != want {
				return fmt.Errorf("Bedrock Evaluation Job %s status = %s, want %s", arn, got, want)
			}
		}

		return nil
	}
}

func testAccCheckEvaluationJobNotStopped(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_evaluation_job" {
				continue
			}

			output, err := tfbedrock.FindEvaluationJobByARN(ctx, conn, rs.Primary.Attributes["job_arn"])
			if err != nil {
				return err
			}

			if got, want := string(output.Status), "InProgress"; got != want {
				return fmt.Errorf("Bedrock Evaluation Job %s status = %s, want %s", rs.Primary.Attributes["job_arn"], got, want)
			}
		}

		return nil
	}
}

func testAccPreCheckEvaluationJob(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

	input := bedrock.ListEvaluationJobsInput{}

	_, err := conn.ListEvaluationJobs(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccEvaluationJobConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "dataset" {
  bucket  = aws_s3_bucket.test.id
  key     = "datasets/dataset.jsonl"
  content = <<-EOT
    {"prompt": "What is the capital of France?", "referenceResponse": "Paris", "category": "geography", "modelResponses": [{"response": "The capital of France is Paris.", "modelIdentifier": "my-model-v1"}]}
    {"prompt": "What is 2 + 2?", "referenceResponse": "4", "category": "math", "modelResponses": [{"response": "2 + 2 equals 4.", "modelIdentifier": "my-model-v1"}]}
  EOT
}

resource "aws_iam_role" "test" {
  name = %[1]q

  # See https://docs.aws.amazon.com/bedrock/latest/userguide/model-evaluation-security.html.
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "bedrock.amazonaws.com"
      }
      Action = "sts:AssumeRole"
      Condition = {
        StringEquals = {
          "aws:SourceAccount" = data.aws_caller_identity.current.account_id
        }
        ArnEquals = {
          "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:evaluation-job/*"
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowAccessToDatasetAndOutputBucket"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket",
          "s3:PutObject",
          "s3:GetBucketLocation",
          "s3:AbortMultipartUpload",
          "s3:ListBucketMultipartUploads",
        ]
        Resource = [
          aws_s3_bucket.test.arn,
          "${aws_s3_bucket.test.arn}/*",
        ]
      }
    ]
  })
}
`, rName)
}

func testAccEvaluationJobConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEvaluationJobConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrock_evaluation_job" "test" {
  job_name = %[1]q
  role_arn = aws_iam_role.test.arn

  evaluation_config {
    human {
      dataset_metric_config {
        task_type = "Generation"

        dataset {
          name = "custom-dataset"

          dataset_location {
            s3_uri = "s3://${aws_s3_bucket.test.id}/${aws_s3_object.dataset.key}"
          }
        }

        metric_names = ["overall"]
      }

      custom_metric {
        name          = "overall"
        rating_method = "ThumbsUpDown"
        description   = "Overall quality of the response."
      }

      human_workflow_config {
        flow_definition_arn = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:flow-definition/%[1]s"
        instructions        = "Rate the response."
      }
    }
  }

  inference_config {
    model {
      precomputed_inference_source {
        inference_source_identifier = "my-model-v1"
      }
    }
  }

  output_data_config {
    s3_uri = "s3://${aws_s3_bucket.test.id}/output/"
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.dataset]
}
`, rName))
}

func testAccEvaluationJobConfig_skipDestroy(rName string) string {
	return acctest.ConfigCompose(testAccEvaluationJobConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrock_evaluation_job" "test" {
  job_name     = %[1]q
  role_arn     = aws_iam_role.test.arn
  skip_destroy = true

  evaluation_config {
    human {
      dataset_metric_config {
        task_type = "Generation"

        dataset {
          name = "custom-dataset"

          dataset_location {
            s3_uri = "s3://${aws_s3_bucket.test.id}/${aws_s3_object.dataset.key}"
          }
        }

        metric_names = ["overall"]
      }

      custom_metric {
        name          = "overall"
        rating_method = "ThumbsUpDown"
      }

      human_workflow_config {
        flow_definition_arn = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:flow-definition/%[1]s"
      }
    }
  }

  inference_config {
    model {
      precomputed_inference_source {
        inference_source_identifier = "my-model-v1"
      }
    }
  }

  output_data_config {
    s3_uri = "s3://${aws_s3_bucket.test.id}/output/"
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.dataset]
}
`, rName))
}

func testAccEvaluationJobConfig_optional(rName, jobName, description string) string {
	return acctest.ConfigCompose(testAccEvaluationJobConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrock_evaluation_job" "test" {
  job_name         = %[1]q
  job_description  = %[2]q
  role_arn         = aws_iam_role.test.arn
  application_type = "ModelEvaluation"

  evaluation_config {
    human {
      dataset_metric_config {
        task_type = "Generation"

        dataset {
          name = "custom-dataset"

          dataset_location {
            s3_uri = "s3://${aws_s3_bucket.test.id}/${aws_s3_object.dataset.key}"
          }
        }

        metric_names = ["overall"]
      }

      custom_metric {
        name          = "overall"
        rating_method = "ThumbsUpDown"
      }

      human_workflow_config {
        flow_definition_arn = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:flow-definition/%[1]s"
      }
    }
  }

  inference_config {
    model {
      precomputed_inference_source {
        inference_source_identifier = "my-model-v1"
      }
    }
  }

  output_data_config {
    s3_uri = "s3://${aws_s3_bucket.test.id}/output/"
  }

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.dataset]
}
`, jobName, description))
}
