---
subcategory: "Elastic Beanstalk"
layout: "aws"
page_title: "AWS: aws_elastic_beanstalk_configuration_template"
description: |-
  Provides an Elastic Beanstalk Configuration Template
---

# Resource: aws_elastic_beanstalk_configuration_template

Provides an Elastic Beanstalk Configuration Template, which are associated with
a specific application and are used to deploy different versions of the
application with the same configuration settings.

## Example Usage

```terraform
resource "aws_elastic_beanstalk_application" "tftest" {
  name        = "tf-test-name"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_configuration_template" "tf_template" {
  name                = "tf-test-template-config"
  application         = aws_elastic_beanstalk_application.tftest.name
  solution_stack_name = "64bit Amazon Linux 2015.09 v2.0.8 running Go 1.4"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) A unique name for this Template.
* `application` - (Required) name of the application to associate with this configuration template
* `description` - (Optional) Short description of the Template
* `environment_id` - (Optional) The ID of the environment used with this configuration template
* `setting` - (Optional) Option settings to configure the new Environment. These
  override specific values that are set as defaults. The format is detailed
  below in [Option Settings](#option-settings)
* `solution_stack_name` - (Optional) A solution stack to base your Template
off of. Example stacks can be found in the [Amazon API documentation][1]

## Option Settings

The `setting` field supports the following format:

* `namespace` - unique namespace identifying the option's associated AWS resource
* `name` - name of the configuration option
* `value` - value for the configuration option
* `resource` - (Optional) resource name for [scheduled action](https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/command-options-general.html#command-options-general-autoscalingscheduledaction)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `name`
* `application`
* `description`
* `environment_id`
* `option_settings`
* `solution_stack_name`

[1]: https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/concepts.platforms.html
