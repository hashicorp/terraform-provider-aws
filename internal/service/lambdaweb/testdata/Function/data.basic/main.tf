# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# tflint-ignore: terraform_unused_declarations
data "aws_lambdaweb_function" "test" {
  function_name = aws_lambdaweb_function.test.function_name
}

resource "aws_lambdaweb_function" "test" {
  function_name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}