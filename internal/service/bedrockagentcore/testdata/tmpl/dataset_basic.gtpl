resource "aws_bedrockagentcore_dataset" "test" {
{{- template "region" }}
  name        = var.rName
  schema_type = "AGENTCORE_EVALUATION_PREDEFINED_V1"

  source {
    inline_examples {
      examples = [
        jsonencode({
          scenario_id = "scenario-1"
          turns = [
            { input = "What is 2+2?", expected_response = "4" }
          ]
        })
      ]
    }
  }
{{- template "tags" . }}
}
