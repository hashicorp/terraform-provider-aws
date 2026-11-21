# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_apigatewayv2_portal" "test" {
  authorization {
    cognito_config {
      app_client_id    = aws_cognito_user_pool_client.test.id
      user_pool_arn    = aws_cognito_user_pool.test.arn
      user_pool_domain = "${aws_cognito_user_pool_domain.test.domain}.auth.${data.aws_region.current.region}.amazoncognito.com" # nosemgrep:ci.semgrep.domain-names.domain-names-tf
    }
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

resource "aws_cognito_user_pool" "test" {
  name = var.rName
}

resource "aws_cognito_user_pool_client" "test" {
  name         = var.rName
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = var.rName
  user_pool_id = aws_cognito_user_pool.test.id
}

data "aws_region" "current" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
