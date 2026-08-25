# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_redshift_namespace_registration" "test" {
  region = var.region

  consumer_identifier             = format("DataCatalog/%s", data.aws_caller_identity.current.account_id)
  namespace_type                  = "provisioned"
  provisioned_cluster_identifier = aws_redshift_cluster.test.cluster_identifier
}

resource "aws_redshift_cluster" "test" {
  region = var.region

  cluster_identifier  = var.rName
  database_name       = "test"
  master_username     = "testuser"
  master_password     = "Testpass123"
  node_type           = "ra3.large"
  cluster_type        = "single-node"
  skip_final_snapshot = true
}

data "aws_caller_identity" "current" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
