---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_workload_identity"
description: |-
  Manages an AWS Bedrock AgentCore Workload Identity.
---

# Resource: aws_bedrockagentcore_workload_identity

Manages an AWS Bedrock AgentCore Workload Identity. Workload Identity provides OAuth2-based authentication and authorization for AI agents to access external resources securely.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagentcore_workload_identity" "example" {
  name = "example-workload-identity"
  allowed_resource_oauth2_return_urls = [
    "https://example.com/callback"
  ]
}
```

### Workload Identity with Multiple Return URLs

```terraform
resource "aws_bedrockagentcore_workload_identity" "example" {
  name = "example-workload-identity"
  allowed_resource_oauth2_return_urls = [
    "https://app.example.com/oauth/callback",
    "https://api.example.com/auth/return",
    "https://example.com/callback"
  ]
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the workload identity. Must be 3-255 characters and contain only alphanumeric characters, hyphens, periods, and underscores.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `allowed_resource_oauth2_return_urls` - (Optional) Set of allowed OAuth2 return URLs for resources associated with this workload identity. These URLs are used as valid redirect targets during OAuth2 authentication flows.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `workload_identity_arn` - ARN of the Workload Identity.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Workload Identity using the workload identity name. For example:

```terraform
import {
  to = aws_bedrockagentcore_workload_identity.example
  id = "example-workload-identity"
}
```

Using `terraform import`, import Bedrock AgentCore Workload Identity using the workload identity name. For example:

```console
% terraform import aws_bedrockagentcore_workload_identity.example example-workload-identity
```
