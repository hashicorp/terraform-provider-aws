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
    display_name = var.display_name

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

  publish             = var.publish
  publish_description = var.publish_description
}

variable "display_name" {
  description = "Portal display name"
  type        = string
  nullable    = false
}

variable "publish" {
  description = "Whether to publish the portal"
  type        = bool
  nullable    = false
}

variable "publish_description" {
  description = "Publish description"
  type        = string
  nullable    = true
  default     = null
}
