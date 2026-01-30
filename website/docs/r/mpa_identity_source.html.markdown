---
subcategory: "Multi-party Approval"
layout: "aws"
page_title: "AWS: aws_mpa_identity_source"
description: |-
  Manages an AWS Multi-party Approval Identity Source.
---

# Resource: aws_mpa_identity_source

Manages an AWS Multi-party Approval Identity Source.

## Example Usage

### Basic Usage

```terraform
data "aws_ssoadmin_instances" "example" {}
data "aws_region" "current" {}

resource "aws_mpa_identity_source" "example" {
  name = "example-identity-source"

  identity_source_parameters {
    iam_identity_center {
      instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
      region       = data.aws_region.current.name
    }
  }

  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name for the identity source. Changing this value forces a new resource to be created.
* `identity_source_parameters` - (Required) Configuration block for identity source parameters. Changing this value forces a new resource to be created. See [`identity_source_parameters`](#identity_source_parameters) below.

The following arguments are optional:

* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### identity_source_parameters

* `iam_identity_center` - (Required) Configuration block for IAM Identity Center. See [`iam_identity_center`](#iam_identity_center) below.

### iam_identity_center

* `instance_arn` - (Required) ARN of the IAM Identity Center instance.
* `region` - (Required) AWS Region where the IAM Identity Center instance is located.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Identity Source.
* `creation_time` - Timestamp when the identity source was created.
* `id` - ARN of the Identity Source.
* `status` - Status of the identity source (e.g., `ACTIVE`, `CREATING`, `DELETING`, `ERROR`).
* `status_code` - Status code providing additional details about the identity source status.
* `status_message` - Message describing the status of the identity source.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Multi-party Approval Identity Source using the ARN. For example:

```terraform
import {
  to = aws_mpa_identity_source.example
  id = "arn:aws:mpa:us-east-1:123456789012:identity-source/example"
}
```

Using `terraform import`, import Multi-party Approval Identity Source using the ARN. For example:

```console
% terraform import aws_mpa_identity_source.example arn:aws:mpa:us-east-1:123456789012:identity-source/example
```
