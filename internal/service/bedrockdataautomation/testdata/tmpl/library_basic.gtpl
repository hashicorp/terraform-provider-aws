resource "aws_bedrockdataautomation_library" "test" {
{{- template "region" }}
  name = var.rName

{{- template "tags" . }}
}
