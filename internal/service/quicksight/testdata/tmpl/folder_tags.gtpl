resource "aws_quicksight_folder" "test" {
  folder_id = var.rName
  name      = var.rName
{{- template "tags" . }}
}
