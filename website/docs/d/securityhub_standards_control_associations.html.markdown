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

* `standards_arns` - Set of ARNs of the standards that the security control is associated with.
