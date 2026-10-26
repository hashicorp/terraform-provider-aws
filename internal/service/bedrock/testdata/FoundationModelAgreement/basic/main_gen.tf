# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_bedrock_foundation_model_agreement_offers" "test" {
  model_id   = var.AWS_BEDROCK_FOUNDATION_MODEL_ID
  offer_type = "PUBLIC"
}

resource "aws_bedrock_foundation_model_agreement" "test" {
  model_id    = var.AWS_BEDROCK_FOUNDATION_MODEL_ID
  offer_token = data.aws_bedrock_foundation_model_agreement_offers.test.offers[0].offer_token

  lifecycle {
    ignore_changes = [offer_token]
  }
}


variable "AWS_BEDROCK_FOUNDATION_MODEL_ID" {
  type     = string
  nullable = false
}
