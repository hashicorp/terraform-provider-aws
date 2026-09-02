data "aws_ssoadmin_instances" "test" {
{{- template "region" }}
}

resource "aws_accountaccess_application" "test" {
{{- template "region" }}
  identity_source {
    identity_center {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
    }
  }

{{- template "tags" . }}
}
