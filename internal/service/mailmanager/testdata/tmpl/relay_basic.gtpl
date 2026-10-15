resource "aws_mailmanager_relay" "test" {
{{- template "region" }}
  name        = var.rName
  server_name = "smtp.example.com"
  server_port = 25

  authentication {
    no_authentication {}
  }
{{- template "tags" . }}
}
