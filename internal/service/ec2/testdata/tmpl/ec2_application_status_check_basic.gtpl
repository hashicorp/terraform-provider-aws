resource "aws_ec2_application_status_check" "test" {
{{- template "region" }}
  protocol = "http"
  port     = 80

{{- template "tags" . }}
}
