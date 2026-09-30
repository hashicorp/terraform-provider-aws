---
subcategory: "SES Mail Manager"
layout: "aws"
page_title: "AWS: aws_mailmanager_relay"
description: |-
  Manages an AWS SES Mail Manager Relay.
---

# Resource: aws_mailmanager_relay

Manages an AWS SES Mail Manager Relay.

## Example Usage

### Basic Usage

```terraform
resource "aws_mailmanager_relay" "example" {
  name        = "example"
  server_name = "smtp.example.com"
  server_port = 25

  authentication {
    no_authentication {}
  }
}
```

### With Secret Authentication

```terraform
resource "aws_secretsmanager_secret" "example" {
  name = "example"
}

resource "aws_secretsmanager_secret_version" "example" {
  secret_id     = aws_secretsmanager_secret.example.id
  secret_string = jsonencode({ username = "user", password = "pass" })
}

resource "aws_mailmanager_relay" "example" {
  name        = "example"
  server_name = "smtp.example.com"
  server_port = 587

  authentication {
    secret_arn = aws_secretsmanager_secret_version.example.arn
  }
}
```

## Argument Reference

The following arguments are required:

* `authentication` - (Required) Authentication configuration for the relay. See [`authentication` Block](#authentication-block).
* `name` - (Required) Name of the relay.
* `server_name` - (Required) Hostname of the SMTP server.
* `server_port` - (Required) Port of the SMTP server.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `authentication` Block

Exactly one of the following must be configured:

* `no_authentication` - (Optional) No authentication is required to connect to the SMTP server.
* `secret_arn` - (Optional) ARN of the Secrets Manager secret containing the SMTP credentials.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the relay.
* `created_timestamp` - Timestamp when the relay was created.
* `id` - Identifier of the relay.
* `last_modified_timestamp` - Timestamp when the relay was last modified.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_mailmanager_relay.example
  identity = {
    id = "relay-id-12345678"
  }
}

resource "aws_mailmanager_relay" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) Identifier of the relay.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an SES Mail Manager Relay using its identifier. For example:

```terraform
import {
  to = aws_mailmanager_relay.example
  id = "relay-id-12345678"
}

resource "aws_mailmanager_relay" "example" {
  ### Configuration omitted for brevity ###
}
```

Using `terraform import`, import an SES Mail Manager Relay using its identifier. For example:

```console
% terraform import aws_mailmanager_relay.example relay-id-12345678
```
