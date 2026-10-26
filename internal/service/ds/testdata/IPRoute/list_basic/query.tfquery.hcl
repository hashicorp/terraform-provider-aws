# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_directory_service_ip_route" "test" {
  provider = aws

  config {
    directory_id = aws_directory_service_directory.test.id
  }
}
