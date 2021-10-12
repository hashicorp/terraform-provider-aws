---
subcategory: "IoT"
layout: "aws"
page_title: "AWS: aws_iot_greengrass_service_role"
description: |-
  Retrieves the Greengrass service role that is attached to the current account
---

# Data Source: aws_iot_greengrass_service_role

Returns the Greengrass service role that is attached to the current account. See also https://docs.aws.amazon.com/greengrass/latest/apireference/-greengrass-servicerole.html

## Example Usage

```terraform
data "aws_iot_greengrass_service_role" "example" {}

output "greengrass_service_role" {
  value = data.aws_iot_greengrass_service_role.example.role_arn
}
```

## Attributes Reference

* `role_arn` - The ARN of the IAM Greengrass service role which is associated with the account.
