resource "aws_bedrockagentcore_browser" "test" {
{{- template "region" }}
  name = var.rName

  network_configuration {
    network_mode = "PUBLIC"
  }

{{- template "tags" . }}
}
