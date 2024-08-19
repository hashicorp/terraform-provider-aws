---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_classification_export_configuration"
description: |-
  Provides a resource to manage Classification Results - Export Configuration
---

# Resource: aws_macie2_classification_export_configuration

Provides a resource to manage an [Amazon Macie Classification Export Configuration](https://docs.aws.amazon.com/macie/latest/APIReference/classification-export-configuration.html).

## Example Usage

```terraform
resource "aws_macie2_account" "example" {}

resource "aws_macie2_classification_export_configuration" "example" {
  depends_on = [
    aws_macie2_account.example,
  ]
  s3_destination {
    bucket_name = aws_s3_bucket.example.bucket
    key_prefix  = "exampleprefix/"
    kms_key_arn = aws_kms_key.example.arn
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `s3_destination` - (Required) Configuration block for a S3 Destination. Defined below

### s3_destination Configuration Block

The `s3_destination` configuration block supports the following arguments:

* `bucket_name` - (Required) The Amazon S3 bucket name in which Amazon Macie exports the data classification results.
* `key_prefix` - (Optional) The object key for the bucket in which Amazon Macie exports the data classification results.
* `kms_key_arn` - (Required) Amazon Resource Name (ARN) of the KMS key to be used to encrypt the data.

Additional information can be found in the [Storing and retaining sensitive data discovery results with Amazon Macie for AWS Macie documentation](https://docs.aws.amazon.com/macie/latest/user/discovery-results-repository-s3.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The unique identifier (ID) of the configuration.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_macie2_classification_export_configuration` using the account ID and region. For example:

```terraform
import {
  to = aws_macie2_classification_export_configuration.example
  id = "123456789012:us-west-2"
}
```

Using `terraform import`, import `aws_macie2_classification_export_configuration` using the account ID and region. For example:

```console
% terraform import aws_macie2_classification_export_configuration.example 123456789012:us-west-2
```
