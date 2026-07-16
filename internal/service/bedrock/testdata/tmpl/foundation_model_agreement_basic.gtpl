data "aws_bedrock_foundation_model_agreement_offers" "test" {
{{- template "region" }}
  model_id   = var.AWS_BEDROCK_FOUNDATION_MODEL_ID
  offer_type = "PUBLIC"
}

resource "aws_bedrock_foundation_model_agreement" "test" {
{{- template "region" }}
  model_id    = var.AWS_BEDROCK_FOUNDATION_MODEL_ID
  offer_token = data.aws_bedrock_foundation_model_agreement_offers.test.offers[0].offer_token

  lifecycle {
    ignore_changes = [offer_token]
  }
}
