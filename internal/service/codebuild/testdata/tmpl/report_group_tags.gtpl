resource "aws_codebuild_report_group" "test" {
{{- template "region" }}
  name = var.rName
  type = "TEST"

  export_config {
    type = "NO_EXPORT"
  }
{{- template "tags" . }}
}
