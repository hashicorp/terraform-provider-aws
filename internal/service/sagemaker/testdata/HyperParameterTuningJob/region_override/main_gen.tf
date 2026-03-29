# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
  region = var.region

  hyper_parameter_tuning_job_name = var.rName

  hyper_parameter_tuning_job_config {
    strategy = "Bayesian"

    hyper_parameter_tuning_job_objective {
      metric_name = "test:msd"
      type        = "Minimize"
    }

    parameter_ranges {
      integer_parameter_ranges {
        max_value    = "2"
        min_value    = "1"
        name         = "epochs"
      }
    }

    resource_limits {
      max_number_of_training_jobs = 2
      max_parallel_training_jobs  = 1
    }
  }

  training_job_definition {
    role_arn = aws_iam_role.test.arn

    algorithm_specification {
      training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      training_input_mode = "File"
    }

    input_data_config {
      channel_name = "train"

      data_source {
        s3_data_source {
          s3_data_type = "S3Prefix"
          s3_uri       = "s3://${aws_s3_bucket.test.bucket}/input/"
        }
      }
    }

    output_data_config {
      s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
    }

    resource_config {
      instance_count     = 1
      instance_type      = "ml.m5.large"
      volume_size_in_gb  = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}

data "aws_partition" "current" {}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  region = var.region

  repository_name = "kmeans"
}

data "aws_iam_policy_document" "test" {
  statement {
    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }

    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject",
    ]

    resources = [
      "${aws_s3_bucket.test.arn}/*",
    ]
  }

  statement {
    actions = [
      "s3:ListBucket",
    ]

    resources = [
      aws_s3_bucket.test.arn,
    ]
  }

}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.s3.json
}

resource "aws_s3_bucket" "test" {
  region = var.region

  bucket        = "${var.rName}-hptj"
  force_destroy = true
}

resource "aws_s3_object" "input" {
  region = var.region

  bucket  = aws_s3_bucket.test.id
  key     = "input/placeholder.csv"
  content = "feature1,label\n1.0,0\n"
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
