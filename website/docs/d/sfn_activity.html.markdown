---
subcategory: "SFN (Step Functions)"
layout: "aws"
page_title: "AWS: aws_sfn_activity"
description: |-
  Use this data source to get information about a Step Functions Activity.
---

# Data Source: aws_sfn_activity

Provides a Step Functions Activity data source

## Example Usage

```terraform
data "aws_sfn_activity" "sfn_activity" {
  name = "my-activity"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Optional) Name that identifies the activity.
* `arn` - (Optional) ARN that identifies the activity.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ARN that identifies the activity.
* `creation_date` - Date the activity was created.
