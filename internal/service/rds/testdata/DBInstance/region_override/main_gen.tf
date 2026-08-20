# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

ephemeral "aws_secretsmanager_random_password" "test" {
  region = var.region

  password_length     = 20
  exclude_punctuation = true
}

resource "aws_db_instance" "test" {
  region = var.region

  identifier          = var.rName
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  engine_version      = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot = true
  password_wo         = ephemeral.aws_secretsmanager_random_password.test.random_password
  password_wo_version = 1
  username            = "tfacctest"
}

# testAccInstanceConfig_orderableClassMySQL

data "aws_rds_engine_version" "default" {
  region = var.region

  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  region = var.region

  engine         = data.aws_rds_engine_version.default.engine
  engine_version = data.aws_rds_engine_version.default.version
  license_model  = "general-public-license"
  storage_type   = "gp2"

  preferred_instance_classes = ["db.t4g.micro", "db.t4g.small"]
}

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
