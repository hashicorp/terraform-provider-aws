resource "aws_key_pair" "test" {
{{- template "region" }}
  key_name   = var.rName
  public_key = var.public_key
{{- template "tags" . }}
}
