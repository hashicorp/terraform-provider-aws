---
subcategory: "CodeCommit"
layout: "aws"
page_title: "AWS: aws_codecommit_approval_rule_template"
description: |-
  Provides a CodeCommit Approval Rule Template Resource.
---

# Resource: aws_codecommit_approval_rule_template

Provides a CodeCommit Approval Rule Template Resource.

## Example Usage

```terraform
resource "aws_codecommit_approval_rule_template" "example" {
  name        = "MyExampleApprovalRuleTemplate"
  description = "This is an example approval rule template"

  content = <<EOF
{
    "Version": "2018-11-08",
    "DestinationReferences": ["refs/heads/master"],
    "Statements": [{
        "Type": "Approvers",
        "NumberOfApprovalsNeeded": 2,
        "ApprovalPoolMembers": ["arn:aws:sts::123456789012:assumed-role/CodeCommitReview/*"]
    }]
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `content` - (Required) The content of the approval rule template. Maximum of 3000 characters.
* `name` - (Required) The name for the approval rule template. Maximum of 100 characters.
* `description` - (Optional) The description of the approval rule template. Maximum of 1000 characters.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `approval_rule_template_id` - The ID of the approval rule template
* `creation_date` - The date the approval rule template was created, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `last_modified_date` - The date the approval rule template was most recently changed, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `last_modified_user` - The Amazon Resource Name (ARN) of the user who made the most recent changes to the approval rule template.
* `rule_content_sha256` - The SHA-256 hash signature for the content of the approval rule template.

## Import

CodeCommit approval rule templates can be imported using the `name`, e.g.

```
$ terraform import aws_codecommit_approval_rule_template.imported ExistingApprovalRuleTemplateName
```
