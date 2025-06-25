---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_code_signing_config"
description: |-
  Provides details about an AWS Lambda Code Signing Config.
---

# Data Source: aws_lambda_code_signing_config

Provides details about an AWS Lambda Code Signing Config. Use this data source to retrieve information about an existing code signing configuration for Lambda functions to ensure code integrity and authenticity.

For information about Lambda code signing configurations and how to use them, see [configuring code signing for Lambda functions](https://docs.aws.amazon.com/lambda/latest/dg/configuration-codesigning.html).

## Example Usage

### Basic Usage

```terraform
data "aws_lambda_code_signing_config" "example" {
  arn = "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-0f6c334abcdea4d8b"
}

output "config_details" {
  value = {
    config_id   = data.aws_lambda_code_signing_config.example.config_id
    description = data.aws_lambda_code_signing_config.example.description
    policy      = data.aws_lambda_code_signing_config.example.policies[0].untrusted_artifact_on_deployment
  }
}
```

### Use in Lambda Function

```terraform
# Get existing code signing configuration
data "aws_lambda_code_signing_config" "security_config" {
  arn = var.code_signing_config_arn
}

# Create Lambda function with code signing
resource "aws_lambda_function" "example" {
  filename                = "function.zip"
  function_name           = "secure-function"
  role                    = aws_iam_role.lambda_role.arn
  handler                 = "index.handler"
  runtime                 = "nodejs20.x"
  code_signing_config_arn = data.aws_lambda_code_signing_config.security_config.arn

  tags = {
    Environment = "production"
    Security    = "code-signed"
  }
}
```

### Validate Signing Profiles

```terraform
data "aws_lambda_code_signing_config" "example" {
  arn = var.code_signing_config_arn
}

# Check if specific signing profile is allowed
locals {
  allowed_profiles = data.aws_lambda_code_signing_config.example.allowed_publishers[0].signing_profile_version_arns
  required_profile = "arn:aws:signer:us-west-2:123456789012:/signing-profiles/MyProfile"
  profile_allowed  = contains(local.allowed_profiles, local.required_profile)
}

# Conditional resource creation based on signing profile validation
resource "aws_lambda_function" "conditional" {
  count = local.profile_allowed ? 1 : 0

  filename                = "function.zip"
  function_name           = "conditional-function"
  role                    = aws_iam_role.lambda_role.arn
  handler                 = "index.handler"
  runtime                 = "python3.12"
  code_signing_config_arn = data.aws_lambda_code_signing_config.example.arn
}

output "deployment_status" {
  value = {
    profile_allowed  = local.profile_allowed
    function_created = local.profile_allowed
    message          = local.profile_allowed ? "Function deployed with valid signing profile" : "Deployment blocked - signing profile not allowed"
  }
}
```

### Multi-Environment Configuration

```terraform
# Production code signing config
data "aws_lambda_code_signing_config" "prod" {
  arn = "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-prod-123"
}

# Development code signing config
data "aws_lambda_code_signing_config" "dev" {
  arn = "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-dev-456"
}

# Compare configurations
locals {
  prod_policy = data.aws_lambda_code_signing_config.prod.policies[0].untrusted_artifact_on_deployment
  dev_policy  = data.aws_lambda_code_signing_config.dev.policies[0].untrusted_artifact_on_deployment

  config_comparison = {
    prod_enforcement = local.prod_policy
    dev_enforcement  = local.dev_policy
    policies_match   = local.prod_policy == local.dev_policy
  }
}

output "environment_comparison" {
  value = local.config_comparison
}
```

## Argument Reference

The following arguments are required:

* `arn` - (Required) ARN of the code signing configuration.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `allowed_publishers` - List of allowed publishers as signing profiles for this code signing configuration. [See below](#allowed_publishers-attribute-reference).
* `config_id` - Unique identifier for the code signing configuration.
* `description` - Code signing configuration description.
* `last_modified` - Date and time that the code signing configuration was last modified.
* `policies` - List of code signing policies that control the validation failure action for signature mismatch or expiry. [See below](#policies-attribute-reference).

### allowed_publishers Attribute Reference

* `signing_profile_version_arns` - Set of ARNs for each of the signing profiles. A signing profile defines a trusted user who can sign a code package.

### policies Attribute Reference

* `untrusted_artifact_on_deployment` - Code signing configuration policy for deployment validation failure. Valid values: `Warn`, `Enforce`.
