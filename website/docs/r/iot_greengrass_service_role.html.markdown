---
subcategory: "IoT"
layout: "aws"
page_title: "AWS: aws_iot_greengrass_service_role"
description: |-
    Manages the Greengrass service role of the account.
---

# Resource: aws_iot_greengrass_service_role

Manages the Greengrass service role of the account.

## Example Usage

```hcl
resource "aws_iam_role" "greengrass_service_role" {
  name = "greengrass_service_role"
  assume_role_policy = <<EOF
{
"Version": "2012-10-17",
"Statement": [
	{
	"Effect": "Allow",
	"Principal": {
		"Service": "greengrass.amazonaws.com"
	},
	"Action": "sts:AssumeRole"
	}
]
}
EOF
}

resource "aws_iot_greengrass_service_role" "example" {
  role_arn    = aws_iam_role.greengrass_service_role.arn
}
```


## Argument Reference

* `role_arn` - (Required)  ARN of the IAM role to set as the Greengrass service role for the account 

## Attributes Reference

In addition to the arguments, the following attributes are exported:

* `associated_at` - The time when the service role was associated with the account.

