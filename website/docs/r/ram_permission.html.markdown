---
subcategory: "RAM (Resource Access Manager)"
layout: "aws"
page_title: "AWS: aws_ram_permission"
description: |-
  Manages an AWS RAM (Resource Access Manager) Permission.
---

# Resource: aws_ram_permission

Manages an AWS RAM (Resource Access Manager) Permission.

## Example Usage

### Basic Usage

```terraform
resource "aws_ram_permission" "example" {
  name            = "custom-backup"
  policy_template = <<EOF
{
    "Effect": "Allow",
    "Action": [
	"backup:ListProtectedResourcesByBackupVault",
	"backup:ListRecoveryPointsByBackupVault",
	"backup:DescribeRecoveryPoint",
	"backup:DescribeBackupVault"
    ]
}
EOF
  resource_type   = "backup:BackupVault"

  tags = {
    Name = "custom-backup"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Specifies the name of the customer managed permission. The name must be unique within the AWS Region.
* `policy_template` - (Required) A string in JSON format string that contains the following elements of a resource-based policy: Effect, Action and Condition.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_type` - Specifies the name of the resource type that this customer managed permission applies to. The format is `<service-code>:<resource-type>` and is not case sensitive.
* `tags` - (Optional) A map of tags to assign to the resource share. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the permission.
* `default_version` - Specifies whether the version of the managed permission used by this resource share is the default version for this managed permission.
* `status` - The current status of the permission.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `version` - The version of the permission associated with this resource share.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ram_permission.example
  identity = {
    arn = "arn:aws:ram:us-west-1:123456789012:permission/test-permission"
  }
}

resource "aws_ram_permission" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` (String) Permission ARN.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RAM (Resource Access Manager) Permission using the `arn`. For example:

```terraform
import {
  to = aws_ram_permission.example
  id = "arn:aws:ram:us-west-1:123456789012:permission/test-permission"
}
```

Using `terraform import`, import RAM (Resource Access Manager) Permission using the `example_id_arg`. For example:

```console
% terraform import aws_ram_permission.example arn:aws:ram:us-west-1:123456789012:permission/test-permission
```
