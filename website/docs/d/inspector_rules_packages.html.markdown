---
subcategory: "Inspector Classic"
layout: "aws"
page_title: "AWS: aws_inspector_rules_packages"
description: |-
    Provides a list of Amazon Inspector Classic Rules packages which can be used by Amazon Inspector Classic.
---

# Data Source: aws_inspector_rules_packages

The Amazon Inspector Classic Rules Packages data source allows access to the list of AWS
Inspector Rules Packages which can be used by Amazon Inspector Classic within the region
configured in the provider.

## Example Usage

```terraform
# Declare the data source
data "aws_inspector_rules_packages" "rules" {}

# e.g., Use in aws_inspector_assessment_template
resource "aws_inspector_resource_group" "group" {
  tags = {
    test = "test"
  }
}

resource "aws_inspector_assessment_target" "assessment" {
  name               = "test"
  resource_group_arn = aws_inspector_resource_group.group.arn
}

resource "aws_inspector_assessment_template" "assessment" {
  name       = "Test"
  target_arn = aws_inspector_assessment_target.assessment.arn
  duration   = "60"

  rules_package_arns = data.aws_inspector_rules_packages.rules.arns
}
```

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `arns` - List of the Amazon Inspector Classic Rules Packages arns available in the AWS region.
