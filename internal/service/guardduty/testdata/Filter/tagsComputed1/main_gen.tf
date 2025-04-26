# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_guardduty_filter" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "test-filter"
  action      = "ARCHIVE"
  rank        = 1

  finding_criteria {
    criterion {
      field  = "region"
      equals = [data.aws_region.current.name]
    }
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_guardduty_detector" "test" {}

data "aws_region" "current" {}

resource "null_resource" "test" {}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
