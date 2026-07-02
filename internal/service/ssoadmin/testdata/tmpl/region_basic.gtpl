resource "aws_ssoadmin_region" "test" {
{{- template "region" }}
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  region_name  = "us-east-1"
}

data "aws_ssoadmin_instances" "test" {
{{- template "region" -}}
}
