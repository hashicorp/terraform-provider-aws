resource "aws_sesv2_contact_list" "test" {
  contact_list_name = var.rName
{{- template "tags" . }}
}
