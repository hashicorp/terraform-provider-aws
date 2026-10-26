# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_feature_v2" "test" {
  feature_name   = "NETWORK_SCANNING"
  feature_status = var.feature_status

  depends_on = [aws_securityhub_account_v2.test]
}

resource "aws_securityhub_account_v2" "test" {}

variable "feature_status" {
  type     = string
  nullable = false
}