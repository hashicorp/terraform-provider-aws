resource "aws_config_organization_managed_rule" "test" {
{{- template "region" }}
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  name            = var.rName
  rule_identifier = "IAM_PASSWORD_POLICY"
}

data "aws_partition" "current" {}

resource "aws_config_configuration_recorder" "test" {
{{- template "region" }}
  depends_on = [aws_iam_role_policy_attachment.test]

  name     = var.rName
  role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWS_ConfigRole"
  role       = aws_iam_role.test.name
}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config-multiaccountsetup.amazonaws.com"]
  feature_set                   = "ALL"
}
