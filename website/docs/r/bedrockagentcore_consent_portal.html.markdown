---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_consent_portal"
description: |-
  Manages an AWS Bedrock AgentCore Consent Portal.
---

# Resource: aws_bedrockagentcore_consent_portal

Manages an AWS Bedrock AgentCore Consent Portal. Consent portals authenticate end users with an OpenID Connect (OIDC) identity provider and collect consent before an agent accesses downstream resources on their behalf.

-> **Note:** The gateway must use a JWT authorizer with the same OIDC issuer as the OAuth2 credential provider. See the [consent portal prerequisites](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/identity-consent-portal-prerequisites.html) and [execution role requirements](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/identity-consent-portal-execution-role.html).

## Example Usage

```terraform
resource "aws_bedrockagentcore_consent_portal" "example" {
  name               = "example"
  execution_role_arn = aws_iam_role.example.arn

  idp_config {
    credential_provider_arn = aws_bedrockagentcore_oauth2_credential_provider.example.credential_provider_arn
    scopes                  = ["openid"]
  }

  sources {
    identifier = aws_bedrockagentcore_gateway.example.gateway_id
    type       = "agentcore-gateway"
  }
}
```

## Argument Reference

The following arguments are required:

* `execution_role_arn` - (Required) ARN of the IAM role assumed by the consent portal.
* `idp_config` - (Required) Identity provider configuration. See [`idp_config` Block](#idp_config-block) below.
* `name` - (Required) Name of the consent portal. Changing this value forces replacement.
* `sources` - (Required) Source served by the consent portal. Exactly one source is supported. Changing this value forces replacement. See [`sources` Block](#sources-block) below.

The following arguments are optional:

* `description` - (Optional) Description of the consent portal. Must contain between 1 and 4096 characters.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `idp_config` Block

The following arguments are required:

* `credential_provider_arn` - (Required) ARN of the OAuth2 credential provider used to authenticate end users.
* `scopes` - (Required) Set of OAuth2 scopes requested during authentication. At least one scope is required. Each scope must contain between 1 and 255 characters. Include `openid`.
The following arguments are optional:

* `audience` - (Optional) Audience included in token requests to the identity provider. Must not be empty. Omit this argument to clear the audience.

### `sources` Block

The `sources` block supports:

* `identifier` - (Required) ID or ARN of the gateway served by the consent portal.
* `type` - (Required) Type of source. Valid value is `agentcore-gateway`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `consent_portal_arn` - ARN of the consent portal.
* `consent_portal_id` - ID of the consent portal.
* `portal_url` - URL used to access the consent portal.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute to import Bedrock AgentCore Consent Portals. For example:

```terraform
import {
  to = aws_bedrockagentcore_consent_portal.example
  identity = {
    consent_portal_id = "example-1234567890"
  }
}
```

### Identity Schema

#### Required

* `consent_portal_id` (String) ID of the consent portal.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Consent Portals using `consent_portal_id`. For example:

```terraform
import {
  to = aws_bedrockagentcore_consent_portal.example
  id = "example-1234567890"
}
```

Using `terraform import`, import Bedrock AgentCore Consent Portals using `consent_portal_id`. For example:

```console
% terraform import aws_bedrockagentcore_consent_portal.example example-1234567890
```
