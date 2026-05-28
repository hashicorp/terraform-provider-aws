# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_appflow_connector_profile" "test" {
  region = var.region

  count = var.resource_count

  name            = "${var.rName}-${count.index}"
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
}

data "aws_availability_zones" "available" {
  region = var.region

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  region = var.region

  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_internet_gateway" "test" {
  region = var.region

  vpc_id = aws_vpc.test.id
}

data "aws_route_table" "test" {
  region = var.region

  vpc_id = aws_vpc.test.id
}

resource "aws_route" "test" {
  region = var.region

  route_table_id         = data.aws_route_table.test.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
}

resource "aws_redshift_subnet_group" "test" {
  region = var.region

  name       = var.rName
  subnet_ids = [aws_subnet.test.id]
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
  region = var.region

  name   = var.rName
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group_rule" "test" {
  region = var.region

  type = "ingress"

  security_group_id = aws_security_group.test.id

  from_port   = 0
  to_port     = 65535
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]
}

resource "aws_redshift_cluster" "test" {
  region = var.region

  cluster_identifier = var.rName

  availability_zone         = data.aws_availability_zones.available.names[0]
  cluster_subnet_group_name = aws_redshift_subnet_group.test.name
  vpc_security_group_ids    = [aws_security_group.test.id]

  master_password = "TestPassword123!"
  master_username = "testusername"

  publicly_accessible = false

  node_type           = "ra3.large"
  skip_final_snapshot = true
  encrypted           = true
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
