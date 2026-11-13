resource "aws_config_organization_custom_policy_rule" "test" {
{{- template "region" }}
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  name = var.rName

  trigger_types  = ["ConfigurationItemChangeNotification"]
  policy_runtime = "guard-2.x.x"
  policy_text    = "let var = 5"
}

data "aws_partition" "current" {}

resource "aws_config_configuration_recorder" "test" {
{{- template "region" }}
  depends_on = [aws_iam_role_policy_attachment.config]

  name     = var.rName
  role_arn = aws_iam_role.config.arn
}

resource "aws_iam_role" "config" {
  name = "${var.rName}-config"

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

resource "aws_iam_role_policy_attachment" "config" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWS_ConfigRole"
  role       = aws_iam_role.config.name
}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config-multiaccountsetup.amazonaws.com"]
  feature_set                   = "ALL"
}
