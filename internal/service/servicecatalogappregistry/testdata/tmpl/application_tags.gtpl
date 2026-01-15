resource "aws_servicecatalogappregistry_application" "test" {
  name        = var.rName
  description = "Example Description"
{{- template "tags" . }}
}