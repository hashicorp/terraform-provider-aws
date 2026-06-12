---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_region"
description: |-
  Terraform resource for managing an AWS SSO Admin Region.
---

# Resource: aws_ssoadmin_region

Terraform resource for managing an AWS SSO Admin Region.

Adds an additional AWS Region to an IAM Identity Center instance, replicating the instance to that Region. The operation is asynchronous — Terraform waits for the Region to reach `ACTIVE` status before completing.

~> Only one add or remove Region workflow can be in progress at a time for a given instance. Manage multiple regions by applying them sequentially or using `depends_on` to sequence resources.

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

* `instance_arn` - (Required) ARN of the IAM Identity Center instance to replicate to the target Region.
* `region_name` - (Required) Name of the AWS Region to add to the IAM Identity Center instance (for example, `us-east-1`). Changing this forces a new resource to be created.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `status` - Current status of the Region. Valid values are `ACTIVE`, `ADDING`, and `REMOVING`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`) Time to wait for the Region to finish replicating (reach `ACTIVE` status).
* `delete` - (Default `30m`) Time to wait for the Region to finish being removed.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

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

* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSO Admin Region using the `instance_arn` and `region_name`, separated by a comma (`,`). For example:

```terraform
import {
  to = aws_ssoadmin_region.example
  id = "arn:aws:sso:::instance/ssoins-1234567890abcdef,us-east-1"
}
```

Using `terraform import`, import SSO Admin Region using the `instance_arn` and `region_name`, separated by a comma (`,`). For example:

```console
% terraform import aws_ssoadmin_region.example arn:aws:sso:::instance/ssoins-1234567890abcdef,us-east-1
```
