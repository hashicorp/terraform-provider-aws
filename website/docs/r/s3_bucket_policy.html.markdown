---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_policy"
description: |-
  Attaches a policy to an S3 bucket resource.
---

# Resource: aws_s3_bucket_policy

Attaches a policy to an S3 bucket resource.

-> Policies can be attached to both S3 general purpose buckets and S3 directory buckets.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "my-tf-test-bucket"
}

resource "aws_s3_bucket_policy" "allow_access_from_another_account" {
  bucket = aws_s3_bucket.example.id
  policy = data.aws_iam_policy_document.allow_access_from_another_account.json
}

data "aws_iam_policy_document" "allow_access_from_another_account" {
  statement {
    principals {
      type        = "AWS"
      identifiers = ["123456789012"]
    }

    actions = [
      "s3:GetObject",
      "s3:ListBucket",
    ]

    resources = [
      aws_s3_bucket.example.arn,
      "${aws_s3_bucket.example.arn}/*",
    ]
  }
}
```

-> Only one `aws_s3_bucket_policy` resource should be defined per S3 bucket. Defining multiple `aws_s3_bucket_policy` resources with different Terraform names but the same `bucket` value may result in unexpected policy overwrites. Each resource uses the `PutBucketPolicy` API, which replaces the entire existing policy without error or warning. Because Terraform treats each resource independently, the policy applied last will silently override any previously applied policy.

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `bucket` - (Required) Name of the bucket to which to apply the policy.
* `policy` - (Required) Text of the policy. Although this is a bucket policy rather than an IAM policy, the [`aws_iam_policy_document`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) data source may be used, so long as it specifies a principal. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy). Note: Bucket policies are limited to 20 KB in size.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 bucket policies using the bucket name. For example:

```terraform
import {
  to = aws_s3_bucket_policy.allow_access_from_another_account
  id = "my-tf-test-bucket"
}
```

Using `terraform import`, import S3 bucket policies using the bucket name. For example:

```console
% terraform import aws_s3_bucket_policy.allow_access_from_another_account my-tf-test-bucket
```
