# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_apigatewayv2_route" "test" {
  provider = aws

  config {
    api_id = aws_apigatewayv2_api.test.id
  }
}
