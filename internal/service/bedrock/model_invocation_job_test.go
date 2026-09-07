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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockModelInvocationJob_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_model_invocation_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckModelInvocationJob(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelInvocationJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccModelInvocationJobConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckModelInvocationJobExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "job_arn"),
					resource.TestCheckResourceAttr(resourceName, "job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "model_id", "us.amazon.nova-2-lite-v1:0"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "submit_time"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.s3_input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.0.s3_output_data_config.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "job_arn", "bedrock", regexache.MustCompile(`model-invocation-job/.+`)),
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
				ImportStateVerifyIgnore: []string{
					names.AttrStatus,
				},
			},
		},
	})
}

func TestAccBedrockModelInvocationJob_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	acctest.SkipIfEnvVarNotSet(t, "TF_AWS_ALLOW_SKIP_DESTROY")

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_model_invocation_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckModelInvocationJob(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelInvocationJobNotStopped(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccModelInvocationJobConfig_skipDestroy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckModelInvocationJobExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Submitted"),
				),
			},
		},
	})
}

// testAccCheckModelInvocationJobExists verifies the job is reachable via the AWS API.
func testAccCheckModelInvocationJobExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
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

		_, err := tfbedrock.FindModelInvocationJobByARN(ctx, conn, arn)
		return err
	}
}

// AWS does not support permanent deletion of model invocation jobs.
func testAccCheckModelInvocationJobDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_model_invocation_job" {
				continue
			}

			arn := rs.Primary.Attributes["job_arn"]
			output, err := tfbedrock.FindModelInvocationJobByARN(ctx, conn, arn)

			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}

			if got, want := string(output.Status), "Stopped"; got != want {
				return fmt.Errorf("Bedrock Model Invocation Job %s status = %s, want %s", arn, got, want)
			}
		}

		return nil
	}
}

func testAccCheckModelInvocationJobNotStopped(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_model_invocation_job" {
				continue
			}

			output, err := tfbedrock.FindModelInvocationJobByARN(ctx, conn, rs.Primary.Attributes["job_arn"])
			if err != nil {
				return err
			}

			// skip_destroy = true means destroy is a no-op, so Stop is never
			// called. The job may legitimately be in any non-stopped status
			// (Submitted, Validating, Scheduled, InProgress, Stopping, or a
			// terminal status it reached on its own) by the time this runs;
			// only "Stopped" would indicate our provider incorrectly called
			// StopModelInvocationJob.
			if got, unwanted := string(output.Status), "Stopped"; got == unwanted {
				return fmt.Errorf("Bedrock Model Invocation Job %s got status = %s when skip_destroy is true", rs.Primary.Attributes["job_arn"], got)
			}
		}

		return nil
	}
}

func testAccPreCheckModelInvocationJob(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

	input := bedrock.ListModelInvocationJobsInput{}

	_, err := conn.ListModelInvocationJobs(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccModelInvocationJobConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "input" {
  bucket  = aws_s3_bucket.test.id
  key     = "input/records.jsonl"
  content = <<-EOT
    {"recordId": "1", "modelInput": {"schemaVersion": "messages-v1", "messages": [{"role": "user", "content": [{"text": "Hello, world!"}]}]}}
  EOT
}

resource "aws_iam_role" "test" {
  name = %[1]q

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
        ArnLike = {
          "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:model-invocation-job/*"
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
        Sid    = "S3Access"
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket",
          "s3:PutObject",
          "s3:GetBucketLocation",
        ]
        Resource = [
          aws_s3_bucket.test.arn,
          "${aws_s3_bucket.test.arn}/*",
        ]
      },
      {
        Sid    = "CrossRegionInference"
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel",
        ]
        Resource = [
          "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:inference-profile/us.amazon.nova-2-lite-v1:0",
          "arn:${data.aws_partition.current.partition}:bedrock:*::foundation-model/amazon.nova-2-lite-v1:0",
        ]
      }
    ]
  })
}
`, rName)
}

func testAccModelInvocationJobConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccModelInvocationJobConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrock_model_invocation_job" "test" {
  job_name = %[1]q
  model_id = "us.amazon.nova-2-lite-v1:0"
  role_arn = aws_iam_role.test.arn

  input_data_config {
    s3_input_data_config {
      s3_uri = "s3://${aws_s3_bucket.test.id}/input/"
    }
  }

  output_data_config {
    s3_output_data_config {
      s3_uri = "s3://${aws_s3_bucket.test.id}/output/"
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}
`, rName))
}

func testAccModelInvocationJobConfig_tags(rName, tagVal1, tagVal2 string) string {
	return acctest.ConfigCompose(testAccModelInvocationJobConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrock_model_invocation_job" "test" {
  job_name = %[1]q
  model_id = "us.amazon.nova-2-lite-v1:0"
  role_arn = aws_iam_role.test.arn

  input_data_config {
    s3_input_data_config {
      s3_uri = "s3://${aws_s3_bucket.test.id}/input/"
    }
  }

  output_data_config {
    s3_output_data_config {
      s3_uri = "s3://${aws_s3_bucket.test.id}/output/"
    }
  }

  tags = {
    key1 = %[2]q
    key2 = %[3]q
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}
`, rName, tagVal1, tagVal2))
}

func testAccModelInvocationJobConfig_skipDestroy(rName string) string {
	return acctest.ConfigCompose(testAccModelInvocationJobConfig_base(rName), fmt.Sprintf(`
resource "aws_bedrock_model_invocation_job" "test" {
  job_name     = %[1]q
  model_id     = "us.amazon.nova-2-lite-v1:0"
  role_arn     = aws_iam_role.test.arn
  skip_destroy = true

  input_data_config {
    s3_input_data_config {
      s3_uri = "s3://${aws_s3_bucket.test.id}/input/"
    }
  }

  output_data_config {
    s3_output_data_config {
      s3_uri = "s3://${aws_s3_bucket.test.id}/output/"
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}
`, rName))
}
