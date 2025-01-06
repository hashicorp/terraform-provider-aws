resource "aws_bedrock_inference_profile" "test" {
  name = var.rName

  model_source {
    copy_from = "arn:aws:bedrock:eu-central-1::foundation-model/anthropic.claude-3-5-sonnet-20240620-v1:0"
  }

{{- template "tags" . }}
}
