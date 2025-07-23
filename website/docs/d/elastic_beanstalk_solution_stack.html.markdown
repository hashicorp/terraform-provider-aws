---
subcategory: "Elastic Beanstalk"
layout: "aws"
page_title: "AWS: aws_elastic_beanstalk_solution_stack"
description: |-
  Get an elastic beanstalk solution stack.
---

# Data Source: aws_elastic_beanstalk_solution_stack

Use this data source to get the name of a elastic beanstalk solution stack.

## Example Usage

```terraform
data "aws_elastic_beanstalk_solution_stack" "multi_docker" {
  most_recent = true

  name_regex = "^64bit Amazon Linux (.*) Multi-container Docker (.*)$"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `most_recent` - (Optional) If more than one result is returned, use the most
recent solution stack.
* `name_regex` - Regex string to apply to the solution stack list returned
by AWS. See [Elastic Beanstalk Supported Platforms][beanstalk-platforms] from
AWS documentation for reference solution stack names.

~> **NOTE:** If more or less than a single match is returned by the search,
Terraform will fail. Ensure that your search is specific enough to return
a single solution stack, or use `most_recent` to choose the most recent one.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `name` - Name of the solution stack.

[beanstalk-platforms]: http://docs.aws.amazon.com/elasticbeanstalk/latest/dg/concepts.platforms.html "AWS Elastic Beanstalk Supported Platforms documentation"
