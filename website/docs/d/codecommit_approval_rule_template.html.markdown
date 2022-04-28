---
subcategory: "CodeCommit"
layout: "aws"
page_title: "AWS: aws_codecommit_approval_rule_template"
description: |-
  Provides details about a specific CodeCommit Approval Rule Template.
---

# Data Source: aws_codecommit_approval_rule_template

Provides details about a specific CodeCommit Approval Rule Template.

## Example Usage

```terraform
data "aws_codecommit_approval_rule_template" "example" {
  name = "MyExampleApprovalRuleTemplate"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name for the approval rule template. This needs to be less than 100 characters.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `approval_rule_template_id` - The ID of the approval rule template.
* `content` - The content of the approval rule template.
* `creation_date` - The date the approval rule template was created, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `description` - The description of the approval rule template.
* `last_modified_date` - The date the approval rule template was most recently changed, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `last_modified_user` - The Amazon Resource Name (ARN) of the user who made the most recent changes to the approval rule template.
* `rule_content_sha256` - The SHA-256 hash signature for the content of the approval rule template.
