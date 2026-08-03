resource "aws_appfabric_app_bundle" "test" {
{{- template "region" -}}
{{- template "tags" . }}
}
