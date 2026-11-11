# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_rds_tenant_database" "test" {
  count  = var.resource_count
  region = var.region

  db_instance_identifier          = aws_db_instance.test.identifier
  master_user_password_wo         = ephemeral.aws_secretsmanager_random_password.test.random_password
  master_user_password_wo_version = 1
  master_username                 = "tfacctest"
  skip_final_snapshot             = true
  tenant_db_name                  = "TFPDB${count.index}"
}

resource "aws_db_instance" "test" {
  region = var.region

  allocated_storage   = 20
  apply_immediately   = true
  db_name             = "TFTEST"
  engine              = data.aws_rds_orderable_db_instance.test.engine
  engine_version      = data.aws_rds_orderable_db_instance.test.engine_version
  identifier          = var.rName
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  license_model       = data.aws_rds_orderable_db_instance.test.license_model
  multi_tenant        = true
  password_wo         = ephemeral.aws_secretsmanager_random_password.test.random_password
  password_wo_version = 1
  skip_final_snapshot = true
  storage_type        = data.aws_rds_orderable_db_instance.test.storage_type
  username            = "tfacctest"
}

ephemeral "aws_secretsmanager_random_password" "test" {
  region = var.region

  password_length     = 20
  exclude_punctuation = true
}

# The multi-tenant configuration requires Oracle Database 19c RU 2022-01 or higher.
data "aws_rds_engine_version" "test" {
  region = var.region

  engine  = "oracle-se2-cdb"
  version = "19"
  latest  = true
}

data "aws_rds_orderable_db_instance" "test" {
  region = var.region

  engine         = data.aws_rds_engine_version.test.engine
  engine_version = data.aws_rds_engine_version.test.version_actual
  license_model  = "bring-your-own-license"
  storage_type   = "gp3"

  preferred_instance_classes = ["db.t3.small", "db.t3.medium", "db.m5.large"]
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
