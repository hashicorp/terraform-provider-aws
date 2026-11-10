# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_lambdaweb_function_endpoint" "test" {
  provider = aws

  config {
    function_name = aws_lambdaweb_function.test.function_name
  }
}
