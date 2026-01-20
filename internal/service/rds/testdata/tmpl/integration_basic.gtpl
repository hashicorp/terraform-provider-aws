resource "aws_rds_integration" "test" {
{{- template "region" }}
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
{{- template "tags" . }}
}

# testAccIntegrationConfig_base

resource "aws_redshiftserverless_namespace" "test" {
{{- template "region" }}
  namespace_name = var.rName
}

resource "aws_redshiftserverless_workgroup" "test" {
{{- template "region" }}
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
{{- template "region" }}
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
{{- template "region" }}
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
{{- template "region" }}
  name       = var.rName
  subnet_ids = aws_subnet.test[*].id
}

data "aws_rds_engine_version" "test" {
{{- template "region" }}
  engine  = "aurora-mysql"
  version = "8.0"
  latest  = true
}

resource "aws_rds_cluster_parameter_group" "test" {
{{- template "region" }}
  name   = var.rName
  family = data.aws_rds_engine_version.test.parameter_group_family

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
{{- template "region" }}
  cluster_identifier  = var.rName
  engine              = data.aws_rds_engine_version.test.engine
  engine_version      = data.aws_rds_engine_version.test.version_actual
  database_name       = "test"
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  skip_final_snapshot = true

  vpc_security_group_ids          = [aws_security_group.test.id]
  db_subnet_group_name            = aws_db_subnet_group.test.name
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.test.name

  apply_immediately = true
}

data "aws_rds_orderable_db_instance" "test" {
{{- template "region" }}
  engine                     = data.aws_rds_engine_version.test.engine
  engine_version             = data.aws_rds_engine_version.test.version_actual
  preferred_instance_classes = local.mainInstanceClasses
  supports_clusters          = true
  supports_global_databases  = true
}

resource "aws_rds_cluster_instance" "test" {
{{- template "region" }}
  identifier         = var.rName
  cluster_identifier = aws_rds_cluster.test.cluster_identifier
  engine             = aws_rds_cluster.test.engine
  engine_version     = aws_rds_cluster.test.engine_version
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}

resource "aws_redshift_cluster" "test" {
{{- template "region" }}
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

{{ template "acctest.ConfigVPCWithSubnets" 3 }}

locals {
  mainInstanceClasses = [
    "db.t4g.micro",
    "db.t3.micro",
    "db.t4g.small",
    "db.t3.small",
    "db.t4g.medium",
    "db.t3.medium",
    "db.t4g.large",
    "db.t3.large",
    "db.m6g.large",
    "db.m7g.large",
    "db.m5.large",
    "db.m6i.large",
    "db.m6gd.large",
    "db.m5d.large",
    "db.r6g.large",
    "db.m6id.large",
    "db.r7g.large",
    "db.r5.large",
    "db.r6i.large",
    "db.r6gd.large",
    "db.m6in.large",
    "db.t4g.xlarge",
    "db.t3.xlarge",
    "db.r5d.large",
    "db.m6idn.large",
    "db.r5b.large",
    "db.r6id.large",
    "db.m6g.xlarge",
    "db.x2g.large",
    "db.m7g.xlarge",
    "db.m5.xlarge",
    "db.m6i.xlarge",
    "db.r6in.large",
    "db.m6gd.xlarge",
    "db.r6idn.large",
    "db.m5d.xlarge",
    "db.r6g.xlarge",
    "db.m6id.xlarge",
    "db.r7g.xlarge",
    "db.r5.xlarge",
    "db.r6i.xlarge",
    "db.r6gd.xlarge",
    "db.m6in.xlarge",
    "db.t4g.2xlarge",
    "db.t3.2xlarge",
    "db.r5d.xlarge",
    "db.m6idn.xlarge",
    "db.r5b.xlarge",
    "db.r6id.xlarge",
  ]
}
