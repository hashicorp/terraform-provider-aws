# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_rds_integration" "test" {
  integration_name = var.rName
  source_arn       = aws_rds_cluster.test.arn
  target_arn       = aws_redshiftserverless_namespace.test.arn

  depends_on = [
    aws_rds_cluster.test,
    aws_rds_cluster_instance.test,
    aws_redshiftserverless_namespace.test,
    aws_redshiftserverless_workgroup.test,
    aws_redshift_resource_policy.test,
  ]
}

# testAccIntegrationConfig_base

resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = var.rName
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = var.rName
  base_capacity  = 8

  publicly_accessible = false
  subnet_ids          = aws_subnet.test[*].id

  config_parameter {
    parameter_key   = "enable_case_sensitive_identifier"
    parameter_value = "true"
  }
  config_parameter {
    parameter_key   = "auto_mv"
    parameter_value = "true"
  }
  config_parameter {
    parameter_key   = "datestyle"
    parameter_value = "ISO, MDY"
  }
  config_parameter {
    parameter_key   = "enable_user_activity_logging"
    parameter_value = "true"
  }
  config_parameter {
    parameter_key   = "max_query_execution_time"
    parameter_value = "14400"
  }
  config_parameter {
    parameter_key   = "query_group"
    parameter_value = "default"
  }
  config_parameter {
    parameter_key   = "require_ssl"
    parameter_value = "true"
  }
  config_parameter {
    parameter_key   = "search_path"
    parameter_value = "$user, public"
  }
  config_parameter {
    parameter_key   = "use_fips_ssl"
    parameter_value = "false"
  }
}

# The "aws_redshiftserverless_resource_policy" resource doesn't support the following action types.
# Therefore we need to use the "aws_redshift_resource_policy" resource for RedShift-serverless instead.
resource "aws_redshift_resource_policy" "test" {
  resource_arn = aws_redshiftserverless_namespace.test.arn
  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action   = "redshift:CreateInboundIntegration"
      Resource = aws_redshiftserverless_namespace.test.arn
      }, {
      Effect = "Allow"
      Principal = {
        Service = "redshift.amazonaws.com"
      }
      Action   = "redshift:AuthorizeInboundIntegration"
      Resource = aws_redshiftserverless_namespace.test.arn
      Condition = {
        StringEquals = {
          "aws:SourceArn" = aws_rds_cluster.test.arn
        }
      }
    }]
  })
}

# testAccIntegrationConfig_baseClusterWithInstance

locals {
  cluster_parameters = {
    "binlog_replication_globaldb" = {
      value        = "0"
      apply_method = "pending-reboot"
    },
    "binlog_format" = {
      value        = "ROW"
      apply_method = "pending-reboot"
    },
    "binlog_row_metadata" = {
      value        = "full"
      apply_method = "immediate"
    },
    "binlog_row_image" = {
      value        = "full"
      apply_method = "immediate"
    },
    "aurora_enhanced_binlog" = {
      value        = "1"
      apply_method = "pending-reboot"
    },
    "binlog_backup" = {
      value        = "0"
      apply_method = "pending-reboot"
    },
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_security_group" "test" {
  name   = var.rName
  vpc_id = aws_vpc.test.id

  ingress {
    protocol  = -1
    self      = true
    from_port = 0
    to_port   = 0
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_subnet_group" "test" {
  name       = var.rName
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_rds_cluster_parameter_group" "test" {
  name   = var.rName
  family = "aurora-mysql8.0"

  dynamic "parameter" {
    for_each = local.cluster_parameters
    content {
      name         = parameter.key
      value        = parameter.value["value"]
      apply_method = parameter.value["apply_method"]
    }
  }
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = var.rName
  engine              = "aurora-mysql"
  engine_version      = "8.0.mysql_aurora.3.05.2"
  database_name       = "test"
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  skip_final_snapshot = true

  vpc_security_group_ids          = [aws_security_group.test.id]
  db_subnet_group_name            = aws_db_subnet_group.test.name
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.test.name

  apply_immediately = true
}

resource "aws_rds_cluster_instance" "test" {
  identifier         = var.rName
  cluster_identifier = aws_rds_cluster.test.id
  instance_class     = "db.r6g.large"
  engine             = aws_rds_cluster.test.engine
  engine_version     = aws_rds_cluster.test.engine_version
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier  = var.rName
  availability_zone   = data.aws_availability_zones.available.names[0]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "Mustbe8characters"
  node_type           = "ra3.large"
  cluster_type        = "single-node"
  skip_final_snapshot = true

  availability_zone_relocation_enabled = true
  publicly_accessible                  = false
  encrypted                            = true
}

# acctest.ConfigVPCWithSubnets(rName, 3)

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 3

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
  exclude_zone_ids = local.default_exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
