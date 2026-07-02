---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_browser_profile"
description: |-
  Manages an AWS Bedrock AgentCore Browser Profile.
---

# Resource: aws_bedrockagentcore_browser_profile

Manages an AWS Bedrock AgentCore Browser Profile. Browser profiles define browser state that can be re-used across different browser sessions within AgentCore Browser. Browser state includes cookies and local storage.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagentcore_browser_profile" "example" {
  name = "example"
}
```

### With Description

```terraform
resource "aws_bedrockagentcore_browser_profile" "example" {
  name        = "example"
  description = "Example browser profile for web data extraction"
}
```

### With Tags

```terraform
resource "aws_bedrockagentcore_browser_profile" "example" {
  name        = "example"
  description = "Browser profile with tags"

  tags = {
    Environment = "production"
    Team        = "data-engineering"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the browser profile. Must start with a letter and can contain alphanumeric characters and underscores, up to 48 characters.

The following arguments are optional:

* `description` - (Optional) Description of the browser profile. Must be between 1 and 4096 characters.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `profile_arn` - ARN of the Browser Profile.
* `profile_id` - Unique identifier of the Browser Profile.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_browser_profile.example
  identity = {
    profile_id = "browser-profile-id-12345678"
  }
}

resource "aws_bedrockagentcore_browser_profile" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `profile_id` (String) ID of the browser profile.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Browser Profiles using `profile_id`. For example:

```terraform
import {
  to = aws_bedrockagentcore_browser_profile.example
  id = "browser-profile-id-12345678"
}
```

Using `terraform import`, import Bedrock AgentCore Browser Profiles using `profile_id`. For example:

```console
% terraform import aws_bedrockagentcore_browser_profile.example browser-profile-id-12345678
```
