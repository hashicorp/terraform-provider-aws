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

This resource supports the following arguments:

* `approval_rule_template_name` - (Required) The name for the approval rule template.
* `repository_name` - (Required) The name of the repository that you want to associate with the template.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the approval rule template and name of the repository, separated by a comma (`,`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeCommit approval rule template associations using the `approval_rule_template_name` and `repository_name` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_codecommit_approval_rule_template_association.example
  id = "approver-rule-for-example,MyExampleRepo"
}
```

Using `terraform import`, import CodeCommit approval rule template associations using the `approval_rule_template_name` and `repository_name` separated by a comma (`,`). For example:

```console
% terraform import aws_codecommit_approval_rule_template_association.example approver-rule-for-example,MyExampleRepo
```
