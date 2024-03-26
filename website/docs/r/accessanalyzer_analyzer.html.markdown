---
subcategory: "IAM Access Analyzer"
layout: "aws"
page_title: "AWS: aws_accessanalyzer_analyzer"
description: |-
  Manages an Access Analyzer Analyzer
---

# Resource: aws_accessanalyzer_analyzer

Manages an Access Analyzer Analyzer. More information can be found in the [Access Analyzer User Guide](https://docs.aws.amazon.com/IAM/latest/UserGuide/what-is-access-analyzer.html).

## Example Usage

### Account Analyzer

```terraform
resource "aws_accessanalyzer_analyzer" "example" {
  analyzer_name = "example"
}
```

### Organization Analyzer

```terraform
resource "aws_organizations_organization" "example" {
  aws_service_access_principals = ["access-analyzer.amazonaws.com"]
}

resource "aws_accessanalyzer_analyzer" "example" {
  depends_on = [aws_organizations_organization.example]

  analyzer_name = "example"
  type          = "ORGANIZATION"
}
```

## Argument Reference

The following arguments are required:

* `analyzer_name` - (Required) Name of the Analyzer.

The following arguments are optional:

* `configuration` - (Optional) A block that specifies the configuration of the analyzer. [Documented below](#configuration-argument-reference)
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Optional) Type of Analyzer. Valid values are `ACCOUNT`, `ORGANIZATION`, `ACCOUNT_UNUSED_ACCESS `, `ORGANIZATION_UNUSED_ACCESS`. Defaults to `ACCOUNT`.

### `configuration` Argument Reference

* `unused_access` - A block that specifies the configuration of an unused access analyzer for an AWS organization or account. [Documented below](#unused_access-argument-reference)

### `unused_access` Argument Reference

* `unused_access_age` - The specified access age in days for which to generate findings for unused access.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Analyzer.
* `id` - Analyzer name.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Access Analyzer Analyzers using the `analyzer_name`. For example:

```terraform
import {
  to = aws_accessanalyzer_analyzer.example
  id = "example"
}
```

Using `terraform import`, import Access Analyzer Analyzers using the `analyzer_name`. For example:

```console
% terraform import aws_accessanalyzer_analyzer.example example
```
