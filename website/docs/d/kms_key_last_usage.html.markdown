---
subcategory: "KMS (Key Management)"
layout: "aws"
page_title: "AWS: aws_kms_key_last_usage"
description: |-
  Get last usage information for a KMS key
---

# Data Source: aws_kms_key_last_usage

Use this data source to get last usage information for a KMS key, including the most recent cryptographic operation performed with the key.

~> **Note:** AWS KMS tracks only the last successful cryptographic operation per key. There may be a delay of up to one hour between when an operation occurs and when it is recorded. Use `tracking_start_date` together with `key_creation_date` to interpret an empty `key_last_usage`:

* If `key_last_usage` is present, the key has been used for a cryptographic operation since tracking began.
* If `key_last_usage` is empty and `key_creation_date` is on or after `tracking_start_date`, the key has not been used since it was created.
* If `key_last_usage` is empty and `key_creation_date` is before `tracking_start_date`, the key has no recorded usage since tracking began but may have been used prior to that date. Examine past CloudTrail logs to determine earlier usage.

## Example Usage

```terraform
data "aws_kms_key_last_usage" "example" {
  key_id = "1234abcd-12ab-34cd-56ef-1234567890ab"
}
```

```terraform
data "aws_kms_key_last_usage" "by_arn" {
  key_id = "arn:aws:kms:us-east-1:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `key_id` - (Required) Key identifier. Must be a key ID (e.g., `1234abcd-12ab-34cd-56ef-1234567890ab`) or key ARN (e.g., `arn:aws:kms:us-east-1:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab`). Alias names are not supported.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `key_creation_date` - Date and time when the KMS key was created, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339#section-5.8).
* `key_last_usage` - Information about the last successful cryptographic operation performed with the key. Empty if the key has not been used since tracking began. See [`key_last_usage`](#key_last_usage) below.
* `tracking_start_date` - Date from which KMS began recording cryptographic activity for this key, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339#section-5.8).

### key_last_usage

* `cloud_trail_event_id` - CloudTrail event ID associated with the last successful cryptographic operation.
* `kms_request_id` - KMS request ID associated with the last successful cryptographic operation.
* `operation` - Last successful cryptographic operation the key was used for. Possible values: `Decrypt`, `DeriveSharedSecret`, `Encrypt`, `GenerateDataKey`, `GenerateDataKeyPair`, `GenerateDataKeyPairWithoutPlaintext`, `GenerateDataKeyWithoutPlaintext`, `GenerateMac`, `ReEncrypt`, `Sign`, `Verify`, `VerifyMac`.
* `timestamp` - Date and time when the key was most recently used for a successful cryptographic operation, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339#section-5.8).
