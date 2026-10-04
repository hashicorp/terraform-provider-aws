# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_network_acl" "test" {
  provider = aws

  config {
    network_acl_ids = [
      aws_network_acl.test[0].id,
      aws_network_acl.test[1].id,
    ]
  }
}
