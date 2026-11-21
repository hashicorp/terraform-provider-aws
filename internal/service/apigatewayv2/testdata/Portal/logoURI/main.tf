# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_apigatewayv2_portal" "test" {
  authorization {
    none {}
  }

  endpoint_configuration {
    none {}
  }

  portal_content {
    display_name = var.rName

    theme {
      custom_colors {
        accent_color           = "#FF9900"
        background_color       = "#FFFFFF"
        error_validation_color = "#D13212"
        header_color           = "#232F3E"
        navigation_color       = "#232F3E"
        text_color             = "#16191F"
      }
    }
  }

  logo_uri = var.logo ? "s3://${aws_s3_object.test.bucket}/${aws_s3_object.test.key}" : null
}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket       = aws_s3_bucket.test.bucket
  key          = "logo.png"
  source       = "test-fixtures/portal-logo.png"
  content_type = "image/png"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "logo" {
  description = "Whether to set the logo"
  type        = bool
  nullable    = false
}
