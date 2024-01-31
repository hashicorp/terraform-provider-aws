---
subcategory: "SSM Contacts"
layout: "aws"
page_title: "AWS: aws_ssmcontacts_rotation"
description: |-
  Provides a Terraform data source for managing a Contacts Rotation in AWS Systems Manager Incident Manager
---

# Data Source: aws_ssmcontacts_rotation

Provides a Terraform data source for managing a Contacts Rotation in AWS Systems Manager Incident Manager

## Example Usage

### Basic Usage

```terraform
data "aws_ssmcontacts_rotation" "example" {
  arn = "arn:aws:ssm-contacts:us-east-1:012345678910:rotation/example"
}
```

## Argument Reference

The following arguments are required:

* `arn` - (Required) The Amazon Resource Name (ARN) of the rotation.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `contact_ids` - The Amazon Resource Names (ARNs) of the contacts to add to the rotation. The order in which you list the contacts is their shift order in the rotation schedule.
* `name` - The name for the rotation.
* `time_zone_id` - The time zone to base the rotation’s activity on in Internet Assigned Numbers Authority (IANA) format.
* `recurrence` - Information about when an on-call rotation is in effect and how long the rotation period lasts.
* `start_time` - The date and time, in RFC 3339 format, that the rotation goes into effect.
* `tags` - A map of tags to assign to the resource.
