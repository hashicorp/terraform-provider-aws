---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_security_controls"
description: |-
  Lists security controls.
---

# Data Source: aws_securityhub_security_controls

Lists security controls.

## Example Usage

### All Controls

```terraform
data "aws_securityhub_security_controls" "example" {}
```

### `HIGH` or `CRITICAL` Severity Controls

```terraform
data "aws_securityhub_security_controls" "example" {}

output "security_control_definitions" {
  value = [for d in data.aws_securityhub_security_controls.example.security_control_definitions : d if contains(["HIGH", "CRITICAL"], d.severity_rating)]
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `standards_arn` - (Optional) ARN of the standard that you want to list controls for. If omitted, all controls are returned.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `security_control_definitions` - List of controls. See below for details.

### `security_control_definitions`

Each control has the following attributes:

* `current_region_availability` - Whether the security control is available in the current AWS Region. Valid values: `AVAILABLE`, `UNAVAILABLE`.
* `customizable_properties` - Security control properties that you can customize.
* `description` - Description of the security control across standards.
* `remediation_url` - Link to Security Hub CSPM documentation that explains how to remediate a failed finding for the security control.
* `security_control_id` - Unique identifier of the security control across standards.
* `severity_rating` - Severity of the security control. Valid values: `LOW`, `MEDIUM`, `HIGH`, `CRITICAL`.
* `title` - Title of the security control.
