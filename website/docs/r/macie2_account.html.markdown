---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_account"
description: |-
  Provides a resource to manage Amazon Macie on an AWS Account.
---

# Resource: aws_macie2_account

Provides a resource to manage an [AWS Macie Account](https://docs.aws.amazon.com/macie/latest/APIReference/macie.html).

## Example Usage

```terraform
resource "aws_macie2_account" "test" {
  finding_publishing_frequency = "FIFTEEN_MINUTES"
  status                       = "ENABLED"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `finding_publishing_frequency` -  (Optional) Specifies how often to publish updates to policy findings for the account. This includes publishing updates to AWS Security Hub and Amazon EventBridge (formerly called Amazon CloudWatch Events). Valid values are `FIFTEEN_MINUTES`, `ONE_HOUR` or `SIX_HOURS`.
* `status` - (Optional) Specifies the status for the account. To enable Amazon Macie and start all Macie activities for the account, set this value to `ENABLED`. Valid values are `ENABLED` or `PAUSED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The unique identifier (ID) of the macie account.
* `service_role` - The Amazon Resource Name (ARN) of the service-linked role that allows Macie to monitor and analyze data in AWS resources for the account.
* `created_at` - The date and time, in UTC and extended RFC 3339 format, when the Amazon Macie account was created.
* `updated_at` - The date and time, in UTC and extended RFC 3339 format, of the most recent change to the status of the Macie account.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_macie2_account` using the id. For example:

```terraform
import {
  to = aws_macie2_account.example
  id = "abcd1"
}
```

Using `terraform import`, import `aws_macie2_account` using the id. For example:

```console
% terraform import aws_macie2_account.example abcd1
```
