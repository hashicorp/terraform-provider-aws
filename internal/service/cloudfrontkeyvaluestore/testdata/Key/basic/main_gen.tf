# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudfrontkeyvaluestore_key" "test" {
  key                 = var.rName
  key_value_store_arn = aws_cloudfront_key_value_store.test.arn
  value               = var.rName
}

resource "aws_cloudfront_key_value_store" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
