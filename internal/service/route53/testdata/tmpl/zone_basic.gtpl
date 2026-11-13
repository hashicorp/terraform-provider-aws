resource "aws_route53_zone" "test" {
{{- template "region" }}
  comment = var.rName
  name    = "${var.zoneName}."
{{- template "tags" . }}
}
