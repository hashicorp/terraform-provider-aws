# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_sagemaker_hub_content_reference" "test" {
  hub_name                         = aws_sagemaker_hub.test.hub_name
  hub_content_name                 = var.rName
  sagemaker_public_hub_content_arn = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.name}:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct"
}

data "aws_partition" "current" {}
data "aws_region" "current" {
}

resource "aws_sagemaker_hub" "test" {
  hub_name        = var.rName
  hub_description = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
