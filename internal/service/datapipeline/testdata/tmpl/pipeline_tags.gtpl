resource "aws_datapipeline_pipeline" "test" {
  name = var.rName
{{- template "tags" . }}
}
