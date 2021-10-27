---
subcategory: "Elastic Beanstalk"
layout: "aws"
page_title: "AWS: aws_elastic_beanstalk_environment"
description: |-
  Retrieve information about an Elastic Beanstalk Environment
---

# Data Source: aws_elastic_beanstalk_environment

Retrieve information about an Elastic Beanstalk Environment.

## Example Usage

```terraform
data "aws_elastic_beanstalk_environment" "example" {
  name = "my-app"
}
output "endpoint_url" {
  value = data.aws_elastic_beanstalk_environment.example.endpoint_url
}
```

## Argument Reference

* `name` - (Required) The name of the environment.

## Attributes Reference

* `arn` -  The Amazon Resource Name (ARN) of the environment.
* `application` - The application to which the environment belongs.
* `version_label` - The version of the environment.
* `solution_stack_name` - The name of the SolutionStack deployed with the environment.
* `platform_arn` - The ARN of the platform version.
* `template_name` - The name of the configuration template used to originally launch the environment.
* `description` - The description of the environment.
* `endpoint_url` - For load-balanced, autoscaling environments, the URL to the LoadBalancer. For single-instance environments, the IP address of the instance.
* `cname` - The URL to the CNAME for the environment.
* `resources` - The description of the AWS resources used by this environment.
* `tier` - The tier of the environment.