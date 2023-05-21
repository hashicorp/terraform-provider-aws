---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_template_alias"
description: |-
  Terraform resource for managing an AWS QuickSight Template Alias.
---

# Resource: aws_quicksight_template_alias

Terraform resource for managing an AWS QuickSight Template Alias.

## Example Usage

### Basic Usage

```terraform
resource "aws_quicksight_template_alias" "example" {
  alias_name              = "example-alias"
  template_id             = aws_quicksight_template.test.template_id
  template_version_number = aws_quicksight_template.test.version_number
}
```

## Argument Reference

The following arguments are required:

* `alias_name` - (Required, Forces new resource) Display name of the template alias.
* `template_id` - (Required, Forces new resource) ID of the template.
* `template_version_number` - (Required) Version number of the template.

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the template alias.
* `id` - A comma-delimited string joining AWS account ID, template ID, and alias name.

## Import

QuickSight Template Alias can be imported using the AWS account ID, template ID, and alias name separated by a comma (`,`) e.g.,

```
$ terraform import aws_quicksight_template_alias.example 123456789012,example-id,example-alias
```
