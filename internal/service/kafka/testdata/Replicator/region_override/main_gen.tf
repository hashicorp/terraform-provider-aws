# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_msk_replicator" "test" {
  region = var.region

  replicator_name            = var.rName
  service_execution_role_arn = aws_iam_role.source.arn

  kafka_cluster {
    amazon_msk_cluster {
      msk_cluster_arn = aws_msk_cluster.source.arn
    }

    vpc_config {
      subnet_ids          = aws_subnet.source[*].id
      security_groups_ids = [aws_security_group.source.id]
    }
  }

  kafka_cluster {
    amazon_msk_cluster {
      msk_cluster_arn = aws_msk_cluster.target.arn
    }

    vpc_config {
      subnet_ids          = aws_subnet.target[*].id
      security_groups_ids = [aws_security_group.target.id]
    }
  }

  replication_info_list {
    source_kafka_cluster_arn = aws_msk_cluster.source.arn
    target_kafka_cluster_arn = aws_msk_cluster.target.arn
    target_compression_type  = "NONE"


    topic_replication {
      topics_to_replicate = [".*"]
    }

    consumer_group_replication {
      consumer_groups_to_replicate = [".*"]
    }
  }
}

# testAccReplicatorConfig_source

resource "aws_msk_configuration" "test" {
  region = var.region

  kafka_versions = ["3.8.x"]
  name           = var.rName

  server_properties = <<-PROPERTIES
    allow.everyone.if.no.acl.found = false
  PROPERTIES
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
  region = var.region

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

data "aws_caller_identity" "current" {}

resource "aws_msk_cluster" "source" {
  region = var.region

  cluster_name           = "${var.rName}-s"
  kafka_version          = "3.8.x"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = aws_subnet.source[*].id
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.source.id]

    connectivity_info {
      vpc_connectivity {
        client_authentication {
          sasl {
            iam = true
          }
        }
      }
    }

    storage_info {
      ebs_storage_info {
        volume_size = 10
      }
    }
  }

  configuration_info {
    arn      = aws_msk_configuration.test.arn
    revision = aws_msk_configuration.test.latest_revision
  }

  client_authentication {
    sasl {
      iam = true
    }
  }
}

resource "aws_msk_cluster_policy" "source" {
  region = var.region

  cluster_arn = aws_msk_cluster.source.arn

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Sid    = "testMskClusterPolicy"
      Effect = "Allow"
      Principal = {
        "Service" = "kafka.amazonaws.com"
      }
      Action = [
        "kafka:CreateVpcConnection",
        "kafka:GetBootstrapBrokers",
        "kafka:DescribeCluster",
        "kafka:DescribeClusterV2"
      ]
      Resource = aws_msk_cluster.source.arn
    }]
  })

  depends_on = [aws_msk_cluster.source]
}

resource "aws_iam_role" "source" {
  name = "${var.rName}-s"

  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "kafka.amazonaws.com"
        },
        "Action" : "sts:AssumeRole",
        "Condition" : {
          "StringEquals" : {
            "aws:SourceAccount" : data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "source" {
  name = "${var.rName}-s"
  role = aws_iam_role.source.name

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Resource" : "*",
        "Action" : [
          "kafka-cluster:Connect",
          "kafka-cluster:DescribeCluster",
          "kafka-cluster:AlterCluster",
          "kafka-cluster:ReadData",
          "kafka-cluster:WriteData",
          "kafka-cluster:DescribeTopic",
          "kafka-cluster:CreateTopic",
          "kafka-cluster:AlterTopic",
          "kafka-cluster:AlterGroup",
          "kafka-cluster:DescribeGroup",
          "kafka-cluster:DescribeTopicDynamicConfiguration",
          "kafka-cluster:AlterTopicDynamicConfiguration"
        ]
      }
    ]
  })
}

resource "aws_security_group" "source" {
  region = var.region

  name   = "${var.rName}-s"
  vpc_id = aws_vpc.source.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = -1
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_vpc" "source" {
  region = var.region

  cidr_block = "10.0.0.0/16"

  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_subnet" "source" {
  region = var.region

  count = 3

  vpc_id            = aws_vpc.source.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.source.cidr_block, 8, count.index)
}

# testAccReplicatorConfig_target

resource "aws_msk_cluster" "target" {
  region = var.region

  cluster_name           = "${var.rName}-t"
  kafka_version          = "3.8.x"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = aws_subnet.target[*].id
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.target.id]

    storage_info {
      ebs_storage_info {
        volume_size = 10
      }
    }
  }
  configuration_info {
    arn      = aws_msk_configuration.test.arn
    revision = aws_msk_configuration.test.latest_revision
  }

  client_authentication {
    sasl {
      iam = true
    }
  }
}

resource "aws_security_group" "target" {
  region = var.region

  name   = "${var.rName}-t"
  vpc_id = aws_vpc.target.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = -1
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_vpc" "target" {
  region = var.region

  cidr_block = "10.1.0.0/16"

  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_subnet" "target" {
  region = var.region

  count = 3

  vpc_id            = aws_vpc.target.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.target.cidr_block, 8, count.index)
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
