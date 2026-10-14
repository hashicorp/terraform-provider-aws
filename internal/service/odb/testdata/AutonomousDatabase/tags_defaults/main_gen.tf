# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
}

variable "odb_test_admin_password" {
  type      = string
  sensitive = true
}

variable "odb_test_network_id" {
  type = string
}

resource "aws_odb_autonomous_database" "test" {

  admin_password           = var.odb_test_admin_password
  character_set            = "AL32UTF8"
  compute_count            = 2
  data_storage_size_in_tbs = 1
  db_name                  = "TF${substr(md5(var.rName), 0, 12)}"
  db_workload              = "OLTP"
  display_name             = var.rName
  license_model            = "LICENSE_INCLUDED"
  odb_network_id           = var.odb_test_network_id
  source                   = "NONE"

  tags = var.resource_tags
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

variable "provider_tags" {
  type     = map(string)
  nullable = false
}
