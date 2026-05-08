---
subcategory: "EBS (EC2)"
layout: "aws"
page_title: "AWS: aws_ebs_default_kms_key"
description: |-
  Provides metadata about the KMS key set for EBS default encryption
---

# Data Source: aws_ebs_default_kms_key

Use this data source to get the default EBS encryption KMS key in the current region.

## Example Usage

```terraform
data "aws_ebs_default_kms_key" "current" {}

resource "aws_ebs_volume" "example" {
  availability_zone = "us-west-2a"

  encrypted  = true
  kms_key_id = data.aws_ebs_default_kms_key.current.key_arn
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `key_arn` - ARN of the default KMS key uses to encrypt an EBS volume in this region when no key is specified in an API call that creates the volume and encryption by default is enabled.
* `id` - Region of the default KMS Key.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
