# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
  count = var.resource_count

  name = "${substr(var.rName, 0, 20)}-${count.index}"

  config {
    strategy = "Bayesian"

    objective {
      metric_name = "validation:accuracy"
      type        = "Maximize"
    }

    parameter_ranges {
      continuous_parameter_ranges {
        max_value = "0.5"
        min_value = "0.1"
        name      = "learning_rate"
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
      algorithm_name      = aws_sagemaker_algorithm.test.algorithm_name
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
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }

  tags = var.resource_tags

  depends_on = [aws_iam_role_policy.test, aws_s3_object.input]
}

resource "aws_sagemaker_algorithm" "test" {
  algorithm_name = "${var.rName}-algorithm"

  training_specification {
    training_image                    = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    supported_training_instance_types = ["ml.m5.large"]

    metric_definitions {
      name  = "validation:accuracy"
      regex = "validation:accuracy=(.*?);"
    }

    supported_hyper_parameters {
      default_value = "0.2"
      description   = "Learning rate"
      is_required   = false
      is_tunable    = true
      name          = "learning_rate"
      type          = "Continuous"

      range {
        continuous_parameter_range_specification {
          min_value = "0.1"
          max_value = "0.5"
        }
      }
    }

    supported_tuning_job_objective_metrics {
      metric_name = "validation:accuracy"
      type        = "Maximize"
    }

    training_channels {
      name                    = "train"
      supported_content_types = ["text/csv"]
      supported_input_modes   = ["File"]
    }
  }
}

data "aws_partition" "current" {}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
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

  statement {
    actions = [
      "sagemaker:DescribeAlgorithm",
    ]

    resources = [
      "*",
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
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_object" "input" {
  bucket  = aws_s3_bucket.test.id
  key     = "input/placeholder.csv"
  content = "feature1,label\n1.0,0\n"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource"
  type        = map(string)
  nullable    = false
}

