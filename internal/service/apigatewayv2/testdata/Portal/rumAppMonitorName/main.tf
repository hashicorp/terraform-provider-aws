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

  rum_app_monitor_name = var.rum_app_monitor ? aws_rum_app_monitor.test.name : null
}

resource "aws_rum_app_monitor" "test" {
  name   = var.rName
  domain = "example.com"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "rum_app_monitor" {
  description = "Whether to attach the RUM app monitor"
  type        = bool
  nullable    = false
}
