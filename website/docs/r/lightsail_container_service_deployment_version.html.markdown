---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_container_service_deployment_version"
description: |-
  Manages a deployment version for a Lightsail container service.
---

# Resource: aws_lightsail_container_service_deployment_version

Manages a Lightsail container service deployment version. Use this resource to deploy containerized applications to your Lightsail container service with specific container configurations and settings.

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

The following arguments are required:

* `container` - (Required) Set of configuration blocks that describe the settings of the containers that will be launched on the container service. Maximum of 53. [See below](#container).
* `service_name` - (Required) Name of the container service.

The following arguments are optional:

* `public_endpoint` - (Optional) Configuration block that describes the settings of the public endpoint for the container service. [See below](#public_endpoint).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `container`

The `container` configuration block supports the following arguments:

* `command` - (Optional) Launch command for the container. A list of strings.
* `container_name` - (Required) Name of the container.
* `environment` - (Optional) Key-value map of the environment variables of the container.
* `image` - (Required) Name of the image used for the container. Container images sourced from your Lightsail container service, that are registered and stored on your service, start with a colon (`:`). For example, `:container-service-1.mystaticwebsite.1`. Container images sourced from a public registry like Docker Hub don't start with a colon. For example, `nginx:latest` or `nginx`.
* `ports` - (Optional) Key-value map of the open firewall ports of the container. Valid values: `HTTP`, `HTTPS`, `TCP`, `UDP`.

### `public_endpoint`

The `public_endpoint` configuration block supports the following arguments:

* `container_name` - (Required) Name of the container for the endpoint.
* `container_port` - (Required) Port of the container to which traffic is forwarded to.
* `health_check` - (Required) Configuration block that describes the health check configuration of the container. [See below](#health_check).

### `health_check`

The `health_check` configuration block supports the following arguments:

* `healthy_threshold` - (Optional) Number of consecutive health check successes required before moving the container to the Healthy state. Defaults to 2.
* `interval_seconds` - (Optional) Approximate interval, in seconds, between health checks of an individual container. You can specify between 5 and 300 seconds. Defaults to 5.
* `path` - (Optional) Path on the container on which to perform the health check. Defaults to "/".
* `success_codes` - (Optional) HTTP codes to use when checking for a successful response from a container. You can specify values between 200 and 499. Defaults to "200-499".
* `timeout_seconds` - (Optional) Amount of time, in seconds, during which no response means a failed health check. You can specify between 2 and 60 seconds. Defaults to 2.
* `unhealthy_threshold` - (Optional) Number of consecutive health check failures required before moving the container to the Unhealthy state. Defaults to 2.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - Date and time when the deployment was created.
* `id` - `service_name` and `version` separated by a slash (`/`).
* `state` - Current state of the container service.
* `version` - Version number of the deployment.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lightsail Container Service Deployment Version using the `service_name` and `version` separated by a slash (`/`). For example:

```terraform
import {
  to = aws_lightsail_container_service_deployment_version.example
  id = "container-service-1/1"
}
```

Using `terraform import`, import Lightsail Container Service Deployment Version using the `service_name` and `version` separated by a slash (`/`). For example:

```console
% terraform import aws_lightsail_container_service_deployment_version.example container-service-1/1
```
