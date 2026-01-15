resource "aws_lakeformation_identity_center_configuration" "test" {
{{- template "region" }}
  instance_arn = local.identity_center_instance_arn
}

locals {
  identity_center_instance_arn = data.aws_ssoadmin_instances.test.arns[0]
}

data "aws_ssoadmin_instances" "test" {}
