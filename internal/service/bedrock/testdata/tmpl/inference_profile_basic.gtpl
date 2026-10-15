resource "aws_bedrock_inference_profile" "test" {
  name = var.rName

  model_source {
    copy_from = data.aws_bedrock_foundation_model.test.model_arn
  }

{{- template "tags" . }}
}

data "aws_bedrock_foundation_model" "test" {
  model_id = "amazon.nova-lite-v1:0"
}
