resource "aws_ssoadmin_application_assignment_configuration" "test" {
{{- template "region" }}
  application_arn     = aws_ssoadmin_application.test.application_arn
  assignment_required = true
}

resource "aws_ssoadmin_application" "test" {
{{- template "region" }}
  name                     = var.rName
  application_provider_arn = local.test_application_provider_arn
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

data "aws_ssoadmin_instances" "test" {
{{- template "region" -}}
}

locals {
  test_application_provider_arn = "arn:aws:sso::aws:applicationProvider/custom"
}
