# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_lambdamicrovms_microvm" "test" {
  provider = aws

  config {
    image_arn = aws_lambdamicrovms_image.test[0].arn
  }
}
