# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_msk_topic" "test" {
  provider = aws

  config {
    cluster_arn = aws_msk_cluster.test.arn
    region      = var.region
  }
}
