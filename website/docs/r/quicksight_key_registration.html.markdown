---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_key_registration"
description: |-
  Registers customer managed keys in a Amazon QuickSight account.
---

# Resource: aws_quicksight_key_registration

Registers customer managed keys in a Amazon QuickSight account.

~> Deletion of this resource clears all CMK registrations from a QuickSight account. QuickSight then uses AWS owned keys to encrypt your resources.

## Example Usage

```terraform
resource "aws_quicksight_key_registration" "example" {
  key_registration {
    key_arn = aws_kms_key.example1.arn
  }

  key_registration {
    key_arn     = aws_kms_key.example2.arn
    default_key = true
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `key_registration` - (Required) Registered keys. See [key_registration](#key_registration).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### key_registration

* `default_key` - (Optional) Whether the key is set as the default key for encryption and decryption use.
* `key_arn` - (Required) ARN of the AWS KMS key that is registered for encryption and decryption use.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `aws_account_id` - The ID for the AWS account that contains the settings.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import QuickSight key registration using the AWS account ID. For example:

```terraform
import {
  to = aws_quicksight_key_registration.example
  id = "012345678901"
}
```

Using `terraform import`, import QuickSight key registration using the AWS account ID. For example:

```console
% terraform import aws_quicksight_key_registration.example "012345678901"
```
