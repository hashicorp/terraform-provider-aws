---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_standards_subscription"
description: |-
  Subscribes to a Security Hub standard.
---

# Resource: aws_securityhub_standards_subscription

Subscribes to a Security Hub standard.

## Example Usage

```terraform
resource "aws_securityhub_account" "example" {}

data "aws_region" "current" {}

resource "aws_securityhub_standards_subscription" "cis" {
  depends_on    = [aws_securityhub_account.example]
  standards_arn = "arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
}

resource "aws_securityhub_standards_subscription" "pci_321" {
  depends_on    = [aws_securityhub_account.example]
  standards_arn = "arn:aws:securityhub:${data.aws_region.current.name}::standards/pci-dss/v/3.2.1"
}
```

## Argument Reference

This resource supports the following arguments:

* `standards_arn` - (Required) The ARN of a standard - see below.

Currently available standards (remember to replace `${var.partition}` and `${var.region}` as appropriate):

| Name                                     | ARN                                                                                                          |
|------------------------------------------|--------------------------------------------------------------------------------------------------------------|
| AWS Foundational Security Best Practices | `arn:${var.partition}:securityhub:${var.region}::standards/aws-foundational-security-best-practices/v/1.0.0` |
| AWS Resource Tagging Standard            | `arn:${var.partition}:securityhub:${var.region}::standards/aws-resource-tagging-standard/v/1.0.0`            |
| CIS AWS Foundations Benchmark v1.2.0     | `arn:${var.partition}:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0`                           |
| CIS AWS Foundations Benchmark v1.4.0     | `arn:${var.partition}:securityhub:${var.region}::standards/cis-aws-foundations-benchmark/v/1.4.0`            |
| NIST SP 800-53 Rev. 5                    | `arn:${var.partition}:securityhub:${var.region}::standards/nist-800-53/v/5.0.0`                              |
| PCI DSS                                  | `arn:${var.partition}:securityhub:${var.region}::standards/pci-dss/v/3.2.1`                                  |

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ARN of a resource that represents your subscription to a supported standard.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Hub standards subscriptions using the standards subscription ARN. For example:

```terraform
import {
  to = aws_securityhub_standards_subscription.cis
  id = "arn:aws:securityhub:eu-west-1:123456789012:subscription/cis-aws-foundations-benchmark/v/1.2.0"
}
```

```terraform
import {
  to = aws_securityhub_standards_subscription.pci_321
  id = "arn:aws:securityhub:eu-west-1:123456789012:subscription/pci-dss/v/3.2.1"
}
```

```terraform
import {
  to = aws_securityhub_standards_subscription.nist_800_53_rev_5
  id = "arn:aws:securityhub:eu-west-1:123456789012:subscription/nist-800-53/v/5.0.0"
}
```

Using `terraform import`, import Security Hub standards subscriptions using the standards subscription ARN. For example:

```console
% terraform import aws_securityhub_standards_subscription.cis arn:aws:securityhub:eu-west-1:123456789012:subscription/cis-aws-foundations-benchmark/v/1.2.0
```

```console
% terraform import aws_securityhub_standards_subscription.pci_321 arn:aws:securityhub:eu-west-1:123456789012:subscription/pci-dss/v/3.2.1
```

```console
% terraform import aws_securityhub_standards_subscription.nist_800_53_rev_5 arn:aws:securityhub:eu-west-1:123456789012:subscription/nist-800-53/v/5.0.0
```
