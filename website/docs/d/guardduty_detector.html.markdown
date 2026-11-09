---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_detector"
description: |-
  Retrieve information about a GuardDuty detector.
---

# Data Source: aws_guardduty_detector

Retrieve information about a GuardDuty detector.

## Example Usage

```terraform
data "aws_guardduty_detector" "example" {}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Optional) ID of the detector.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the detector.
* `features` - Current configuration of the detector features.
    * `additional_configuration` - Additional feature configuration.
        * `name` - Name of the additional configuration.
        * `status` - Status of the additional configuration.
    * `name` - Name of the detector feature.
    * `status` - Status of the detector feature.
* `finding_publishing_frequency` - Frequency of notifications sent about subsequent finding occurrences.
* `service_role_arn` - Service-linked role that grants GuardDuty access to the resources in the AWS account.
* `status` - Current status of the detector.
* `tags` - Map of tags for the resource.
