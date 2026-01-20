# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_redshift_cluster_snapshot" "test" {
  cluster_identifier  = aws_redshift_cluster.test.cluster_identifier
  snapshot_identifier = var.rName

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }
}

# testAccClusterConfig_basic

resource "aws_redshift_cluster" "test" {
  cluster_identifier    = var.rName
  database_name         = "mydb"
  master_username       = "foo_test"
  master_password       = "Mustbe8characters"
  node_type             = "ra3.large"
  allow_version_upgrade = false
  skip_final_snapshot   = true
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}
