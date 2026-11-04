resource "aws_bedrockagentcore_memory_strategy" "test" {
{{- template "region" }}
  name                = var.rName
  memory_id           = aws_bedrockagentcore_memory.test.id
  type                = "SEMANTIC"
  namespace_templates = ["default"]
}

resource "aws_bedrockagentcore_memory" "test" {
{{- template "region" }}
  name                  = "${var.rName}_m"
  event_expiry_duration = 7
}