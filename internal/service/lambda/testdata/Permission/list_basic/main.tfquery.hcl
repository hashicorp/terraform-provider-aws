# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_lambda_permission" "test" {
  provider = aws

  config {
    function_name = aws_lambda_function.test.function_name
  }
}
