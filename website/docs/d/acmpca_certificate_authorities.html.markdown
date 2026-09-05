---
subcategory: "ACM PCA (Certificate Manager Private Certificate Authority)"
layout: "aws"
page_title: "AWS: aws_acmpca_certificate_authorities"
description: |-
  Get information about a set of AWS ACM PCA (Certificate Manager Private Certificate Authority) Certificate Authorities.
---


# Data Source: aws_acmpca_certificate_authorities

Provides details about an AWS ACM PCA (Certificate Manager Private Certificate Authority) Certificate Authorities.

## Example Usage

### Basic Usage

```terraform
data "aws_acmpca_certificate_authorities" "example" {
}
```

### List ACM PCAs shared via RAM

```terraform
data "aws_acmpca_certificate_authorities" "example" {
  resource_owner = "OTHER_ACCOUNTS"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_owner` (Optional) Use this argument to filter the returned set of certificate authorities based on their owner. Valid values are `SELF` and `OTHER_ACCOUNTS` The default is `SELF`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Set of ARNs of the matched ACM PCA
