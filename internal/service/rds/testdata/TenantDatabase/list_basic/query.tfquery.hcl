# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_rds_tenant_database" "test" {
  provider = aws

  config {
    db_instance_identifier = var.rName
  }
}
