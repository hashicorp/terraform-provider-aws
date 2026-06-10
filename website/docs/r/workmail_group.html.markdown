---
subcategory: "WorkMail"
layout: "aws"
page_title: "AWS: aws_workmail_group"
description: |-
  Manages an AWS WorkMail Group.
---

# Resource: aws_workmail_group

Manages an AWS WorkMail Group.

## Example Usage

```terraform
resource "aws_workmail_organization" "example" {
  organization_alias = "example-workmail-org"
  delete_directory   = true
}

resource "aws_workmail_group" "example" {
  organization_id = aws_workmail_organization.example.organization_id
  email           = "engineering@${aws_workmail_organization.example.default_mail_domain}"
  name            = "engineering"
}
```

## Argument Reference

The following arguments are required:

* `email` - Primary email address used to register the group with WorkMail.
* `name` - Name of the group.
* `organization_id` - Identifier of the WorkMail organization where the group is managed.

The following arguments are optional:

* `hidden_from_global_address_list` - Whether to hide the group from the global address list.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `disabled_date` - Timestamp when the group was disabled from WorkMail use.
* `enabled_date` - Timestamp when the group was enabled for WorkMail use.
* `group_id` - Identifier of the group.
* `state` - Current WorkMail state of the group.

## Import

In Terraform v1.12.0 and later, the `import` block can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_workmail_group.example
  identity = {
    organization_id = "m-1234567890abcdef"
    group_id        = "S-1-1-12-1234567890-123456789-123456789-1234"
  }
}

resource "aws_workmail_group" "example" {
  organization_id = "m-1234567890abcdef"
  email           = "engineering@example.awsapps.com"
  name            = "engineering"
}
```

### Identity Schema

#### Required

* `group_id` - Identifier of the group.
* `organization_id` - Identifier of the WorkMail organization where the group is managed.

#### Optional

* `account_id` (String) AWS account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an `import` block to import WorkMail Group using `organization_id,group_id`. For example:

```terraform
import {
  to = aws_workmail_group.example
  id = "m-1234567890abcdef,S-1-1-12-1234567890-123456789-123456789-1234"
}
```

Using `terraform import`, import WorkMail Group using `organization_id,group_id`. For example:

```console
% terraform import aws_workmail_group.example m-1234567890abcdef,S-1-1-12-1234567890-123456789-123456789-1234
```
