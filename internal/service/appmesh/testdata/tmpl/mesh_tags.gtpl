resource "aws_appmesh_mesh" "test" {
  name = var.rName

{{- template "tags" . }}
}
