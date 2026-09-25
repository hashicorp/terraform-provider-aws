# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_directory_service_ip_routes" "test" {
  provider = aws

  config {
    region = var.region
  }
}
