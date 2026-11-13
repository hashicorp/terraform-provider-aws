# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_api_gateway_method" "test" {
  provider = aws

  config {
    rest_api_id = aws_api_gateway_rest_api.test.id
    resource_id = aws_api_gateway_resource.test.id
  }
}
