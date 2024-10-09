---
subcategory: "SFN (Step Functions)"
layout: "aws"
page_title: "AWS: aws_sfn_activity"
description: |-
  Provides a Step Function Activity resource.
---

# Resource: aws_sfn_activity

Provides a Step Function Activity resource

## Example Usage

### Basic

```terraform
resource "aws_sfn_activity" "sfn_activity" {
  name = "my-activity"
}
```

### Encryption

~> *NOTE:* See the section [Data at rest encyption](https://docs.aws.amazon.com/step-functions/latest/dg/encryption-at-rest.html) in the [AWS Step Functions Developer Guide](https://docs.aws.amazon.com/step-functions/latest/dg/welcome.html) for more information about enabling encryption of data using a customer-managed key for Step Functions State Machines data.

```terraform
resource "aws_sfn_activity" "sfn_activity" {
  name = "my-activity"

  encryption_configuration {
    kms_key_id                        = aws_kms_key.kms_key_for_sfn.arn
    type                              = "CUSTOMER_MANAGED_KMS_KEY"
    kms_data_key_reuse_period_seconds = 900
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `encryption_configuration` - (Optional) Defines what encryption configuration is used to encrypt data in the Activity. For more information see the section [Data at rest encyption](https://docs.aws.amazon.com/step-functions/latest/dg/encryption-at-rest.html) in the AWS Step Functions User Guide.
* `name` - (Required) The name of the activity to create.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `encryption_configuration` Configuration Block

* `kms_key_id` - (Optional) The alias, alias ARN, key ID, or key ARN of the symmetric encryption KMS key that encrypts the data key. To specify a KMS key in a different AWS account, the customer must use the key ARN or alias ARN. For more information regarding kms_key_id, see [KeyId](https://docs.aws.amazon.com/kms/latest/APIReference/API_DescribeKey.html#API_DescribeKey_RequestParameters) in the KMS documentation.
* `type` - (Required) The encryption option specified for the activity. Valid values: `AWS_KMS_KEY`, `CUSTOMER_MANAGED_KMS_KEY`
* `kms_data_key_reuse_period_seconds` - (Optional) Maximum duration for which Activities will reuse data keys. When the period expires, Activities will call GenerateDataKey. This setting only applies to customer managed KMS key and does not apply to AWS owned KMS key.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) that identifies the created activity.
* `name` - The name of the activity.
* `creation_date` - The date the activity was created.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import activities using the `arn`. For example:

```terraform
import {
  to = aws_sfn_activity.foo
  id = "arn:aws:states:eu-west-1:123456789098:activity:bar"
}
```

Using `terraform import`, import activities using the `arn`. For example:

```console
% terraform import aws_sfn_activity.foo arn:aws:states:eu-west-1:123456789098:activity:bar
```
