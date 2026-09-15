# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_rds_cluster" "test" {
  region = var.region

  cluster_identifier  = var.rName
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  count  = var.resource_count
  region = var.region

  identifier         = "${var.rName}-${count.index}"
  cluster_identifier = aws_rds_cluster.test.id
  engine             = aws_rds_cluster.test.engine
  engine_version     = aws_rds_cluster.test.engine_version
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}

data "aws_rds_engine_version" "default" {
  region = var.region
  engine = "aurora-mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  region         = var.region
  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version

  preferred_instance_classes = ["db.t4g.medium", "db.t3.medium"]
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
