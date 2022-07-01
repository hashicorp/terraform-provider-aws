---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_container_service_deployment_version"
description: |-
  Provides a resource to manage a deployment version for your Amazon Lightsail container service.
---

# Resource: aws_lightsail_container_service_deployment_version

Provides a resource to manage a deployment version for your Amazon Lightsail container service.

~> **NOTE:** The Amazon Lightsail container service must be enabled to create a deployment.

~> **NOTE:** This resource allows you to manage an Amazon Lightsail container service deployment version but Terraform cannot destroy it. Removing this resource from your configuration will remove it from your statefile and Terraform management.

## Example Usage

### Basic Usage

```terraform
resource "aws_lightsail_container_service_deployment_version" "example" {
  container {
    container_name = "hello-world"
    image          = "amazon/amazon-lightsail:hello-world"

    command = []

    environment = {
      MY_ENVIRONMENT_VARIABLE = "my_value"
    }

    ports = {
      80 = "HTTP"
    }
  }

  public_endpoint {
    container_name = "hello-world"
    container_port = 80

    health_check {
      healthy_threshold   = 2
      unhealthy_threshold = 2
      timeout_seconds     = 2
      interval_seconds    = 5
      path                = "/"
      success_codes       = "200-499"
    }
  }

  service_name = aws_lightsail_container_service.example.name
}
```

## Argument Reference

The following arguments are supported:

* `service_name` - (Required) The name for the container service.
* `container` - (Required) A set of configuration blocks that describe the settings of the containers that will be launched on the container service. Maximum of 53. [Detailed below](#container).
* `public_endpoint` - (Optional) A configuration block that describes the settings of the public endpoint for the container service. [Detailed below](#public_endpoint).

### `container`

The `container` configuration block supports the following arguments:

* `container_name` - (Required) The name for the container.
* `image` - (Required) The name of the image used for the container. Container images sourced from your Lightsail container service, that are registered and stored on your service, start with a colon (`:`). For example, `:container-service-1.mystaticwebsite.1`. Container images sourced from a public registry like Docker Hub don't start with a colon. For example, `nginx:latest` or `nginx`.
* `command` - (Optional) The launch command for the container. A list of string.
* `environment` - (Optional) A key-value map of the environment variables of the container.
* `ports` - (Optional) A key-value map of the open firewall ports of the container. Valid values: `HTTP`, `HTTPS`, `TCP`, `UDP`.

### `public_endpoint`

The `public_endpoint` configuration block supports the following arguments:

* `container_name` - (Required) The name of the container for the endpoint.
* `container_port` - (Required) The port of the container to which traffic is forwarded to.
* `health_check` - (Required) A configuration block that describes the health check configuration of the container. [Detailed below](#health_check).

### `health_check`

The `health_check` configuration block supports the following arguments:

* `healthy_threshold` - (Optional) The number of consecutive health checks successes required before moving the container to the Healthy state. Defaults to 2.
* `unhealthy_threshold` - (Optional) The number of consecutive health checks failures required before moving the container to the Unhealthy state. Defaults to 2.
* `timeout_seconds` - (Optional) The amount of time, in seconds, during which no response means a failed health check. You can specify between 2 and 60 seconds. Defaults to 2.
* `interval_seconds` - (Optional) The approximate interval, in seconds, between health checks of an individual container. You can specify between 5 and 300 seconds. Defaults to 5.
* `path` - (Optional) The path on the container on which to perform the health check. Defaults to "/".
* `success_codes` - (Optional) The HTTP codes to use when checking for a successful response from a container. You can specify values between 200 and 499. Defaults to "200-499".

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The `service_name` and `version` separation by a slash (`/`).
* `created_at` - The timestamp when the deployment was created.
* `state` - The current state of the container service.
* `version` - The version number of the deployment.

## Timeouts

`aws_lightsail_container_service_deployment_version` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Optional, Default: `30m`)

## Import

Lightsail Container Service Deployment Version can be imported using the `service_name` and `version` separated by a slash (`/`), e.g.,

```shell
$ terraform import aws_lightsail_container_service_deployment_version.example container-service-1/1
```
