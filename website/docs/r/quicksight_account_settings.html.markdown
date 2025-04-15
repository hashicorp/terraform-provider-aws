---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_account_settings"
description: |-
  Terraform resource for managing an AWS QuickSight Account Settings.
---

# Resource: aws_quicksight_account_settings

Terraform resource for managing an AWS QuickSight Account Settings.

~> Due to the collision of the `notification_email` attribute of the [`aws_quicksight_account_subscription`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/quicksight_account_subscription) resource and the [`UpdateAccountSettings`](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_UpdateAccountSettings.html) API does not support updating the default namespace, this resource does not support their management.

## Example Usage

```terraform
resource "aws_quicksight_account_subscription" "subscription" {
  account_name          = "quicksight-terraform"
  authentication_method = "IAM_AND_QUICKSIGHT"
  edition               = "ENTERPRISE"
  notification_email    = "notification@email.com"
}

resource "aws_quicksight_account_settings" "example" {
  termination_protection_enabled = false

  depends_on = [aws_quicksight_account_subscription.subscription]
}
```

## Argument Reference

The following arguments are required:

* `termination_protection_enabled` - (Optional) A boolean value that determines whether or not an Amazon QuickSight account can be deleted. If `true`, it does not allow the account to be deleted and results in an error message if a user tries to make a DeleteAccountSubscription request. If `false`, it will allow the account to be deleted.
* `reset_on_delete` - (Optional) If `true`, destroying the resource will reset the TerminationProtectionEnabled account settings to default of `true`, otherwise account settings are not modified.
  Defaults to `false`.
  Will be removed in a future major version of the provider.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `default_namespace` - The default namespace for this Amazon Web Services account. Currently, the default is default.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import QuickSight Account Settings using the AWS account ID. For example:

```terraform
import {
  to = aws_quicksight_account_settings.example
  id = "012345678901"
}
```

Using `terraform import`, import QuickSight Account Settings using the AWS account ID. For example:

```console
% terraform import aws_quicksight_account_settings.example "012345678901"
```
