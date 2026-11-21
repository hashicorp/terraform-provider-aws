---
subcategory: "Meta Data Sources"
layout: "aws"
page_title: "AWS: aws_credentials"
description: |-
  Retrieve the current AWS SDK session credentials.
---

# Ephemeral: aws_credentials

Retrieve the current AWS SDK session credentials configured in the provider.

~> **NOTE:** Ephemeral resources are a new feature and may evolve as we continue to explore their most effective uses. [Learn more](https://developer.hashicorp.com/terraform/language/resources/ephemeral).

## Example Usage

### Basic Usage

```terraform
ephemeral "aws_credentials" "example" {}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `access_key_id` - AWS access key ID for the current session credentials.
* `secret_access_key` - AWS secret access key for the current session credentials.
* `session_token` - AWS session token for the current session credentials. Empty if the credentials are not temporary.
