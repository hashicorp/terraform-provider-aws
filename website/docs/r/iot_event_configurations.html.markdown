---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_event_configurations"
description: |-
    Manages IoT event configurations.
---

# Resource: aws_iot_event_configurations

Manages IoT event configurations.

~> **NOTE:** Deleting this resource does not disable the event configurations, the resource in simply removed from state instead.

## Example Usage

```terraform
resource "aws_iot_event_configurations" "example" {
  event_configurations = {
    "THING"                  = true,
    "THING_GROUP"            = false,
    "THING_TYPE"             = false,
    "THING_GROUP_MEMBERSHIP" = false,
    "THING_GROUP_HIERARCHY"  = false,
    "THING_TYPE_ASSOCIATION" = false,
    "JOB"                    = false,
    "JOB_EXECUTION"          = false,
    "POLICY"                 = false,
    "CERTIFICATE"            = true,
    "CA_CERTIFICATE"         = false,
  }
}
```

## Argument Reference

* `event_configurations` - (Required) Map. The new event configuration values. You can use only these strings as keys: `THING_GROUP_HIERARCHY`, `THING_GROUP_MEMBERSHIP`, `THING_TYPE`, `THING_TYPE_ASSOCIATION`, `THING_GROUP`, `THING`, `POLICY`, `CA_CERTIFICATE`, `JOB_EXECUTION`, `CERTIFICATE`, `JOB`. Use boolean for values of mapping.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IoT Event Configurations using the AWS Region. For example:

```terraform
import {
  to = aws_iot_event_configurations.example
  id = "us-west-2"
}
```

Using `terraform import`, import IoT Event Configurations using the AWS Region. For example:

```console
% terraform import aws_iot_event_configurations.example us-west-2
```
