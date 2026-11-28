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
    description  = var.portal_description

    theme {
      custom_colors {
        accent_color           = var.accent_color
        background_color       = "#FFFFFF"
        error_validation_color = "#D13212"
        header_color           = "#232F3E"
        navigation_color       = "#232F3E"
        text_color             = "#16191F"
      }
    }
  }
}

variable "display_name" {
  description = "Portal display name"
  type        = string
  nullable    = false
}

variable "portal_description" {
  description = "Portal description"
  type        = string
  nullable    = true
  default     = null
}

variable "accent_color" {
  description = "Accent color"
  type        = string
  nullable    = false
}
