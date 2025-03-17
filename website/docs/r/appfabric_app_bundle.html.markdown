---
subcategory: "AppFabric"
layout: "aws"
page_title: "AWS: aws_appfabric_app_bundle"
description: |-
  Terraform resource for managing an AWS AppFabric AppBundle.
---

# Resource: aws_appfabric_app_bundle

Terraform resource for managing an AWS AppFabric AppBundle.

## Example Usage

### Basic Usage

```terraform
resource "aws_appfabric_app_bundle" "example" {
  customer_managed_key_arn = awms_kms_key.example.arn
  tags = {
    Environment = "test"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `customer_managed_key_arn` - (Optional) The Amazon Resource Name (ARN) of the AWS Key Management Service (AWS KMS) key to use to encrypt the application data. If this is not specified, an AWS owned key is used for encryption.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the AppBundle.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppFabric AppBundle using the `arn`. For example:

```terraform
import {
  to = aws_appfabric_app_bundle.example
  id = "arn:aws:appfabric:[region]:[account]:appbundle/ee5587b4-5765-4288-a202-xxxxxxxxxx"
}
```

Using `terraform import`, import AppFabric AppBundle using the `arn`. For example:

```console
% terraform import aws_appfabric_app_bundle.example arn:aws:appfabric:[region]:[account]:appbundle/ee5587b4-5765-4288-a202-xxxxxxxxxx
```
