// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/bedrock/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
)

func TestAccBedrockEvaluationJob_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	//var evaluationjob bedrock.DescribeEvaluationJobResponse
	rName := name_regex(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))
	resourceName := "aws_bedrock_evaluation_job.test"
	bName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	//bucketName := "aws_s3_bucket.test"
	modelName := "aws_bedrock_foundation_model.test"
	iamName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEvaluationJobConfig_basic(rName, iamName, bName),
				Check: resource.ComposeTestCheckFunc(
					//testAccCheckEvaluationJobExists(ctx, resourceName, &evaluationjob),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					//resource.TestCheckResourceAttrPair(resourceName, "customer_cencryption_key_id", ""),
					// eval config
					resource.TestCheckResourceAttr(resourceName, "evaluation_job.automated_evaluation_config.metric_names.0", "Builtin.Accuracy"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_job.automated_evaluation_config.task_type.0", "Summarization"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_job.automated_evaluation_config.data_set.data_set.name", "Builtin.Bold"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_job.automated_evaluation_config.data_set.name", "Builtin.Bold"),
					// eval config end
					// inf config
					resource.TestCheckResourceAttrPair(resourceName, "evaluation_job_inference_config.models.bedrock_model.inference_params", modelName, "inference_types_supported.0"),
					resource.TestCheckResourceAttrPair(resourceName, "evaluation_job_inference_config.models.bedrock_model.id", modelName, names.AttrID),
					// inf config end
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "Summarization"),
					resource.TestCheckResourceAttrSet(resourceName, "output_data_config[0].s3_uri"),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"), // not sure what to do here
					//resource.TestCheckResourceAttrSet(resourceName, "failure_messages"), // not sure if this should be set
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					// resource.TestCheckResourceAttrPairSet(resourceName, "last_modified_time") don't think I need this
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func name_regex(in string) string {
	return in
	var out string
	for index, char := range in {
		if index == 1 {
			out += "-"
		} else if index == 2 {
			out += "*"
		} else {
			out += string(char)
		}
	}
	for len(in) < 61 {
		in += "e"
	}
	return out
}

/*
func TestAccBedrockEvaluationJob_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var evaluationjob bedrock.DescribeEvaluationJobResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_evaluation_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEvaluationJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEvaluationJobConfig_basic(rName, testAccEvaluationJobVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEvaluationJobExists(ctx, resourceName, &evaluationjob),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceEvaluationJob = newResourceEvaluationJob
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrock.ResourceEvaluationJob, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
*/
/*
func testAccCheckEvaluationJobExists(ctx context.Context, name string, evaluationjob *bedrock.DescribeEvaluationJobResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameEvaluationJob, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameEvaluationJob, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)
		resp, err := conn.DescribeEvaluationJob(ctx, &bedrock.DescribeEvaluationJobInput{
			EvaluationJobId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameEvaluationJob, rs.Primary.ID, err)
		}

		*evaluationjob = *resp

		return nil
	}
}
*/

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)

	input := &bedrock.ListEvaluationJobsInput{}
	_, err := conn.ListEvaluationJobs(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccEvaluationJobConfig_base(iamName, bucketName string) string {
	return fmt.Sprintf(`
data "aws_bedrock_foundation_models" "test" {}

data "aws_bedrock_foundation_model" "test" {
  model_id = data.aws_bedrock_foundation_models.test.model_summaries[0].model_id
}

resource "aws_s3_bucket" "test" {
  bucket = %[2]q
  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT", "POST", "DELETE"]
    allowed_origins = ["*"]
    expose_headers  = ["Access-Control-Allow-Origin"]
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
  {
      "Version": "2012-10-17",
    "Statement": [{
        "Sid": "AllowBedrockToAssumeRole",
        "Effect": "Allow",
        "Principal": {
            "Service": "bedrock.amazonaws.com"
        },
        "Action": "sts:AssumeRole",
        "Condition": {
            "StringEquals": {
                "aws:SourceArn": "111122223333"
            },
            "ArnEquals": {
                "aws:SourceArn": "arn:aws:bedrock:AWS Region:111122223333:evaluation-job/*"
            }
        }
    }]
  }
EOF
}
resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id


  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "BedrockConsole",
        "Effect" : "Allow",
        "Action" : [
          "bedrock:CreateEvaluationJob",
          "bedrock:GetEvaluationJob",
          "bedrock:ListEvaluationJobs",
          "bedrock:StopEvaluationJob",
          "bedrock:GetCustomModel",
          "bedrock:ListCustomModels",
          "bedrock:CreateProvisionedModelThroughput",
          "bedrock:UpdateProvisionedModelThroughput",
          "bedrock:GetProvisionedModelThroughput",
          "bedrock:ListProvisionedModelThroughputs",
          "bedrock:ListTagsForResource",
          "bedrock:UntagResource",
          "bedrock:TagResource"
        ],
        "Resource" : "*"
      },
      {
        "Sid" : "AllowConsoleS3AccessForModelEvaluation",
        "Effect" : "Allow",
        "Action" : [
          "s3:GetObject",
          "s3:GetBucketCORS",
          "s3:ListBucket",
          "s3:ListBucketVersions",
          "s3:GetBucketLocation"
        ],
        "Resource" : "*"
      }
    ]
  })
}
	`, iamName, bucketName)
}

func testAccEvaluationJobConfig_basic(iamName, jobName, bucketName string) string {
	return acctest.ConfigCompose(testAccEvaluationJobConfig_base(iamName, bucketName), fmt.Sprintf(`


resource "aws_bedrock_evaluation_job" "test" {

  evaluation_config {
    automated {
      automated_evaluation_config {
        dataset_metric_configs {
          data_set {
            name = "Builtin.Bold"
          }
          metric_names = ["Builtin.Accuracy"]
          task_type    = "Summarization"
        }
      }
    }
  }

  inference_config {
    models {
      bedrock_model {
        inference_params = tolist(data.aws_bedrock_foundation_model.test.inference_types_supported)[0]
        model_identifier = data.aws_bedrock_foundation_model.test.id
      }
    }
  }

  description = "test"
  name        = %[1]q

  output_data_config {
    s3_uri = "s3://${aws_s3_bucket.test.bucket}/bedrock/evaluation_jobs}"
  }

  role_arn = aws_iam_role.test.arn
}
`, jobName))
}
