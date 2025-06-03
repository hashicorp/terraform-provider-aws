---
subcategory: "Audit Manager"
layout: "aws"
page_title: "AWS: aws_auditmanager_framework"
description: |-
  Terraform data source for managing an AWS Audit Manager Framework.
---

# Data Source: aws_auditmanager_framework

Terraform data source for managing an AWS Audit Manager Framework.

## Example Usage

### Basic Usage

```terraform
data "aws_auditmanager_framework" "example" {
  name           = "Essential Eight"
  framework_type = "Standard"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the framework.
* `type` - (Required) Type of framework. Valid values are `Custom` and `Standard`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

See the [`aws_auditmanager_framework` resource](/docs/providers/aws/r/auditmanager_framework.html) for details on the returned attributes - they are identical.
