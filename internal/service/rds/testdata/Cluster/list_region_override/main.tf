# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_rds_cluster" "test" {
  count  = var.resource_count
  region = var.region

  cluster_identifier  = "${var.rName}-${count.index}"
  database_name       = "test"
  engine              = "aurora-mysql"
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  skip_final_snapshot = true
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
