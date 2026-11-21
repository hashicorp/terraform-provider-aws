# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cognito_user_pool_session_configuration" "test" {
  user_pool_id          = aws_cognito_user_pool.test.id
  user_session_tracking = "ENABLED"

  token_session {
    max_per_user = 20
  }
}

resource "aws_cognito_user_pool" "test" {
  name           = var.rName
  user_pool_tier = "ESSENTIALS"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
