---
subcategory: "IAM Access Analyzer"
layout: "aws"
page_title: "AWS: aws_accessanalyzer_archive_rule"
description: |-
  Terraform resource for managing an AWS AccessAnalyzer Archive Rule.
---

# Resource: aws_accessanalyzer_archive_rule

Terraform resource for managing an AWS AccessAnalyzer Archive Rule.

## Example Usage

### Basic Usage

```terraform
resource "aws_accessanalyzer_archive_rule" "example" {
  analyzer_name = "example-analyzer"
  rule_name     = "example-rule"

  filter {
    criteria = "condition.aws:UserId"
    eq       = ["userid"]
  }

  filter {
    criteria = "error"
    exists   = true
  }

  filter {
    criteria = "isPublic"
    eq       = ["false"]
  }
}
```

## Argument Reference

The following arguments are required:

* `analyzer_name` - (Required) Analyzer name.
* `filter` - (Required) Filter criteria for the archive rule. See [Filter](#filter) for more details.
* `rule_name` - (Required) Rule name.

### Filter

**Note** One comparator must be included with each filter.

* `criteria` - (Required) Filter criteria.
* `contains` - (Optional) Contains comparator.
* `eq` - (Optional) Equals comparator.
* `exists` - (Optional) Boolean comparator.
* `neq` - (Optional) Not Equals comparator.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Resource ID in the format: `analyzer_name/rule_name`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AccessAnalyzer ArchiveRule using the `analyzer_name/rule_name`. For example:

```terraform
import {
  to = aws_accessanalyzer_archive_rule.example
  id = "example-analyzer/example-rule"
}
```

Using `terraform import`, import AccessAnalyzer ArchiveRule using the `analyzer_name/rule_name`. For example:

```console
% terraform import aws_accessanalyzer_archive_rule.example example-analyzer/example-rule
```
