# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

output "lambda" {
  value = aws_lambda_function.example_lambda.qualified_arn
}
