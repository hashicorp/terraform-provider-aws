---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_consent_portal"
description: |-
  Gets information about an AWS Bedrock AgentCore Consent Portal.
---

# Data Source: aws_bedrockagentcore_consent_portal

Gets information about an AWS Bedrock AgentCore Consent Portal.

## Example Usage

```terraform
data "aws_bedrockagentcore_consent_portal" "example" {
  consent_portal_identifier = "example-1234567890"
}
```

## Argument Reference

This data source supports the following arguments:

* `consent_portal_identifier` - (Required) ID or ARN of the consent portal.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `consent_portal_arn` - ARN of the consent portal.
* `consent_portal_id` - ID of the consent portal.
* `description` - Description of the consent portal.
* `execution_role_arn` - ARN of the IAM role assumed by the consent portal.
* `idp_config` - Identity provider configuration. See [`idp_config` Block](#idp_config-block) below.
* `name` - Name of the consent portal.
* `portal_url` - URL used to access the consent portal.
* `sources` - Sources served by the consent portal. See [`sources` Block](#sources-block) below.
* `tags` - Map of tags assigned to the resource.

### `idp_config` Block

The `idp_config` attribute exports:

* `audience` - Audience included in token requests to the identity provider.
* `credential_provider_arn` - ARN of the OAuth2 credential provider used to authenticate end users.
* `scopes` - Set of OAuth2 scopes requested during authentication.

### `sources` Block

The `sources` attribute exports:

* `identifier` - Identifier of the source resource.
* `type` - Type of source resource.
