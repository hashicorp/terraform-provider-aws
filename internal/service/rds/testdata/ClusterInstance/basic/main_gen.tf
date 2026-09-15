# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_rds_cluster_instance" "test" {
  identifier              = var.rName
  engine                  = data.aws_rds_engine_version.default.engine
  cluster_identifier      = aws_rds_cluster.test.id
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  db_parameter_group_name = aws_db_parameter_group.test.name
  promotion_tier          = "3"
}

resource "aws_db_parameter_group" "test" {
  name   = var.rName
  family = data.aws_rds_engine_version.default.parameter_group_family

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }
}

resource "aws_rds_cluster" "test" {
  cluster_identifier = var.rName
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  engine              = data.aws_rds_engine_version.default.engine
  engine_version      = data.aws_rds_engine_version.default.version
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

data "aws_rds_engine_version" "default" {
  engine = "aurora-mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                        = aws_rds_cluster.test.engine
  engine_version                = aws_rds_cluster.test.engine_version
  supports_performance_insights = true
  preferred_instance_classes    = ["db.t3.medium", "db.r5.large", "db.r4.large"]
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
