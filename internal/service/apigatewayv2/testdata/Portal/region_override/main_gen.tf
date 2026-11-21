# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_apigatewayv2_portal" "test" {
  region = var.region

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
