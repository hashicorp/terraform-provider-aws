data "aws_ssoadmin_instances" "test" {
{{- template "region" }}
}

resource "aws_accountaccess_application" "test" {
{{- template "region" }}
  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

{{- template "tags" . }}
}
