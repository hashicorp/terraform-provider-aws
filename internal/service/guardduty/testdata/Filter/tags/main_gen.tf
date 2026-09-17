# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_guardduty_filter" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "test-filter"
  action      = "ARCHIVE"
  rank        = 1

  finding_criteria {
    criterion {
      field  = "region"
      equals = [data.aws_region.current.region]
    }
  }

  tags = var.resource_tags
}

resource "aws_guardduty_detector" "test" {}

data "aws_region" "current" {}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
