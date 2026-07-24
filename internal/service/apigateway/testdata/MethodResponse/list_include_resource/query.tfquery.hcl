# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_api_gateway_method_response" "test" {
  provider = aws

  include_resource = true

  config {
    rest_api_id = aws_api_gateway_rest_api.test.id
    resource_id = aws_api_gateway_resource.test.id
    http_method = aws_api_gateway_method.test.http_method
  }
}
