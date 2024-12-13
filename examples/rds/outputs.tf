# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

output "subnet_group" {
  value = aws_db_subnet_group.default.name
}

output "db_instance_id" {
  value = aws_db_instance.default.identifier
}

output "db_instance_address" {
  value = aws_db_instance.default.address
}
