# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Setup infrastructure for TestAccKafkaReplicator_selfManagedSASLSCRAM.
#
# The acceptance test provisions ONLY the replicator; it wires up to the infrastructure
# below via environment variables. Apply this config, then map its outputs to the env vars
# (see README.md), then run the test.
#
# The "self-managed" source is an MSK cluster impersonating a self-managed Apache Kafka
# cluster: its server certificate is signed by a public CA, so MSK Replicator trusts it
# without a custom root CA, and SASL/SCRAM users are created by associating a Secrets
# Manager secret to the cluster.

terraform {
  required_providers {
    aws    = { source = "hashicorp/aws" }
    random = { source = "hashicorp/random" }
  }
}

variable "name_prefix" {
  type    = string
  default = "tf-msk-replicator-test"
}

variable "kafka_version" {
  type    = string
  default = "3.6.0"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

# ---------------------------------------------------------------------------
# Networking (both clusters + the replicator ENIs share this VPC)
# ---------------------------------------------------------------------------
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags                 = { Name = var.name_prefix }
}

resource "aws_subnet" "test" {
  count             = 3
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  tags              = { Name = "${var.name_prefix}-${count.index}" }
}

# Self-referencing SG so the replicator ENIs (which we place in this SG) can reach the
# broker ENIs (also in this SG) on the Kafka ports.
resource "aws_security_group" "test" {
  name   = var.name_prefix
  vpc_id = aws_vpc.test.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = var.name_prefix }
}

# Interface VPC endpoints so the replicator ENIs (private IP only) can reach Secrets Manager
# and KMS to fetch and decrypt the SASL/SCRAM credentials. Without egress to these services
# (via endpoints or NAT) the replicator times out connecting to the source cluster.
resource "aws_security_group" "endpoints" {
  name   = "${var.name_prefix}-endpoints"
  vpc_id = aws_vpc.test.id

  ingress {
    from_port       = 443
    to_port         = 443
    protocol        = "tcp"
    security_groups = [aws_security_group.test.id]
  }

  tags = { Name = "${var.name_prefix}-endpoints" }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "secretsmanager" {
  vpc_id              = aws_vpc.test.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.secretsmanager"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = aws_subnet.test[*].id
  security_group_ids  = [aws_security_group.endpoints.id]
  private_dns_enabled = true
}

resource "aws_vpc_endpoint" "kms" {
  vpc_id              = aws_vpc.test.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.kms"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = aws_subnet.test[*].id
  security_group_ids  = [aws_security_group.endpoints.id]
  private_dns_enabled = true
}

# ---------------------------------------------------------------------------
# CMK + SASL/SCRAM secret (username/password). MSK requires a customer-managed
# key and an "AmazonMSK_"-prefixed secret name for SCRAM association.
# ---------------------------------------------------------------------------
resource "aws_kms_key" "scram" {
  description = "${var.name_prefix} MSK SASL/SCRAM secret"
}

resource "random_password" "scram" {
  length  = 32
  special = false
}

resource "aws_secretsmanager_secret" "scram" {
  name       = "AmazonMSK_${var.name_prefix}"
  kms_key_id = aws_kms_key.scram.key_id
}

resource "aws_secretsmanager_secret_version" "scram" {
  secret_id     = aws_secretsmanager_secret.scram.id
  secret_string = jsonencode({ username = "msk-replicator", password = random_password.scram.result })
}

# ---------------------------------------------------------------------------
# Source cluster (impersonates the self-managed Apache Kafka cluster), SASL/SCRAM enabled
# ---------------------------------------------------------------------------
resource "aws_msk_cluster" "source" {
  cluster_name           = "${var.name_prefix}-source"
  kafka_version          = var.kafka_version
  number_of_broker_nodes = 3

  broker_node_group_info {
    instance_type   = "kafka.m5.large"
    client_subnets  = aws_subnet.test[*].id
    security_groups = [aws_security_group.test.id]

    storage_info {
      ebs_storage_info {
        volume_size = 10
      }
    }
  }

  client_authentication {
    sasl {
      scram = true
    }
  }

  encryption_info {
    encryption_in_transit {
      client_broker = "TLS"
      in_cluster    = true
    }
  }
}

# Creates the SCRAM user on the source cluster from the secret.
resource "aws_msk_scram_secret_association" "source" {
  cluster_arn     = aws_msk_cluster.source.arn
  secret_arn_list = [aws_secretsmanager_secret.scram.arn]

  depends_on = [aws_secretsmanager_secret_version.scram]
}

# ---------------------------------------------------------------------------
# Target cluster
# ---------------------------------------------------------------------------
resource "aws_msk_cluster" "target" {
  cluster_name           = "${var.name_prefix}-target"
  kafka_version          = var.kafka_version
  number_of_broker_nodes = 3

  broker_node_group_info {
    instance_type   = "kafka.m5.large"
    client_subnets  = aws_subnet.test[*].id
    security_groups = [aws_security_group.test.id]

    storage_info {
      ebs_storage_info {
        volume_size = 10
      }
    }
  }
}

# ---------------------------------------------------------------------------
# Outputs -> environment variables (see README.md). MSK_ONPREM_KAFKA_CLUSTER_ID is NOT
# here: the Kafka cluster.id is not exposed by any control-plane API and must be fetched
# from the source cluster's data plane (see README.md).
# ---------------------------------------------------------------------------
output "MSK_ONPREM_KAFKA_BOOTSTRAP_BROKERS" {
  description = "Source SASL/SCRAM (TLS) bootstrap brokers."
  value       = aws_msk_cluster.source.bootstrap_brokers_sasl_scram
}

output "MSK_ONPREM_KAFKA_SASL_SCRAM_SECRET_ARN" {
  value = aws_secretsmanager_secret.scram.arn
}

output "MSK_ONPREM_KAFKA_TARGET_CLUSTER_ARN" {
  value = aws_msk_cluster.target.arn
}

output "MSK_ONPREM_KAFKA_SUBNET_IDS" {
  description = "Comma-separated subnet IDs for the replicator ENIs."
  value       = join(",", aws_subnet.test[*].id)
}

output "MSK_ONPREM_KAFKA_SECURITY_GROUP_IDS" {
  value = aws_security_group.test.id
}
