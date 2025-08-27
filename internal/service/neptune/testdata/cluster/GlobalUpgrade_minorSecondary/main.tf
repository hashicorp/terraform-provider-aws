# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "awsalternate" {
  region = var.alt_region
}

variable "engine" {
  type    = string
  default = "neptune"
}

variable "rname" {
  type    = string
  default = "tf-neptune-test"
}

variable "alt_rname" {
  type    = string
  default = "tf-neptune-test-alternate"
}

variable "upgrade" {
  type    = bool
  default = false
}

variable "alt_region" {
  type    = string
  default = "us-west-2"
}

locals {
  engine_version = var.upgrade ? data.aws_neptune_engine_version.upgrade.version_actual : data.aws_neptune_engine_version.test.version_actual
}

data "aws_neptune_engine_version" "test" {
  engine                    = var.engine
  latest                    = true
  preferred_upgrade_targets = [data.aws_neptune_engine_version.upgrade.version_actual]
}

data "aws_neptune_engine_version" "upgrade" {
  engine = var.engine
}

resource "aws_neptune_global_cluster" "test" {
  global_cluster_identifier = var.rname
  engine                    = var.engine
  engine_version            = local.engine_version
}

resource "aws_neptune_cluster" "primary" {
  apply_immediately         = true
  cluster_identifier        = var.rname
  engine                    = aws_neptune_global_cluster.test.engine
  engine_version            = aws_neptune_global_cluster.test.engine_version
  global_cluster_identifier = aws_neptune_global_cluster.test.global_cluster_identifier
  skip_final_snapshot       = true
}

resource "aws_neptune_cluster_instance" "primary" {
  apply_immediately  = true
  cluster_identifier = aws_neptune_cluster.primary.id
  engine             = aws_neptune_cluster.primary.engine
  engine_version     = aws_neptune_cluster.primary.engine_version
  identifier         = var.rname
  instance_class     = "db.r6g.large"
}

data "aws_availability_zones" "alternate" {
  provider = awsalternate

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alternate" {
  provider = awsalternate

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = var.alt_rname
  }
}

resource "aws_subnet" "alternate" {
  provider = awsalternate
  count    = 3

  availability_zone = data.aws_availability_zones.alternate.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.alternate.id

  tags = {
    Name = var.alt_rname
  }
}

resource "aws_neptune_subnet_group" "alternate" {
  provider = awsalternate

  name       = var.alt_rname
  subnet_ids = aws_subnet.alternate[*].id
}

resource "aws_neptune_cluster" "secondary" {
  provider = awsalternate

  apply_immediately         = true
  cluster_identifier        = var.alt_rname
  engine                    = aws_neptune_cluster_instance.primary.engine
  engine_version            = aws_neptune_cluster_instance.primary.engine_version
  global_cluster_identifier = aws_neptune_cluster.primary.global_cluster_identifier
  neptune_subnet_group_name = aws_neptune_subnet_group.alternate.name
  skip_final_snapshot       = true

  lifecycle {
    ignore_changes = [replication_source_identifier]
  }
}

resource "aws_neptune_cluster_instance" "secondary" {
  provider = awsalternate

  apply_immediately  = true
  cluster_identifier = aws_neptune_cluster.secondary.cluster_identifier
  engine_version     = aws_neptune_global_cluster.test.engine_version
  identifier         = var.alt_rname
  instance_class     = "db.r6g.large"
}
