# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_sagemaker_training_job" "test" {
  count = var.resource_count

  training_job_name = "${var.rName}-${count.index}"
  role_arn          = aws_iam_role.test.arn
  tags              = var.resource_tags

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity", "sts:TagSession"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "linear-learner"
  image_tag       = "1"
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
  type     = map(string)
  nullable = true
}
