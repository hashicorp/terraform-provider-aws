resource "aws_quicksight_vpc_connection" "test" {
  vpc_connection_id = var.rName
  name              = var.rName
  role_arn          = aws_iam_role.test.arn
  security_group_ids = [
    aws_security_group.test.id,
  ]
  subnet_ids = aws_subnet.test[*].id
{{- template "tags" . }}
}

# testAccBaseVPCConnectionConfig

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = "sts:AssumeRole"
        Principal = {
          Service = "quicksight.amazonaws.com"
        }
      }
    ]
  })
  inline_policy {
    name = "QuicksightVPCConnectionRolePolicy"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Effect = "Allow"
          Action = [
            "ec2:CreateNetworkInterface",
            "ec2:ModifyNetworkInterfaceAttribute",
            "ec2:DeleteNetworkInterface",
            "ec2:DescribeSubnets",
            "ec2:DescribeSecurityGroups"
          ]
          Resource = ["*"]
        }
      ]
    })
  }
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
