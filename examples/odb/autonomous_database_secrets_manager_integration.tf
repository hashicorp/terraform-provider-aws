// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

# Enables the Oracle Database@AWS Autonomous Database Serverless integration
# with AWS Secrets Manager and returns the service role ARN to trust from a
# customer-managed IAM role.
resource "aws_odb_autonomous_database_secrets_manager_integration" "example" {}

output "autonomous_database_secrets_manager_service_role_arn" {
  value = aws_odb_autonomous_database_secrets_manager_integration.example.role_arn
}
