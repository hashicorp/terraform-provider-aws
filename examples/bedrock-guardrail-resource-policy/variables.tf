# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

variable "region" {
  description = "AWS region in which to create the guardrail. This should be the management account's home region."
  type        = string
  default     = "us-east-1"
}

variable "guardrail_name" {
  description = "Name for the Bedrock guardrail."
  type        = string
  default     = "org-guardrail"
}

variable "guardrail_profile_arn" {
  description = <<-EOT
    ARN of the system-defined Bedrock guardrail profile to attach a resource-based
    policy to. Only required when Cross-Region Inference (CRIS) is enabled for the
    organisation-level guardrail enforcement policy (BEDROCK_POLICY).

    AWS defines these ARNs; they are not created by Terraform. Leave empty to skip
    creating the profile policy.

    Retrieve available profile ARNs with:
    aws bedrock list-foundation-model-availability --query 'guardrailProfiles[*].guardrailProfileArn'

    Example: arn:aws:bedrock:us-east-1::guardrail-profile/system-defined-profile-id
  EOT
  type    = string
  default = "arn:aws:bedrock:us-east-1:271376211545:guardrail-profile/us.guardrail.v1:0"
}
