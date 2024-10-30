---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_package_configuration"
description: |-
  Terraform resource for managing the AWS IoT Core Package Configuration.
---

# Resource: aws_iot_package_configuration

Terraform resource for managing the AWS IoT Core Package Configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_role" "example" {
  name = "example_role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "iot.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iot_package_configuration" "example" {
  version_update_by_jobs {
    enabled  = true
    role_arn = aws_iam_role.example.arn
  }
}
```

## Argument Reference

* `version_update_by_jobs` - (Optional) Configuration to manage job's package version reporting. This updates the thing's reserved named shadow that the job targets. Detailed below.

### version_update_by_jobs

* `enabled` - (Required) Whether to enable job's package version reporting.
* `role_arn` - (Optional) ARN of the IAM role that grants necessary permissions for IoT to update the thing's reserved named shadow.

## Attribute Reference

No additional attributes are exported.
