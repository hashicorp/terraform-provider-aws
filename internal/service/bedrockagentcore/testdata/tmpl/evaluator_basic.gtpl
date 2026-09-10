resource "aws_bedrockagentcore_evaluator" "test" {
{{- template "region" }}
  evaluator_name = var.rName
  level          = "TRACE"

  evaluator_config {
    llm_as_a_judge {
      instructions = "Given the {context} and the {assistant_turn}, compare against {expected_response} and rate from 1 to 5."

      rating_scale {
        numerical {
          definition = "Not helpful at all."
          value      = 1
          label      = "1"
        }
        numerical {
          definition = "Extremely helpful."
          value      = 5
          label      = "5"
        }
      }

      model_config {
        bedrock_evaluator_model_config {
          model_id = "us.amazon.nova-2-lite-v1:0"
        }
      }
    }
  }
{{- template "tags" . }}
}
