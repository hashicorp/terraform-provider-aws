# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bedrock_evaluation_job" "test" {
  region = var.region

  job_name = var.rName
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
      }

      human_workflow_config {
        flow_definition_arn = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:flow-definition/${var.rName}"
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

# testAccEvaluationJobConfig_base

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {
  region = var.region
}

resource "aws_s3_bucket" "test" {
  region = var.region

  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_object" "dataset" {
  region = var.region

  bucket  = aws_s3_bucket.test.id
  key     = "datasets/dataset.jsonl"
  content = <<-EOT
    {"prompt": "What is the capital of France?", "referenceResponse": "Paris", "category": "geography", "modelResponses": [{"response": "The capital of France is Paris.", "modelIdentifier": "my-model-v1"}]}
    {"prompt": "What is 2 + 2?", "referenceResponse": "4", "category": "math", "modelResponses": [{"response": "2 + 2 equals 4.", "modelIdentifier": "my-model-v1"}]}
  EOT
}

resource "aws_iam_role" "test" {
  name = var.rName

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
  name = var.rName
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

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
