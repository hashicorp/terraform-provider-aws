---
subcategory: "KMS (Key Management)"
layout: "aws"
page_title: "AWS: aws_kms_key"
description: |-
  Manages a single-Region or multi-Region primary KMS key.
---

# Resource: aws_kms_key

Manages a single-Region or multi-Region primary KMS key.

## Example Usage

```terraform
resource "aws_kms_key" "a" {
  description             = "KMS key 1"
  deletion_window_in_days = 10
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) The description of the key as viewed in AWS console.
* `key_usage` - (Optional) Specifies the intended use of the key. Valid values: `ENCRYPT_DECRYPT` or `SIGN_VERIFY`.
Defaults to `ENCRYPT_DECRYPT`.
* `customer_master_key_spec` - (Optional) Specifies whether the key contains a symmetric key or an asymmetric key pair and the encryption algorithms or signing algorithms that the key supports.
Valid values: `SYMMETRIC_DEFAULT`,  `RSA_2048`, `RSA_3072`, `RSA_4096`, `HMAC_256`, `ECC_NIST_P256`, `ECC_NIST_P384`, `ECC_NIST_P521`, or `ECC_SECG_P256K1`. Defaults to `SYMMETRIC_DEFAULT`. For help with choosing a key spec, see the [AWS KMS Developer Guide](https://docs.aws.amazon.com/kms/latest/developerguide/symm-asymm-choose.html).
* `policy` - (Optional) A valid policy JSON document. Although this is a key policy, not an IAM policy, an [`aws_iam_policy_document`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document), in the form that designates a principal, can be used. For more information about building policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).

~> **NOTE:** Note: All KMS keys must have a key policy. If a key policy is not specified, AWS gives the KMS key a [default key policy](https://docs.aws.amazon.com/kms/latest/developerguide/key-policies.html#key-policy-default) that gives all principals in the owning account unlimited access to all KMS operations for the key. This default key policy effectively delegates all access control to IAM policies and KMS grants.

* `bypass_policy_lockout_safety_check` - (Optional) A flag to indicate whether to bypass the key policy lockout safety check.
Setting this value to true increases the risk that the KMS key becomes unmanageable. Do not set this value to true indiscriminately.
For more information, refer to the scenario in the [Default Key Policy](https://docs.aws.amazon.com/kms/latest/developerguide/key-policies.html#key-policy-default-allow-root-enable-iam) section in the _AWS Key Management Service Developer Guide_.
The default value is `false`.
* `deletion_window_in_days` - (Optional) The waiting period, specified in number of days. After the waiting period ends, AWS KMS deletes the KMS key.
If you specify a value, it must be between `7` and `30`, inclusive. If you do not specify a value, it defaults to `30`.
If the KMS key is a multi-Region primary key with replicas, the waiting period begins when the last of its replica keys is deleted. Otherwise, the waiting period begins immediately.
* `is_enabled` - (Optional) Specifies whether the key is enabled. Defaults to `true`.
* `enable_key_rotation` - (Optional) Specifies whether [key rotation](http://docs.aws.amazon.com/kms/latest/developerguide/rotate-keys.html) is enabled. Defaults to false.
* `multi_region` - (Optional) Indicates whether the KMS key is a multi-Region (`true`) or regional (`false`) key. Defaults to `false`.
* `tags` - (Optional) A map of tags to assign to the object. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the key.
* `key_id` - The globally unique identifier for the key.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

KMS Keys can be imported using the `id`, e.g.,

```
$ terraform import aws_kms_key.a 1234abcd-12ab-34cd-56ef-1234567890ab
```
