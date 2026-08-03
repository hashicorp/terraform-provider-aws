---
subcategory: "App Runner"
layout: "aws"
page_title: "AWS: aws_apprunner_vpc_ingress_connection"
description: |-
  Manages an App Runner VPC Ingress Connection.
---

# Resource: aws_apprunner_vpc_ingress_connection

Manages an App Runner VPC Ingress Connection.

## Example Usage

```terraform
resource "aws_apprunner_vpc_ingress_connection" "example" {
  name        = "example"
  service_arn = aws_apprunner_service.example.arn

  ingress_vpc_configuration {
    vpc_id          = aws_default_vpc.default.id
    vpc_endpoint_id = aws_vpc_endpoint.apprunner.id
  }

  tags = {
    foo = "bar"
  }
}

```

## Argument Reference

This resource supports the following arguments:

* `ingress_vpc_configuration` - (Required) Specifications for the customer’s Amazon VPC and the related AWS PrivateLink VPC endpoint that are used to create the VPC Ingress Connection resource. See [`ingress_vpc_configuration` Block](#ingress_vpc_configuration-block) below for more details.
* `name` - (Required) Name for the VPC Ingress Connection resource. It must be unique across all the active VPC Ingress Connections in your AWS account in the AWS Region.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `service_arn` - (Required) Amazon Resource Name (ARN) for this App Runner service that is used to create the VPC Ingress Connection resource.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `ingress_vpc_configuration` Block

The `ingress_vpc_configuration` block supports the following argument:

* `vpc_endpoint_id` - (Required) ID of the VPC endpoint that your App Runner service connects to.
* `vpc_id` - (Required) ID of the VPC that is used for the VPC endpoint.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the VPC Ingress Connection.
* `domain_name` - Domain name associated with the VPC Ingress Connection resource.
* `status` - Current status of the VPC Ingress Connection.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_apprunner_vpc_ingress_connection.example
  identity = {
    "arn" = "arn:aws:apprunner:us-east-1:123456789012:vpcingressconnection/example-vpc-ingress-connection/a1b2c3d4567890ab"
  }
}

resource "aws_apprunner_vpc_ingress_connection" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) Amazon Resource Name (ARN) of the App Runner VPC ingress connection.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import App Runner VPC Ingress Connection using the `arn`. For example:

```terraform
import {
  to = aws_apprunner_vpc_ingress_connection.example
  id = "arn:aws:apprunner:us-west-2:837424938642:vpcingressconnection/example/b379f86381d74825832c2e82080342fa"
}
```

Using `terraform import`, import App Runner VPC Ingress Connection using the `arn`. For example:

```console
% terraform import aws_apprunner_vpc_ingress_connection.example "arn:aws:apprunner:us-west-2:837424938642:vpcingressconnection/example/b379f86381d74825832c2e82080342fa"
```
