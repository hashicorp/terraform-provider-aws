# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

variable "odb_network_id" {
  description = "ID of an existing ODB network"
  type        = string
}

variable "autonomous_database_admin_password" {
  description = "ADMIN password for the Autonomous Database"
  type        = string
}

resource "aws_odb_autonomous_database" "example" {
  admin_password           = var.autonomous_database_admin_password
  compute_count            = 2
  data_storage_size_in_tbs = 1
  db_name                  = "TFADBEXAMPLE"
  db_workload              = "OLTP"
  display_name             = "terraform-adbs-example"
  license_model            = "LICENSE_INCLUDED"
  odb_network_id           = var.odb_network_id
  source                   = "NONE"

  tags = {
    Environment = "example"
  }
}

data "aws_odb_autonomous_database" "example" {
  id = aws_odb_autonomous_database.example.id
}

output "autonomous_database_status" {
  value = data.aws_odb_autonomous_database.example.status
}
