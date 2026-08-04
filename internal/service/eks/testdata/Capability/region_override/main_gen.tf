# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_eks_capability" "test" {
  region = var.region

  cluster_name              = aws_eks_cluster.test.name
  capability_name           = "${var.rName}-capability"
  type                      = "KRO"
  role_arn                  = aws_iam_role.capability.arn
  delete_propagation_policy = "RETAIN"

  depends_on = [aws_iam_role_policy_attachment.capability]
}

resource "aws_eks_cluster" "test" {
  region = var.region

  name     = "${var.rName}-cluster"
  role_arn = aws_iam_role.cluster.arn

  access_config {
    authentication_mode                         = "API"
    bootstrap_cluster_creator_admin_permissions = true
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}

resource "aws_iam_role" "capability" {
  name = "${var.rName}-capability"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "capabilities.eks.amazonaws.com"
      }
      Action = [
        "sts:AssumeRole",
        "sts:TagSession"
      ]
    }]
  })
}

resource "aws_iam_role_policy_attachment" "capability" {
  role       = aws_iam_role.capability.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AdministratorAccess"
}

data "aws_partition" "current" {}

data "aws_service_principal" "eks" {
  service_name = "eks"
}

resource "aws_iam_role" "cluster" {
  name = "${var.rName}-cluster"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "${data.aws_service_principal.eks.name}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.0.0.0/16"

  tags = {
    Name                                 = var.rName
    "kubernetes.io/cluster/${var.rName}" = "shared"
  }
}

resource "aws_subnet" "test" {
  region = var.region

  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name                                 = var.rName
    "kubernetes.io/cluster/${var.rName}" = "shared"
  }
}

# acctest.ConfigAvailableAZsNoOptIn

data "aws_availability_zones" "available" {
  region = var.region

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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
