---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_enabled_standards"
description: |-
  Lists the standards that are currently enabled.
---

# Data Source: aws_securityhub_enabled_standards

Lists the standards that are currently enabled.

## Example Usage

```terraform
data "aws_securityhub_enabled_standards" "example" {}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `standards_subscription_arns` - (Optional) List of the standards subscription ARNs for the standards to retrieve.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `standards_subscriptions` - List of controls that apply to the specified standard. See below for details.

### `standards_subscriptions`

Each standard has the following attributes:

* `standards_arn` - ARN of the standard.
* `standards_controls_updatable` - Whether you can retrieve information about and configure individual controls that apply to the standard. Valid values: `READY_FOR_UPDATES`, `NOT_READY_FOR_UPDATES`.
* `standards_inputs` - Key-value map of input for the standard.
* `standards_status` - Status of your subscription to the standard. Valid values: `PENDING`, `READY`, `FAILED`, `DELETING`, `INCOMPLETE`.
* `standards_status_reason` - Reason for the current status. See below for details.
* `standards_subscription_arn` - ARN of the resource that represents your subscription to the standard.

### `standards_status_reason`

* `status_reason_code` - Reason code that represents the reason for the current status of a standard subscription. Valid values: `NO_AVAILABLE_CONFIGURATION_RECORDER`, `MAXIMUM_NUMBER_OF_CONFIG_RULES_EXCEEDED`, `INTERNAL_ERROR`.
