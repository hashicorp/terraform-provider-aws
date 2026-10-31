---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ami_watermark"
description: |-
  Attaches a watermark to an Amazon Machine Image (AMI).
---

# Resource: aws_ami_watermark

Attaches a watermark to a non-public Amazon Machine Image (AMI). Watermarks are structured identifiers that propagate automatically to all derivative images created through `CreateImage` or `CopyImage`. Only the AMI owner can attach watermarks, and an AMI can have up to 5 watermarks.

## Example Usage

```terraform
resource "aws_ami_watermark" "example" {
  image_id       = aws_ami_copy.example.id
  watermark_name = "prod-baseline"
}
```

## Argument Reference

This resource supports the following arguments:

* `image_id` - (Required, Forces new resource) ID of the AMI.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `watermark_name` - (Required, Forces new resource) Name for the watermark. Combined with the caller's account ID to form the `watermark_key` (`accountId:watermarkName`). Must be 3-128 characters and may contain alphanumeric characters, parentheses `()`, square brackets `[]`, spaces, periods `.`, slashes `/`, dashes `-`, single quotes `'`, at-signs `@`, or underscores `_`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Watermark key in `accountId:watermarkName` format (for example, `123456789012:prod-baseline`).
* `watermark_key` - Watermark identifier in `accountId:watermarkName` format.

## Import

In Terraform v1.12.0 and later, the `import` block can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ami_watermark.example
  identity = {
    image_id      = "ami-12345678"
    watermark_key = "123456789012:prod-baseline"
  }
}

resource "aws_ami_watermark" "example" {
  image_id       = "ami-12345678"
  watermark_name = "prod-baseline"
}
```

### Identity Schema

#### Required

* `image_id` - ID of the AMI.
* `watermark_key` - Watermark identifier in `accountId:watermarkName` format.

#### Optional

* `account_id` (String) AWS account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AMI Watermarks using `IMAGE-ID,WATERMARK-KEY`. For example:

```terraform
import {
  to = aws_ami_watermark.example
  id = "ami-12345678,123456789012:prod-baseline"
}
```

Using `terraform import`, import AMI Watermarks using `IMAGE-ID,WATERMARK-KEY`. For example:

```console
% terraform import aws_ami_watermark.example ami-12345678,123456789012:prod-baseline
```

~> **Note:** The `watermark_name` attribute cannot be recovered on import because `DescribeImages` only returns the `watermark_key`. The resource will show a diff on the next plan after import; to suppress it, set `watermark_name` to the name portion of the `watermark_key`.
