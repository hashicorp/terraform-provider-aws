resource "aws_rekognition_collection" "test" {
{{- template "region" }}
  collection_id = var.rName

{{- template "tags" . }}
}
