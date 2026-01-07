# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

plugin "aws" {
  enabled = true
  version = "0.41.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

# Terraform Rules
# https://github.com/terraform-linters/tflint-ruleset-terraform/blob/main/docs/rules/README.md

rule "terraform_comment_syntax" {
  enabled = true
}

rule "terraform_required_providers" {
  enabled = false
}

rule "terraform_required_version" {
  enabled = false
}

# AWS Rules
# https://github.com/terraform-linters/tflint-ruleset-aws/blob/master/docs/rules/README.md

rule "aws_acm_certificate_lifecycle" {
  enabled = false
}

# Rule needs to be disabled due to enum value case inconsistencies
rule "aws_dms_s3_endpoint_invalid_compression_type" {
  enabled = false
}

# Rule needs to be disabled due to enum value case inconsistencies
rule "aws_dms_s3_endpoint_invalid_date_partition_sequence" {
  enabled = false
}

# Rule needs to be disabled due to enum value case inconsistencies
rule "aws_dms_s3_endpoint_invalid_encryption_mode" {
  enabled = false
}

# Avoids errant findings related to directory paths in generated configuration files
rule "aws_iam_saml_provider_invalid_saml_metadata_document" {
  enabled = false
}

# Rule needs to be disabled due to bad email regex in the linter rule
rule "aws_guardduty_member_invalid_email" {
  enabled = false
}

rule "aws_api_gateway_domain_name_invalid_security_policy" {
  enabled = false
}
