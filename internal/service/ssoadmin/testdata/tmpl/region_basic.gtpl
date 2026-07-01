data "aws_ssoadmin_instances" "test" {
{{- template "region" -}}
}

resource "aws_ssoadmin_region" "test" {
{{- template "region" }}
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
{{- if .WithRegion }}
  region_name  = var.region
{{- else }}
  region_name  = "us-west-2"
{{- end }}
}
