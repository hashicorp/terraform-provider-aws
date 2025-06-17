# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_appflow_connector_profile" "test" {
  name            = var.rName
  connector_type  = "Redshift"
  connection_mode = "Public"

  connector_profile_config {

    connector_profile_credentials {
      redshift {
        password = aws_redshift_cluster.test.master_password
        username = aws_redshift_cluster.test.master_username
      }
    }

    connector_profile_properties {
      redshift {
        bucket_name        = var.rName
        cluster_identifier = aws_redshift_cluster.test.cluster_identifier
        database_name      = "dev"
        database_url       = "jdbc:redshift://${aws_redshift_cluster.test.endpoint}/dev"
        data_api_role_arn  = aws_iam_role.test.arn
        role_arn           = aws_iam_role.test.arn
      }
    }
  }

  depends_on = [
    aws_route.test,
    aws_security_group_rule.test,
  ]
}

# testAccConnectorProfileConfig_base

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

data "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route" "test" {
  route_table_id         = data.aws_route_table.test.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
}

resource "aws_redshift_subnet_group" "test" {
  name       = var.rName
  subnet_ids = aws_subnet.test[*].id
}

data "aws_iam_policy" "test" {
  name = "AmazonRedshiftFullAccess"
}

resource "aws_iam_role" "test" {
  name = var.rName

  managed_policy_arns = [data.aws_iam_policy.test.arn]

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "appflow.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_security_group" "test" {
  name   = var.rName
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group_rule" "test" {
  type = "ingress"

  security_group_id = aws_security_group.test.id

  from_port   = 0
  to_port     = 65535
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier = var.rName

  availability_zone         = data.aws_availability_zones.available.names[0]
  cluster_subnet_group_name = aws_redshift_subnet_group.test.name
  vpc_security_group_ids    = [aws_security_group.test.id]

  master_password = "testPassword123!"
  master_username = "testusername"

  publicly_accessible = false

  node_type           = "ra3.large"
  skip_final_snapshot = true
  encrypted           = true
}

# acctest.ConfigVPCWithSubnets(rName, 1)

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 1

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
