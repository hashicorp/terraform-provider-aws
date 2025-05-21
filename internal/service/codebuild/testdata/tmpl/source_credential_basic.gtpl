resource "aws_codebuild_source_credential" "test" {
{{- template "region" }}
  auth_type   = "PERSONAL_ACCESS_TOKEN"
  server_type = "GITHUB"
  token       = var.rName
}
