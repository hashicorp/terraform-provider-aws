# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

# tflint-ignore: terraform_unused_declarations
data "aws_guardduty_detector" "test" {
  id = aws_guardduty_detector.test.id
}

resource "aws_guardduty_detector" "test" {

  tags = var.resource_tags
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
