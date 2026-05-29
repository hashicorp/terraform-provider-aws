data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_region" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
{{- template "region" -}}
  region_name = "us-west-2"
}
