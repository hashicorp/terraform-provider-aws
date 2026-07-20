# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_osis_pipeline" "test" {
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

  tags = var.resource_tags
}

data "aws_region" "current" {
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

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
