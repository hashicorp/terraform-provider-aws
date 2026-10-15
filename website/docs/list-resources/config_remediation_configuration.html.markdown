---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_remediation_configuration"
description: |-
  Lists Config Remediation Configuration resources.
---

# List Resource: aws_config_remediation_configuration

Lists Config Remediation Configuration resources.

## Example Usage

```terraform
list "aws_config_remediation_configuration" "example" {
  provider = aws

  config_rule_names = ["example-rule-1", "example-rule-2"]
}
```

## Argument Reference

This list resource supports the following arguments:

* `config_rule_names` - (Required) Names of the AWS Config rules.
* `region` - (Optional) Region to query. Defaults to provider region.
