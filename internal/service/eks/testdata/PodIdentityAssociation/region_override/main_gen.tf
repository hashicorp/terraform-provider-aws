# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_eks_pod_identity_association" "test" {
  region = var.region

  cluster_name    = aws_eks_cluster.test.name
  namespace       = "${var.rName}-ns"
  service_account = "${var.rName}-sa"
  role_arn        = aws_iam_role.pods.arn
}

resource "aws_iam_role" "pods" {
  name = "${var.rName}-pods"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "pods.eks.amazonaws.com"
      },
      "Action": [
        "sts:AssumeRole",
        "sts:TagSession"
      ]
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "pods" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonS3ReadOnlyAccess"
  role       = aws_iam_role.pods.name
}

resource "aws_eks_cluster" "test" {
  region = var.region

  name     = var.rName
  role_arn = aws_iam_role.cluster.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.cluster-AmazonEKSClusterPolicy]
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

resource "aws_iam_role_policy_attachment" "cluster-AmazonEKSClusterPolicy" {
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
