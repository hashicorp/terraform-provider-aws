---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_finding_ids"
description: |-
  Terraform data source for managing an AWS GuardDuty Finding Ids.
---

# Data Source: aws_guardduty_finding_ids

Terraform data source for managing an AWS GuardDuty Finding Ids.

## Example Usage

### Basic Usage

```terraform
data "aws_guardduty_finding_ids" "example" {
  detector_id = aws_guardduty_detector.example.id
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `detector_id` - (Required) ID of the GuardDuty detector.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `has_findings` - Indicates whether findings are present for the specified detector.
* `finding_ids` - A list of finding IDs for the specified detector.
