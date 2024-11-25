resource "aws_bedrock_guardrail" "test" {
  name                      = var.rName
  blocked_input_messaging   = "test"
  blocked_outputs_messaging = "test"
  description               = "test"

  content_policy_config {
    filters_config {
      input_strength  = "HIGH"
      output_strength = "HIGH"
      type            = "VIOLENCE"
    }
#    filters_config {
#      input_strength  = "HIGH"
#      output_strength = "HIGH"
#      type            = "VIOLENCE"
#    }
  }

#  word_policy_config {
#    managed_word_lists_config {
#      type = "PROFANITY"
#    }
#    words_config {
#      text = "HATE"
#    }
#  }

{{- template "tags" . }}
}
