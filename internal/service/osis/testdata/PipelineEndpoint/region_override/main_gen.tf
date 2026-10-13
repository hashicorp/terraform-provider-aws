# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_osis_pipeline_endpoint" "test" {
  region = var.region

  pipeline_arn = aws_osis_pipeline.pipeline.pipeline_arn

  vpc_options {
    subnet_ids         = [aws_subnet.test.id]
    security_group_ids = [aws_security_group.test.id]
  }
}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block           = "10.1.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_subnet" "test" {
  region = var.region

  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id
}

resource "aws_security_group" "test" {
  region = var.region

  name   = "${var.rName}-endpoint"
  vpc_id = aws_vpc.test.id
}

# testAccPipelineEndpointConfig_pipeline

resource "aws_osis_pipeline" "pipeline" {
  region = var.region

  pipeline_name               = var.rName
  pipeline_configuration_body = <<-EOT
            version: "2"
            test-pipeline:
              source:
                http:
                  path: "/test"
              sink:
                - s3:
                    aws:
                      sts_role_arn: "${aws_iam_role.test.arn}"
                      region: "${data.aws_region.current.region}"
                    bucket: "test"
                    threshold:
                      event_collect_timeout: "60s"
                    codec:
                      ndjson:
        EOT
  max_units                   = 1
  min_units                   = 1

  vpc_options {
    security_group_ids      = [aws_security_group.pipeline.id]
    subnet_ids              = [aws_subnet.pipeline.id]
    vpc_endpoint_management = "SERVICE"
  }
}

resource "aws_vpc" "pipeline" {
  region = var.region

  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_subnet" "pipeline" {
  region = var.region

  cidr_block = "10.0.1.0/24"
  vpc_id     = aws_vpc.pipeline.id
}

resource "aws_security_group" "pipeline" {
  region = var.region

  name   = var.rName
  vpc_id = aws_vpc.pipeline.id
}

data "aws_region" "current" {
  region = var.region

}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "osis-pipelines.amazonaws.com"
        }
      },
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
