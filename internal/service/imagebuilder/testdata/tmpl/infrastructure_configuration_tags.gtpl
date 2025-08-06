resource "aws_imagebuilder_infrastructure_configuration" "test" {
{{- template "region" }}
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = var.rName
{{- template "tags" }}
}

resource "aws_iam_instance_profile" "test" {
  name = var.rName
  role = aws_iam_role.test.name
}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
  })
  name = var.rName
}
