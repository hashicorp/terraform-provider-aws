# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

variable "engine" {
  type    = string
  default = "neptune"
}

variable "rname" {
  type    = string
  default = "tf-neptune-test"
}

variable "upgrade" {
  type    = bool
  default = false
}

# Get the current and upgrade engine versions

data "aws_neptune_engine_version" "test" {
  engine                  = var.engine
  latest                  = true
  preferred_major_targets = [data.aws_neptune_engine_version.upgrade.version_actual]
}

data "aws_neptune_engine_version" "upgrade" {
  engine = var.engine
}

locals {
  engine_version         = var.upgrade ? data.aws_neptune_engine_version.upgrade.version_actual : data.aws_neptune_engine_version.test.version_actual
  parameter_group_family = var.upgrade ? data.aws_neptune_engine_version.upgrade.parameter_group_family : data.aws_neptune_engine_version.test.parameter_group_family
}

resource "aws_neptune_global_cluster" "test" {
  engine                    = var.engine
  engine_version            = local.engine_version
  global_cluster_identifier = var.rname
}

resource "aws_neptune_cluster" "test" {
  apply_immediately                    = true
  cluster_identifier                   = var.rname
  engine                               = aws_neptune_global_cluster.test.engine
  engine_version                       = aws_neptune_global_cluster.test.engine_version
  global_cluster_identifier            = aws_neptune_global_cluster.test.global_cluster_identifier
  neptune_cluster_parameter_group_name = "default.${local.parameter_group_family}"
  skip_final_snapshot                  = true
}

resource "aws_neptune_cluster_instance" "test" {
  apply_immediately            = true
  cluster_identifier           = aws_neptune_cluster.test.id
  engine                       = aws_neptune_cluster.test.engine
  engine_version               = aws_neptune_cluster.test.engine_version
  identifier                   = var.rname
  instance_class               = "db.r6g.large"
  neptune_parameter_group_name = "default.${local.parameter_group_family}"

  lifecycle {
    ignore_changes = [engine_version]
  }
}
