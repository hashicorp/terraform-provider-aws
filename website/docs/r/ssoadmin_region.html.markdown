---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_region"
description: |-
  Terraform resource for managing an AWS SSO Admin Region.
---

# Resource: aws_ssoadmin_region

Terraform resource for managing an AWS SSO Admin Region.

Adds another AWS Region to an IAM Identity Center instance. This operation runs asynchronously, and Terraform waits until the Region status becomes `ACTIVE`.

~> For a given instance, only one Region add or remove operation can run at a time. If you manage multiple regions, apply them one at a time or use `depends_on`.

~> The primary Region of an IAM Identity Center instance cannot be removed.

## Example Usage

### Basic Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_region" "example" {
  instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  region_name  = "us-east-1"
}
```

## Argument Reference

The following arguments are required:

* `instance_arn` - (Required) ARN of the IAM Identity Center instance.
* `region_name` - (Required) AWS Region to add (for example, `us-east-1`). Changing this forces a new resource.

The following arguments are optional:

* `region` - (Optional) Region where Terraform calls the SSO Admin API for this resource. Defaults to the Region in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `status` - Current Region status. Valid values are `ACTIVE`, `ADDING`, and `REMOVING`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`) Time to wait until the Region reaches `ACTIVE`.
* `delete` - (Default `30m`) Time to wait until the Region is removed.

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with `identity`. For example:

```terraform
import {
  to = aws_ssoadmin_region.example
  identity = {
    instance_arn = "arn:aws:sso:::instance/ssoins-1234567890abcdef"
    region_name  = "us-east-1"
  }
}

resource "aws_ssoadmin_region" "example" {
  # Configuration omitted for brevity.
}
```

### Identity Schema

#### Required

* `instance_arn` (String) ARN of the IAM Identity Center instance.
* `region_name` (String) Name of the AWS Region.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import this resource with `instance_arn` and `region_name` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_ssoadmin_region.example
  id = "arn:aws:sso:::instance/ssoins-1234567890abcdef,us-east-1"
}
```

Using `terraform import`, import this resource with `instance_arn` and `region_name` separated by a comma (`,`). For example:

```console
% terraform import aws_ssoadmin_region.example arn:aws:sso:::instance/ssoins-1234567890abcdef,us-east-1
```
