resource "aws_eks_fargate_profile" "test" {
{{- template "region" }}
  cluster_name           = aws_eks_cluster.test.name
  fargate_profile_name   = "${var.rName}-profile"
  pod_execution_role_arn = aws_iam_role.pod.arn
  subnet_ids             = aws_subnet.private[*].id

  selector {
    namespace = "test"
  }

{{- template "tags" . }}

  depends_on = [
    aws_iam_role_policy_attachment.pod-AmazonEKSFargatePodExecutionRolePolicy,
    aws_route_table_association.private,
  ]
}

resource "aws_eks_cluster" "test" {
{{- template "region" }}
  name     = "${var.rName}-cluster"
  role_arn = aws_iam_role.cluster.arn

  vpc_config {
    subnet_ids = aws_subnet.public[*].id
  }

  depends_on = [
    aws_iam_role_policy_attachment.cluster-AmazonEKSClusterPolicy,
    aws_main_route_table_association.test,
  ]
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

data "aws_service_principal" "eks_fargate_pods" {
  service_name = "eks-fargate-pods"
}

resource "aws_iam_role" "pod" {
  name = "${var.rName}-pod"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = data.aws_service_principal.eks_fargate_pods.name
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "pod-AmazonEKSFargatePodExecutionRolePolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSFargatePodExecutionRolePolicy"
  role       = aws_iam_role.pod.name
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name                                 = var.rName
    "kubernetes.io/cluster/${var.rName}" = "shared"
  }
}

resource "aws_internet_gateway" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id

  tags = {
    Name = var.rName
  }
}

resource "aws_route_table" "public" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = var.rName
  }
}

resource "aws_main_route_table_association" "test" {
{{- template "region" }}
  route_table_id = aws_route_table.public.id
  vpc_id         = aws_vpc.test.id
}

resource "aws_subnet" "private" {
{{- template "region" }}
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index + 2)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name                                 = var.rName
    "kubernetes.io/cluster/${var.rName}" = "shared"
  }
}

resource "aws_eip" "private" {
{{- template "region" }}
  count      = 2
  depends_on = [aws_internet_gateway.test]

  domain = "vpc"

  tags = {
    Name = var.rName
  }
}

resource "aws_nat_gateway" "private" {
{{- template "region" }}
  count = 2

  allocation_id = aws_eip.private[count.index].id
  subnet_id     = aws_subnet.private[count.index].id

  tags = {
    Name = var.rName
  }
}

resource "aws_route_table" "private" {
{{- template "region" }}
  count = 2

  vpc_id = aws_vpc.test.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.private[count.index].id
  }

  tags = {
    Name = var.rName
  }
}

resource "aws_route_table_association" "private" {
{{- template "region" }}
  count = 2

  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[count.index].id
}

resource "aws_subnet" "public" {
{{- template "region" }}
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name                                 = var.rName
    "kubernetes.io/cluster/${var.rName}" = "shared"
  }
}

{{ template "acctest.ConfigAvailableAZsNoOptIn" }}
