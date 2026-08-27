---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_application_status_check_association"
description: |-
  Manages an EC2 Application Status Check Association.
---

# Resource: aws_ec2_application_status_check_association

Manages the association between an EC2 Application Status Check and either an EC2 instance or instances selected by a tag.

## Example Usage

### Basic Usage

```hcl
resource "aws_ec2_application_status_check" "example" {
  protocol = "http"
  port     = 80
}

resource "aws_ec2_application_status_check_association" "example" {
  application_status_check_id = aws_ec2_application_status_check.example.id
  target_tag_key              = "Environment"
  target_tag_value            = "production"
}
```

### Instance Association

```hcl
resource "aws_ec2_application_status_check" "example" {
  protocol = "http"
  port     = 80
}

resource "aws_ec2_application_status_check_association" "example" {
  application_status_check_id = aws_ec2_application_status_check.example.id
  instance_id                 = "i-0123456789abcdef0"
}
```

## Argument Reference

The following arguments are required:

* `application_status_check_id` - (Required) ID of the application status check.

The following arguments are optional:

* `instance_id` - (Optional) ID of the EC2 instance to associate. Exactly one target must be specified: `instance_id` or the `target_tag_key` and `target_tag_value` pair.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `target_tag_key` - (Optional) Tag key used to select EC2 instances. Must be specified with `target_tag_value` and cannot be specified with `instance_id`.
* `target_tag_value` - (Optional) Tag value used to select EC2 instances. Must be specified with `target_tag_key` and cannot be specified with `instance_id`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ec2_application_status_check_association.example
  identity = {
    application_status_check_id = "asc-0123456789abcdef0"
    target_tag_key              = "Environment"
    target_tag_value            = "production"
  }
}
```

### Identity Schema

#### Required

* `application_status_check_id` (String) ID of the application status check.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `instance_id` (String) ID of the associated EC2 instance.
* `region` (String) Region where this resource is managed.
* `target_tag_key` (String) Tag key used to select associated EC2 instances.
* `target_tag_value` (String) Tag value used to select associated EC2 instances.

Exactly one target must be specified in the identity: `instance_id` or the `target_tag_key` and `target_tag_value` pair.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Application Status Check Associations using `<application-status-check-id>,instance-id,<instance-id>` for instance associations or `<application-status-check-id>,tag,<URL-query-escaped-tag-key>,<URL-query-escaped-tag-value>` for tag associations. For example:

```terraform
import {
  to = aws_ec2_application_status_check_association.example
  id = "asc-0123456789abcdef0,tag,Environment,production"
}
```

Using `terraform import`, import EC2 Application Status Check Associations using the same identifier format. For example:

```console
% terraform import aws_ec2_application_status_check_association.example asc-0123456789abcdef0,instance-id,i-0123456789abcdef0
```
