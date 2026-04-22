---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_account_settings"
description: |-
  Terraform resource for managing an AWS QuickSight Account Settings.
---

# Resource: aws_quicksight_account_settings

Terraform resource for managing an AWS QuickSight Account Settings.

~> Deletion of this resource will not modify any settings, only remove the resource from state.

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

This resource supports the following arguments:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID. Defaults to automatically determined account ID of the Terraform AWS provider.
* `default_namespace` - (Optional) The default namespace for this Amazon Web Services account. Currently, the default is `default`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `termination_protection_enabled` - (Optional) A boolean value that determines whether or not an Amazon QuickSight account can be deleted. If `true`, it does not allow the account to be deleted and results in an error message if a user tries to make a DeleteAccountSubscription request. If `false`, it will allow the account to be deleted.

## Attribute Reference

This resource exports no additional attributes.

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
