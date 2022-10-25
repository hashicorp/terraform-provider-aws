---
subcategory: "CodeCommit"
layout: "aws"
page_title: "AWS: aws_codecommit_approval_rule_template_association"
description: |-
  Associates a CodeCommit Approval Rule Template with a Repository.
---

# Resource: aws_codecommit_approval_rule_template_association

Associates a CodeCommit Approval Rule Template with a Repository.

## Example Usage

```terraform
resource "aws_codecommit_approval_rule_template_association" "example" {
  approval_rule_template_name = aws_codecommit_approval_rule_template.example.name
  repository_name             = aws_codecommit_repository.example.repository_name
}
```

## Argument Reference

The following arguments are supported:

* `approval_rule_template_name` - (Required) The name for the approval rule template.
* `repository_name` - (Required) The name of the repository that you want to associate with the template.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the approval rule template and name of the repository, separated by a comma (`,`).

## Import

CodeCommit approval rule template associations can be imported using the `approval_rule_template_name` and `repository_name` separated by a comma (`,`), e.g.

```
$ terraform import aws_codecommit_approval_rule_template_association.example approver-rule-for-example,MyExampleRepo
```
