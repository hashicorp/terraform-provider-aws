# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_sagemaker_inference_component" "test" {
  count = var.resource_count

  name          = "${var.rName}-${count.index}"
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    }

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }

    startup_parameters {
      container_startup_health_check_timeout_in_seconds = 300
      model_data_download_timeout_in_seconds            = 300
    }
  }

  runtime_config {
    copy_count = 1
  }

  tags = var.resource_tags
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_iam_role_policy" "s3_access" {
  role = aws_iam_role.test.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket",
        ]
        Resource = [
          "arn:${data.aws_partition.current.partition}:s3:::${var.rName}",
          "arn:${data.aws_partition.current.partition}:s3:::${var.rName}/*",
        ]
      }
    ]
  })
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "sagemaker-tensorflow-serving"
  image_tag       = "1.12-cpu"
}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket       = aws_s3_bucket.test.bucket
  key          = "model/model.tar.gz"
  content_type = "application/gzip"
  source       = "test-fixtures/sagemaker-tensorflow-serving-test-model.tar.gz"
}

resource "aws_sagemaker_endpoint_configuration" "test" {
  name               = var.rName
  execution_role_arn = aws_iam_role.test.arn

  production_variants {
    variant_name           = "variant-1"
    instance_type          = "ml.c5.xlarge"
    initial_instance_count = 1

    managed_instance_scaling {
      status             = "ENABLED"
      min_instance_count = 1
      max_instance_count = 1
    }

    routing_config {
      routing_strategy = "LEAST_OUTSTANDING_REQUESTS"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.s3_access]
}

resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = var.rName

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.s3_access]
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
  description = "Tags to set on resource. To specify no tags, set to `null`"
  type        = map(string)
  nullable    = true
  default     = null
}
