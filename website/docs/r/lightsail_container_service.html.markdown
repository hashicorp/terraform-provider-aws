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

**Note:** For more information about the AWS Regions in which you can create Amazon Lightsail container services,
see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail).

## Example Usage

```terraform
# create a new Lightsail container service
resource "aws_lightsail_container_service" "my_container_service" {
  name = "container-service-1"
  power = "nano"
  scale = 1
  is_disabled = false

  # deployment {
    # example below
  # }

  # public_domain_names {
    # example below
  # }

  tags = {
    foo1 = "bar1"
    foo2 = ""
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name for the container service. Names must be of length 1 to 63, and be
  unique within each AWS Region in your Lightsail account.
* `power` - (Required) The power specification for the container service. The power specifies the amount of memory,
  the number of vCPUs, and the monthly price of each node of the container service. 
  Possible values: `nano`, `micro`, `small`, `medium`, `large`, `xlarge`.
* `scale` - (Required) The scale specification for the container service. The scale specifies the allocated compute
  nodes of the container service.
* `is_disabled` - (Optional) A Boolean value indicating whether the container service is disabled. Defaults to `false`.
* `deployment` - (Optional) A deployment specifies the containers that will be launched on the container service and
  their settings, such as the ports to open, the environment variables to apply, and the launch command to run. It also
  specifies the container that will serve as the public endpoint of the deployment and its settings, such as the HTTP or
  HTTPS port to use, and the health check configuration. Defined below.
* `public_domain_names` - (Optional) The public domain names to use with the container service, such as example.com
  and www.example.com. You can specify up to four public domain names for a container service. The domain names that you
  specify are used when you create a deployment with a container configured as the public endpoint of your container
  service. If you don't specify public domain names, then you can use the default domain of the container service.
  Defined below.
    * WARNING: You must create and validate an SSL/TLS certificate before you can use public domain names with your
      container service.
      For more information, see
      [Enabling and managing custom domains for your Amazon Lightsail container services](https://lightsail.aws.amazon.com/ls/docs/en_us/articles/amazon-lightsail-creating-container-services-certificates).
* `tags` - (Optional) Map of container service tags. To tag at launch, specify the tags in the Launch Template. If
  configured with a provider
  [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block)
  present, tags with matching keys will overwrite those defined at the provider-level.

### deployment

#### Example Usage

```terraform
resource "aws_lightsail_container_service" "my_container_service" {
  # ... other configuration ...

  deployment {
    container {
      container_name = "hello-world"
      image = "amazon/amazon-lightsail:hello-world"

      command = []

      environment {
        key = "MY_ENVIRONMENT_VARIABLE"
        value = "my_value"
      }

      # environment {
        # maybe another environment variable
      # }

      port {
        port_number = 80
        protocol = "HTTP"
      }

      # port {
        # maybe another port
      # }
    }

    # container {
      # maybe another container
    # }

    public_endpoint {
      container_name = "hello-world"
      container_port = 80

      health_check {
        healthy_threshold = 2
        unhealthy_threshold = 2
        timeout_seconds = 2
        interval_seconds = 5
        path = "/"
        success_codes = "200-499"
      }
    }
  }
}
```

* `container` - (Required) Describes the configuration for a container of the deployment.
    * `container_name` - (Required) The name for the container.
    * `image` - (Required) The name of the image used for the container. Container images sourced from your Lightsail
      container service, that are registered and stored on your service, start with a colon `:`. For
      example, `:container-service-1.mystaticwebsite.1`. Container images sourced from a public registry like Docker Hub
      don't start with a colon. For example, nginx:latest or nginx.
    * `command` - (Optional) The launch command for the container. A list of string.
    * `environment` - (Optional) The environment variables of the container.
        * `key` - (Required) The environment variable name.
        * `value` - (Required) The environment variable value.
    * `port` - (Optional) The open firewall ports of the container.
        * `port_number` - (Required) The port number.
        * `protocol` - (Required) The protocol. Possible values: `HTTP`, `HTTPS`, `TCP`, `UDP`.
* `public_endpoint` - (Optional) Describes the public endpoint configuration of a deployment.
    * `container_name` - (Required) The name of the container for the endpoint.
    * `container_port` - (Required) The port of the container to which traffic is forwarded to.
    * `health_check` - (Required) Describes the health check configuration of the container.
        * `healthy_threshold` - (Optional) The number of consecutive health checks successes required before moving the
          container to the Healthy state. Defaults to `2`.
        * `unhealthy_threshold` - (Optional) The number of consecutive health checks failures required before moving the
          container to the Unhealthy state. Defaults to `2`.
        * `timeout_seconds` - (Optional) The amount of time, in seconds, during which no response means a failed health
          check. You can specify between 2 and 60 seconds. Defaults to `2`.
        * `interval_seconds` - (Optional) The approximate interval, in seconds, between health checks of an individual
          container. You can specify between 5 and 300 seconds. Defaults to `5`.
        * `path` - (Optional) The path on the container on which to perform the health check. Defaults to `"/"`.
        * `success_codes` - (Optional) The HTTP codes to use when checking for a successful response from a container.
          You can specify values between 200 and 499. Defaults to `"200-499"`.

### public_domain_names

#### Example Usage

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

* `certificate` - (Required) Describes the details of an SSL/TLS certificate for a container service.
    * `certificate_name` - (Required) The certificate's name.
    * `domain_names` - (Required) The domain names. A list of string.

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

Plus, if you have a deployment, the following attributes are exported under `deployment` block:

* `state` - The state of the deployment.
* `version` - The version number of the deployment.

## Import

`aws_lightsail_container_service` can be imported using their name, e.g.
`$ terraform import aws_lightsail_container_service.my_container_service container-service-1`
