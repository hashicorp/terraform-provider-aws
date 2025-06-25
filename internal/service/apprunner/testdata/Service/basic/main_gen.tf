# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_apprunner_service" "test" {
  service_name = var.rName
  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
