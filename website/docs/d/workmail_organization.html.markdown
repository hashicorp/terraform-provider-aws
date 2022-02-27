---
subcategory: "WorkMail"
layout: "aws"
page_title: "AWS: aws_workmail_organization"
description: |-
  Provides a WorkMail organization data source.
---

# Data Source: aws_workmail_organization

Provides information about a WorkMail organization.

## Example Usage

```terraform
variable "organization_id" {
  type    = string
  default = ""
}

data "aws_workmail_organization" "example" {
  organization_id  = var.organization_id
}
```

## Argument Reference

The following arguments are supported:

- `organization_id` - (Optional) The Organization ID for a specific WorkMail organization.

## Attributes Reference

website/docs/r/workmail_organization.markdown
See the [`aws_workmail_organization` resource](/docs/providers/aws/r/workmail_organization.html) for details on the
returned attributes - they are identical.
