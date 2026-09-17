data "aws_ssoadmin_instances" "test" {
{{- template "region" -}}
}

resource "aws_iam_policy" "test" {
  name = var.rName

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = "s3:ListBucket"
      Effect   = "Allow"
      Resource = "*"
    }]
  })
}

resource "aws_ssoadmin_permission_set" "test" {
{{- template "region" }}
  name         = var.rName
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_customer_managed_policy_attachments_exclusive" "test" {
{{- template "region" }}
  instance_arn       = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  permission_set_arn = aws_ssoadmin_permission_set.test.arn

  customer_managed_policy_reference {
    name = aws_iam_policy.test.name
  }
}
