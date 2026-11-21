# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_apigatewayv2_portal_product_sharing_policy" "test" {
  region = var.region

  portal_product_id = aws_apigatewayv2_portal_product.test.portal_product_id

  policy_document = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action = [
        "apigateway:GetPortalProduct",
        "apigateway:ListProductPages",
        "apigateway:ListProductRestEndpointPages",
        "apigateway:GetProductPage",
        "apigateway:GetProductRestEndpointPage",
      ]
      Resource = [
        aws_apigatewayv2_portal_product.test.portal_product_arn,
        "${aws_apigatewayv2_portal_product.test.portal_product_arn}/*",
      ]
    }]
  })
}

resource "aws_apigatewayv2_portal_product" "test" {
  region = var.region

  display_name = var.rName
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

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
