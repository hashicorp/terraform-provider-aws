resource "aws_bedrockagentcore_code_interpreter" "test" {
{{- template "region" }}
  name = var.rName

  network_configuration {
    network_mode = "SANDBOX"
  }

{{- template "tags" . }}
}
