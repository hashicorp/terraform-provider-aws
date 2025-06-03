---
subcategory: "Elastic Beanstalk"
layout: "aws"
page_title: "AWS: aws_elastic_beanstalk_application"
description: |-
  Retrieve information about an Elastic Beanstalk Application
---

# Data Source: aws_elastic_beanstalk_application

Retrieve information about an Elastic Beanstalk Application.

## Example Usage

```terraform
data "aws_elastic_beanstalk_application" "example" {
  name = "example"
}

output "arn" {
  value = data.aws_elastic_beanstalk_application.example.arn
}

output "description" {
  value = data.aws_elastic_beanstalk_application.example.description
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the application

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Name of the application
* `arn` - ARN of the application.
* `description` - Short description of the application

Application version lifecycle (`appversion_lifecycle`) supports the nested attribute containing.

* `service_role` - ARN of an IAM service role under which the application version is deleted.  Elastic Beanstalk must have permission to assume this role.
* `max_count` - Maximum number of application versions to retain.
* `max_age_in_days` - Number of days to retain an application version.
* `delete_source_from_s3` - Specifies whether delete a version's source bundle from S3 when the application version is deleted.
