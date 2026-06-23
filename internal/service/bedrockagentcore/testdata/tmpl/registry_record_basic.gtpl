resource "aws_bedrockagentcore_registry_record" "test" {
{{- template "region" }}
  name        = "${var.rName}-record"
  registry_id = aws_bedrockagentcore_registry.test.registry_id

  descriptor_type = "CUSTOM"
}

resource "aws_bedrockagentcore_registry" "test" {
{{- template "region" }}
  name = "${var.rName}-registry"
}