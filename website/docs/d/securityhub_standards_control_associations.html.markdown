---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_standards_control_associations"
description: |-
  Terraform data source for managing an AWS Security Hub Standards Control Associations.
---

# Resource: aws_securityhub_standards_control_associations

Terraform data source for managing an AWS Security Hub Standards Control Associations.

## Example Usage

### Basic Usage

```terraform
resource "aws_securityhub_account" "test" {}

data "aws_securityhub_standards_control_associations" "test" {
  security_control_id = "IAM.1"

  depends_on = [aws_securityhub_account.test]
}
```

## Argument Reference

* `security_control_id` - (Required) The identifier of the control (identified with `SecurityControlId`, `SecurityControlArn`, or a mix of both parameters).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `standards_control_associations` - A list that provides the status and other details for each security control that applies to each enabled standard.
See [`standards_control_associations`](#standards_control_associations-attribute-reference) below.

### `standards_control_associations` Attribute Reference

* `association_status` - Enablement status of a control in a specific standard.
* `related_requirements` - List of underlying requirements in the compliance framework related to the standard.
* `security_control_arn` - ARN of the security control.
* `security_control_id` - ID of the security control.
* `standards_arn` - ARN of the standard.
* `standards_control_description` - Description of the standard.
* `standards_control_title` - Title of the standard.
* `updated_at` - Last time that a control's enablement status in a specified standard was updated.
* `updated_reason` - Reason for updating a control's enablement status in a specified standard.
