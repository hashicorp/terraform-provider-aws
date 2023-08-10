---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_product_subscription"
description: |-
  Subscribes to a Security Hub product.
---

# Resource: aws_securityhub_product_subscription

Subscribes to a Security Hub product.

## Example Usage

```terraform
resource "aws_securityhub_account" "example" {}

data "aws_region" "current" {}

resource "aws_securityhub_product_subscription" "example" {
  depends_on  = [aws_securityhub_account.example]
  product_arn = "arn:aws:securityhub:${data.aws_region.current.name}:733251395267:product/alertlogic/althreatmanagement"
}
```

## Argument Reference

This resource supports the following arguments:

* `product_arn` - (Required) The ARN of the product that generates findings that you want to import into Security Hub - see below.

Amazon maintains a list of [Product integrations in AWS Security Hub](https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-findings-providers.html) that changes over time. Any of the products on the linked [Available AWS service integrations](https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-internal-providers.html) or [Available third-party partner product integrations](https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-partner-providers.html) can be configured using `aws_securityhub_product_subscription`.

Available products can also be listed by running the AWS CLI command `aws securityhub describe-products`.

A subset of currently available products (remember to replace `${var.region}` as appropriate) includes:

* `arn:aws:securityhub:${var.region}::product/aws/guardduty`
* `arn:aws:securityhub:${var.region}::product/aws/inspector`
* `arn:aws:securityhub:${var.region}::product/aws/macie`
* `arn:aws:securityhub:${var.region}::product/alertlogic/althreatmanagement`
* `arn:aws:securityhub:${var.region}::product/armordefense/armoranywhere`
* `arn:aws:securityhub:${var.region}::product/barracuda/cloudsecurityguardian`
* `arn:aws:securityhub:${var.region}::product/checkpoint/cloudguard-iaas`
* `arn:aws:securityhub:${var.region}::product/checkpoint/dome9-arc`
* `arn:aws:securityhub:${var.region}::product/crowdstrike/crowdstrike-falcon`
* `arn:aws:securityhub:${var.region}::product/cyberark/cyberark-pta`
* `arn:aws:securityhub:${var.region}::product/f5networks/f5-advanced-waf`
* `arn:aws:securityhub:${var.region}::product/fortinet/fortigate`
* `arn:aws:securityhub:${var.region}::product/guardicore/aws-infection-monkey`
* `arn:aws:securityhub:${var.region}::product/guardicore/guardicore`
* `arn:aws:securityhub:${var.region}::product/ibm/qradar-siem`
* `arn:aws:securityhub:${var.region}::product/imperva/imperva-attack-analytics`
* `arn:aws:securityhub:${var.region}::product/mcafee-skyhigh/mcafee-mvision-cloud-aws`
* `arn:aws:securityhub:${var.region}::product/paloaltonetworks/redlock`
* `arn:aws:securityhub:${var.region}::product/paloaltonetworks/vm-series`
* `arn:aws:securityhub:${var.region}::product/qualys/qualys-pc`
* `arn:aws:securityhub:${var.region}::product/qualys/qualys-vm`
* `arn:aws:securityhub:${var.region}::product/rapid7/insightvm`
* `arn:aws:securityhub:${var.region}::product/sophos/sophos-server-protection`
* `arn:aws:securityhub:${var.region}::product/splunk/splunk-enterprise`
* `arn:aws:securityhub:${var.region}::product/splunk/splunk-phantom`
* `arn:aws:securityhub:${var.region}::product/sumologicinc/sumologic-mda`
* `arn:aws:securityhub:${var.region}::product/symantec-corp/symantec-cwp`
* `arn:aws:securityhub:${var.region}::product/tenable/tenable-io`
* `arn:aws:securityhub:${var.region}::product/trend-micro/deep-security`
* `arn:aws:securityhub:${var.region}::product/turbot/turbot`
* `arn:aws:securityhub:${var.region}::product/twistlock/twistlock-enterprise`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of a resource that represents your subscription to the product that generates the findings that you want to import into Security Hub.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Hub product subscriptions using `product_arn,arn`. For example:

```terraform
import {
  to = aws_securityhub_product_subscription.example
  id = "arn:aws:securityhub:eu-west-1:733251395267:product/alertlogic/althreatmanagement,arn:aws:securityhub:eu-west-1:123456789012:product-subscription/alertlogic/althreatmanagement"
}
```

Using `terraform import`, import Security Hub product subscriptions using `product_arn,arn`. For example:

```console
% terraform import aws_securityhub_product_subscription.example arn:aws:securityhub:eu-west-1:733251395267:product/alertlogic/althreatmanagement,arn:aws:securityhub:eu-west-1:123456789012:product-subscription/alertlogic/althreatmanagement
```
