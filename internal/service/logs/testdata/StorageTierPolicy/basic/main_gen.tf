# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_storage_tier_policy" "test" {
  storage_tier = "INTELLIGENT_TIERING"
}

