# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
}

provider "aws" {
  region = var.region
}

# ---------------------------------------------------------------------------
# Guardrail
# ---------------------------------------------------------------------------

resource "aws_bedrock_guardrail" "this" {
  name                      = var.guardrail_name
  blocked_input_messaging   = "This request has been blocked by your organisation's content policy."
  blocked_outputs_messaging = "This response has been blocked by your organisation's content policy."
  description               = "Organisation-level guardrail managed centrally and enforced via AWS Organizations."

  content_policy_config {
    filters_config {
      type            = "HATE"
      input_strength  = "HIGH"
      output_strength = "HIGH"
    }
    filters_config {
      type            = "VIOLENCE"
      input_strength  = "HIGH"
      output_strength = "HIGH"
    }
    filters_config {
      type            = "SEXUAL"
      input_strength  = "HIGH"
      output_strength = "HIGH"
    }
    filters_config {
      type            = "MISCONDUCT"
      input_strength  = "HIGH"
      output_strength = "HIGH"
    }
    filters_config {
      type            = "PROMPT_ATTACK"
      input_strength  = "HIGH"
      output_strength = "NONE"
    }
  }

  sensitive_information_policy_config {
    pii_entities_config {
      type   = "EMAIL"
      action = "ANONYMIZE"
    }
    pii_entities_config {
      type   = "PHONE"
      action = "ANONYMIZE"
    }
    pii_entities_config {
      type   = "AWS_ACCESS_KEY"
      action = "BLOCK"
    }
    pii_entities_config {
      type   = "AWS_SECRET_KEY"
      action = "BLOCK"
    }
  }
}

# ---------------------------------------------------------------------------
# AWS Organizations — used to scope the resource-based policy to the org
# ---------------------------------------------------------------------------

data "aws_organizations_organization" "org" {}

# ---------------------------------------------------------------------------
# Resource-based policy on the guardrail
#
# Grants every principal in the organisation permission to call
# bedrock:GetGuardrail and bedrock:ApplyGuardrail on this guardrail.
# This is a prerequisite for the BEDROCK_POLICY service control policy
# in AWS Organizations to take effect in member accounts.
# ---------------------------------------------------------------------------

resource "aws_bedrock_guardrail_resource_policy" "guardrail" {
  resource_arn = aws_bedrock_guardrail.this.guardrail_arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = "*"
      Action    = ["bedrock:GetGuardrail", "bedrock:ApplyGuardrail"]
      Resource  = aws_bedrock_guardrail.this.guardrail_arn
      Condition = {
        StringEquals = { "aws:PrincipalOrgID" = data.aws_organizations_organization.org.id }
      }
    }]
  })
}

# ---------------------------------------------------------------------------
# Resource-based policy on a system-defined guardrail profile
#
# Required when Cross-Region Inference (CRIS) is enabled.  The guardrail
# profile ARN is system-defined by AWS and must be supplied as a variable.
# Set var.guardrail_profile_arn to enable this block.
# ---------------------------------------------------------------------------

resource "aws_bedrock_guardrail_resource_policy" "profile" {
  count = var.guardrail_profile_arn != "" ? 1 : 0

  resource_arn = var.guardrail_profile_arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = "*"
      Action    = ["bedrock:ApplyGuardrail"]
      Resource  = var.guardrail_profile_arn
      Condition = {
        StringEquals = { "aws:PrincipalOrgID" = data.aws_organizations_organization.org.id }
      }
    }]
  })
}
