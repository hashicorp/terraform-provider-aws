resource "aws_ecs_gateway_cluster_association" "test" {
{{- template "region" }}
  gateway_arn = aws_ecs_gateway.test.arn
  cluster_arn = aws_ecs_cluster.test.arn
}

resource "aws_ecs_cluster" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_ecs_gateway" "test" {
{{- template "region" }}
  name                    = var.rName
  infrastructure_role_arn = aws_iam_role.test.arn
  deletion_protection     = false

  configuration {
    host_shared_alb {
      network_configuration {
        subnets = aws_subnet.test[*].id
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "test_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test_assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonECSInfrastructureRoleForGatewayHostSharedALB"
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
