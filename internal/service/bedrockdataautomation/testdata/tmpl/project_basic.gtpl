resource "aws_bedrockdataautomation_project" "test" {
{{- template "region" }}
  name = var.rName

{{- template "tags" . }}
}
