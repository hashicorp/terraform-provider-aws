---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_container_service"
description: |- 
  Provides a resource to manage Lightsail container service
---

# Resource: aws_lightsail_container_service

An Amazon Lightsail container service is a highly scalable compute and networking resource on which you can deploy, run,
and manage containers. For more information, see
[Container services in Amazon Lightsail](https://lightsail.aws.amazon.com/ls/docs/en_us/articles/amazon-lightsail-container-services).

~> **Note:** For more information about the AWS Regions in which you can create Amazon Lightsail container services,
see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail).

## Example Usage

### Basic Usage

```terraform
resource "aws_lightsail_container_service" "my_container_service" {
  name        = "container-service-1"
  power       = "nano"
  scale       = 1
  is_disabled = false

  tags = {
    foo1 = "bar1"
    foo2 = ""
  }
}
```

### Public Domain Names

```terraform
resource "aws_lightsail_container_service" "my_container_service" {
  # ... other configuration ...

  public_domain_names {
    certificate {
      certificate_name = "example-certificate"
      domain_names = [
        "www.example.com",
        # maybe another domain name
      ]
    }

    # certificate {
    # maybe another certificate
    # }
  }
}
```

## Argument Reference

~> **NOTE:** You must create and validate an SSL/TLS certificate before you can use `public_domain_names` with your
container service. For more information, see
[Enabling and managing custom domains for your Amazon Lightsail container services](https://lightsail.aws.amazon.com/ls/docs/en_us/articles/amazon-lightsail-creating-container-services-certificates).

The following arguments are supported:

* `name` - (Required) The name for the container service. Names must be of length 1 to 63, and be
  unique within each AWS Region in your Lightsail account.
* `power` - (Required) The power specification for the container service. The power specifies the amount of memory,
  the number of vCPUs, and the monthly price of each node of the container service.
  Possible values: `nano`, `micro`, `small`, `medium`, `large`, `xlarge`.
* `scale` - (Required) The scale specification for the container service. The scale specifies the allocated compute
  nodes of the container service.
* `is_disabled` - (Optional) A Boolean value indicating whether the container service is disabled. Defaults to `false`.
* `public_domain_names` - (Optional) The public domain names to use with the container service, such as example.com
  and www.example.com. You can specify up to four public domain names for a container service. The domain names that you
  specify are used when you create a deployment with a container configured as the public endpoint of your container
  service. If you don't specify public domain names, then you can use the default domain of the container service.
  Defined below.
* `tags` - (Optional) Map of container service tags. To tag at launch, specify the tags in the Launch Template. If
  configured with a provider
  [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block)
  present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the container service.
* `availability_zone` - The Availability Zone. Follows the format us-east-2a (case-sensitive).
* `id` - Same as `name`.
* `power_id` - The ID of the power of the container service.
* `principal_arn`- The principal ARN of the container service. The principal ARN can be used to create a trust
  relationship between your standard AWS account and your Lightsail container service. This allows you to give your
  service permission to access resources in your standard AWS account.
* `private_domain_name` - The private domain name of the container service. The private domain name is accessible only
  by other resources within the default virtual private cloud (VPC) of your Lightsail account.
* `region_name` - The AWS Region name.
* `resource_type` - The Lightsail resource type of the container service (i.e., ContainerService).
* `state` - The current state of the container service.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider
  [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).
* `url` - The publicly accessible URL of the container service. If no public endpoint is specified in the
  currentDeployment, this URL returns a 404 response.

## Timeouts

`aws_lightsail_container_service` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Optional, Default: `30m`)
* `update` - (Optional, Default: `30m`)
* `delete` - (Optional, Default: `30m`)

## Import

Lightsail Container Service can be imported using the `name`, e.g.,

```shell
$ terraform import aws_lightsail_container_service.my_container_service container-service-1
```
