resource "aws_eks_cluster" "test" {
{{- template "region" }}
  name     = var.rName
  role_arn = aws_iam_role.cluster.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

{{- template "tags" . }}

  depends_on = [aws_iam_role_policy_attachment.cluster_AmazonEKSClusterPolicy]
}

data "aws_partition" "current" {}
data "aws_service_principal" "eks" {
{{- template "region" }}
  service_name = "eks"
}

resource "aws_iam_role" "cluster" {
  name = var.rName

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

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.0.0.0/16"

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name                                 = var.rName
    "kubernetes.io/cluster/${var.rName}" = "shared"
  }
}

resource "aws_subnet" "test" {
{{- template "region" }}
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)
  assign_ipv6_address_on_creation = true

  tags = {
    Name                                 = var.rName
    "kubernetes.io/cluster/${var.rName}" = "shared"
  }
}

{{ template "acctest.ConfigAvailableAZsNoOptIn" }}
