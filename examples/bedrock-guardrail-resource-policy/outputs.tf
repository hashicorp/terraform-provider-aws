# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

output "guardrail_arn" {
  description = "ARN of the Bedrock guardrail. Reference this in the BEDROCK_POLICY service control policy."
  value       = aws_bedrock_guardrail.this.guardrail_arn
}

output "guardrail_id" {
  description = "ID of the Bedrock guardrail."
  value       = aws_bedrock_guardrail.this.guardrail_id
}

output "organization_id" {
  description = "AWS Organizations ID used to scope the resource-based policies."
  value       = data.aws_organizations_organization.org.id
}
