---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_code_signing_config"
description: |-
  Manages an AWS Lambda Code Signing Config.
---

# Resource: aws_lambda_code_signing_config

Manages an AWS Lambda Code Signing Config. Use this resource to define allowed signing profiles and code-signing validation policies for Lambda functions to ensure code integrity and authenticity.

For information about Lambda code signing configurations and how to use them, see [configuring code signing for Lambda functions](https://docs.aws.amazon.com/lambda/latest/dg/configuration-codesigning.html).

## Example Usage

### Basic Usage

```terraform
# Create signing profiles for different environments
resource "aws_signer_signing_profile" "prod" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name_prefix = "prod_lambda_"

  tags = {
    Environment = "production"
  }
}

resource "aws_signer_signing_profile" "dev" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name_prefix = "dev_lambda_"

  tags = {
    Environment = "development"
  }
}

# Code signing configuration with enforcement
resource "aws_lambda_code_signing_config" "example" {
  description = "Code signing configuration for Lambda functions"

  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.prod.version_arn,
      aws_signer_signing_profile.dev.version_arn,
    ]
  }

  policies {
    untrusted_artifact_on_deployment = "Enforce" # Block deployments that fail code signing validation
  }

  tags = {
    Environment = "production"
    Purpose     = "code-signing"
  }
}
```

### Warning Only Configuration

```terraform
resource "aws_lambda_code_signing_config" "example" {
  description = "Development code signing configuration"

  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.dev.version_arn,
    ]
  }

  policies {
    untrusted_artifact_on_deployment = "Warn" # Allow deployments but log validation failures
  }

  tags = {
    Environment = "development"
    Purpose     = "code-signing"
  }
}
```

### Multiple Environment Configuration

```terraform
# Production signing configuration
resource "aws_lambda_code_signing_config" "prod" {
  description = "Production code signing configuration with strict enforcement"

  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.prod.version_arn,
    ]
  }

  policies {
    untrusted_artifact_on_deployment = "Enforce"
  }

  tags = {
    Environment = "production"
    Security    = "strict"
  }
}

# Development signing configuration
resource "aws_lambda_code_signing_config" "dev" {
  description = "Development code signing configuration with warnings"

  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.dev.version_arn,
      aws_signer_signing_profile.test.version_arn,
    ]
  }

  policies {
    untrusted_artifact_on_deployment = "Warn"
  }

  tags = {
    Environment = "development"
    Security    = "flexible"
  }
}
```

## Argument Reference

The following arguments are required:

* `allowed_publishers` - (Required) Configuration block of allowed publishers as signing profiles for this code signing configuration. [See below](#allowed_publishers-configuration-block).

The following arguments are optional:

* `description` - (Optional) Descriptive name for this code signing configuration.
* `policies` - (Optional) Configuration block of code signing policies that define the actions to take if the validation checks fail. [See below](#policies-configuration-block).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the object. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### allowed_publishers Configuration Block

* `signing_profile_version_arns` - (Required) Set of ARNs for each of the signing profiles. A signing profile defines a trusted user who can sign a code package. Maximum of 20 signing profiles.

### policies Configuration Block

* `untrusted_artifact_on_deployment` - (Required) Code signing configuration policy for deployment validation failure. If you set the policy to `Enforce`, Lambda blocks the deployment request if code-signing validation checks fail. If you set the policy to `Warn`, Lambda allows the deployment and creates a CloudWatch log. Valid values: `Warn`, `Enforce`. Default value: `Warn`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the code signing configuration.
* `config_id` - Unique identifier for the code signing configuration.
* `last_modified` - Date and time that the code signing configuration was last modified.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Code Signing Configs using their ARN. For example:

```terraform
import {
  to = aws_lambda_code_signing_config.example
  id = "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-0f6c334abcdea4d8b"
}
```

For backwards compatibility, the following legacy `terraform import` command is also supported:

```console
% terraform import aws_lambda_code_signing_config.example arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-0f6c334abcdea4d8b
```
