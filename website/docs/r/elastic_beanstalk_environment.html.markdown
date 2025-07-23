---
subcategory: "Elastic Beanstalk"
layout: "aws"
page_title: "AWS: aws_elastic_beanstalk_environment"
description: |-
  Provides an Elastic Beanstalk Environment Resource
---

# Resource: aws_elastic_beanstalk_environment

Provides an Elastic Beanstalk Environment Resource. Elastic Beanstalk allows
you to deploy and manage applications in the AWS cloud without worrying about
the infrastructure that runs those applications.

Environments are often things such as `development`, `integration`, or
`production`.

## Example Usage

```terraform
resource "aws_elastic_beanstalk_application" "tftest" {
  name        = "tf-test-name"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tfenvtest" {
  name                = "tf-test-name"
  application         = aws_elastic_beanstalk_application.tftest.name
  solution_stack_name = "64bit Amazon Linux 2015.03 v2.0.3 running Go 1.4"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) A unique name for this Environment. This name is used
  in the application URL
* `application` - (Required) Name of the application that contains the version
  to be deployed
* `cname_prefix` - (Optional) Prefix to use for the fully qualified DNS name of
  the Environment.
* `description` - (Optional) Short description of the Environment
* `tier` - (Optional) Elastic Beanstalk Environment tier. Valid values are `Worker`
  or `WebServer`. If tier is left blank `WebServer` will be used.
* `setting` - (Optional) Option settings to configure the new Environment. These
  override specific values that are set as defaults. The format is detailed
  below in [Option Settings](#option-settings)
* `solution_stack_name` - (Optional) A solution stack to base your environment
off of. Example stacks can be found in the [Amazon API documentation][1]
* `template_name` - (Optional) The name of the Elastic Beanstalk Configuration
  template to use in deployment
* `platform_arn` - (Optional) The [ARN][2] of the Elastic Beanstalk [Platform][3]
  to use in deployment
* `wait_for_ready_timeout` - (Default `20m`) The maximum
  [duration](https://golang.org/pkg/time/#ParseDuration) that Terraform should
  wait for an Elastic Beanstalk Environment to be in a ready state before timing
  out.
* `poll_interval` - The time between polling the AWS API to
check if changes have been applied. Use this to adjust the rate of API calls
for any `create` or `update` action. Minimum `10s`, maximum `180s`. Omit this to
use the default behavior, which is an exponential backoff
* `version_label` - (Optional) The name of the Elastic Beanstalk Application Version
to use in deployment.
* `tags` - (Optional) A set of tags to apply to the Environment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Option Settings

Some options can be stack-specific, check [AWS Docs](https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/command-options-general.html)
for supported options and examples.

The `setting` and `all_settings` mappings support the following format:

* `namespace` - unique namespace identifying the option's associated AWS resource
* `name` - name of the configuration option
* `value` - value for the configuration option
* `resource` - (Optional) resource name for [scheduled action](https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/command-options-general.html#command-options-general-autoscalingscheduledaction)

### Example With Options

```terraform
resource "aws_elastic_beanstalk_application" "tftest" {
  name        = "tf-test-name"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_environment" "tfenvtest" {
  name                = "tf-test-name"
  application         = aws_elastic_beanstalk_application.tftest.name
  solution_stack_name = "64bit Amazon Linux 2015.03 v2.0.3 running Go 1.4"

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = "vpc-xxxxxxxx"
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = "subnet-xxxxxxxx"
  }
}
```

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the Elastic Beanstalk Environment.
* `name` - Name of the Elastic Beanstalk Environment.
* `description` - Description of the Elastic Beanstalk Environment.
* `tier` - The environment tier specified.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `application` - The Elastic Beanstalk Application specified for this environment.
* `setting` - Settings specifically set for this Environment.
* `all_settings` - List of all option settings configured in this Environment. These
  are a combination of default settings and their overrides from `setting` in
  the configuration.
* `cname` - Fully qualified DNS name for this Environment.
* `autoscaling_groups` - The autoscaling groups used by this Environment.
* `instances` - Instances used by this Environment.
* `launch_configurations` - Launch configurations in use by this Environment.
* `load_balancers` - Elastic load balancers in use by this Environment.
* `queues` - SQS queues in use by this Environment.
* `triggers` - Autoscaling triggers in use by this Environment.
* `endpoint_url` - The URL to the Load Balancer for this Environment

[1]: https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/concepts.platforms.html
[2]: https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html
[3]: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-beanstalk-environment.html#cfn-beanstalk-environment-platformarn

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Elastic Beanstalk Environments using the `id`. For example:

```terraform
import {
  to = aws_elastic_beanstalk_environment.prodenv
  id = "e-rpqsewtp2j"
}
```

Using `terraform import`, import Elastic Beanstalk Environments using the `id`. For example:

```console
% terraform import aws_elastic_beanstalk_environment.prodenv e-rpqsewtp2j
```
