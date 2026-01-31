data "aws_ssoadmin_instances" "test" {}
data "aws_region" "current" {}

resource "aws_mpa_identity_source" "test" {
  name = var.rName

  identity_source_parameters {
    iam_identity_center {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      region       = data.aws_region.current.name
    }
  }
{{- template "tags" . }}
}
