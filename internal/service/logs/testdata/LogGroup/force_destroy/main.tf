# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Example configuration demonstrating the force_destroy feature
resource "aws_cloudwatch_log_group" "example_with_force_destroy" {
  name              = "/aws/lambda/example-function"
  retention_in_days = 7
  force_destroy     = true

  tags = {
    Environment = "test"
    Purpose     = "example"
  }
}

# Example without force_destroy (default behavior)
resource "aws_cloudwatch_log_group" "example_without_force_destroy" {
  name              = "/aws/lambda/another-function"
  retention_in_days = 7
  force_destroy     = false  # This is the default

  tags = {
    Environment = "test"
    Purpose     = "example"
  }
}
