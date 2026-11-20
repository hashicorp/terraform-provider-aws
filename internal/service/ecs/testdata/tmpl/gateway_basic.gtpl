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
{{- template "tags" . }}

  depends_on = [aws_iam_role_policy_attachment.test]
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs.${data.aws_partition.current.dns_suffix}" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonECSInfrastructureRoleForGatewayHostSharedALB"
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
