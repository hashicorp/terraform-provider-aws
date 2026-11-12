---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_account_default_protect_configuration"
description: |-
  Manages the AWS End User Messaging SMS account default protect configuration.
---

# Resource: aws_pinpointsmsvoicev2_account_default_protect_configuration

Manages the AWS End User Messaging SMS account default protect configuration. AWS applies the account default's country rules to every message that doesn't specify another protect configuration, including messages sent through Amazon SNS and Amazon Cognito.

~> **Note:** An account has one default protect configuration per Region. Creating this resource replaces any existing account default, and destroying it leaves the account without one.

## Example Usage

```terraform
resource "aws_pinpointsmsvoicev2_protect_configuration" "example" {
  sms_country_rules = {
    RU = "BLOCK"
  }
}

resource "aws_pinpointsmsvoicev2_account_default_protect_configuration" "example" {
  protect_configuration_id = aws_pinpointsmsvoicev2_protect_configuration.example.id
}
```

## Argument Reference

The following arguments are required:

* `protect_configuration_id` - (Required) ID of the protect configuration.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

~> **Note:** The `account_default` attribute of `aws_pinpointsmsvoicev2_protect_configuration` changes when this resource is applied, and Terraform shows the new value after the next refresh.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - (**Deprecated**, use `region` instead) Region where the account default is managed.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_account_default_protect_configuration.example
  identity = {
    region = "us-west-2"
  }
}

resource "aws_pinpointsmsvoicev2_account_default_protect_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the AWS End User Messaging SMS account default protect configuration using the Region. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_account_default_protect_configuration.example
  id = "us-west-2"
}
```

Using `terraform import`, import the AWS End User Messaging SMS account default protect configuration using the Region. For example:

```console
% terraform import aws_pinpointsmsvoicev2_account_default_protect_configuration.example us-west-2
```
