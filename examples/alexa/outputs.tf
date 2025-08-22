# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

output "aws_lambda_function_arn" {
  value = aws_lambda_function.default.arn
}
