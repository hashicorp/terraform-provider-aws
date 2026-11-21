---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_portal"
description: |-
  Manages an Amazon API Gateway portal.
---

# Resource: aws_apigatewayv2_portal

Manages an Amazon API Gateway portal. A portal is a developer portal that presents portal products and their REST API documentation to API consumers.

More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-portals.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_apigatewayv2_portal" "example" {
  authorization {
    none {}
  }

  endpoint_configuration {
    none {}
  }

  portal_content {
    display_name = "Example portal"

    theme {
      custom_colors {
        accent_color           = "#FF9900"
        background_color       = "#FFFFFF"
        error_validation_color = "#D13212"
        header_color           = "#232F3E"
        navigation_color       = "#232F3E"
        text_color             = "#16191F"
      }
    }
  }
}
```

### Published Portal with Amazon Cognito Sign-In and a Custom Domain

```terraform
resource "aws_apigatewayv2_portal" "example" {
  authorization {
    cognito_config {
      app_client_id    = aws_cognito_user_pool_client.example.id
      user_pool_arn    = aws_cognito_user_pool.example.arn
      user_pool_domain = "${aws_cognito_user_pool_domain.example.domain}.auth.us-east-1.amazoncognito.com"
    }
  }

  endpoint_configuration {
    acm_managed {
      certificate_arn = aws_acm_certificate_validation.example.certificate_arn
      domain_name     = "developer.example.com"
    }
  }

  portal_content {
    display_name = "Example portal"
    description  = "APIs for example.com partners"

    theme {
      custom_colors {
        accent_color           = "#FF9900"
        background_color       = "#FFFFFF"
        error_validation_color = "#D13212"
        header_color           = "#232F3E"
        navigation_color       = "#232F3E"
        text_color             = "#16191F"
      }
    }
  }

  publish             = true
  publish_description = "Initial release"
}

resource "aws_route53_record" "example" {
  zone_id = aws_route53_zone.example.zone_id
  name    = "developer.example.com"
  type    = "CNAME"
  ttl     = 300
  records = [aws_apigatewayv2_portal.example.portal_default_domain_name]
}
```

## Argument Reference

The following arguments are required:

* `authorization` - (Required) How users sign in to the portal. See [`authorization`](#authorization) below.
* `endpoint_configuration` - (Required) Domain the portal is served from. See [`endpoint_configuration`](#endpoint_configuration) below.
* `portal_content` - (Required) Content and theme of the portal. See [`portal_content`](#portal_content) below.

The following arguments are optional:

* `included_portal_product_arns` - (Optional) Set of ARNs of the portal products to include in the portal. Portal products shared from other accounts can be included.
* `logo_uri` - (Optional) Amazon S3 URI of the logo shown in the portal header, for example `s3://example-bucket/logo.png`. The object must be a JPEG, PNG or SVG image with a matching content type. API Gateway does not return this value, so changes made outside Terraform are not detected.
* `publish` - (Optional) Whether to publish the portal. Defaults to `false`. A published portal is reachable at its domain. Changes to a published portal are not visible until the portal is published again, so Terraform publishes the portal again whenever its configuration changes while `publish` is `true`. A published portal is disabled before it is deleted.
* `publish_description` - (Optional) Description of the publication. Changing this argument while `publish` is `true` publishes the portal again.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `rum_app_monitor_name` - (Optional) Name of the Amazon CloudWatch RUM app monitor that collects portal usage.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `authorization`

Exactly one of the following must be specified:

* `cognito_config` - (Optional) Amazon Cognito user pool that portal users sign in with. See [`cognito_config`](#cognito_config) below.
* `none` - (Optional) Empty block. Portal users do not sign in.

### `cognito_config`

* `app_client_id` - (Required) ID of the user pool app client.
* `user_pool_arn` - (Required) ARN of the user pool.
* `user_pool_domain` - (Required) Domain of the user pool, without the scheme, for example `example.auth.us-east-1.amazoncognito.com`.

### `endpoint_configuration`

Exactly one of the following must be specified:

* `acm_managed` - (Optional) Custom domain name for the portal. See [`acm_managed`](#acm_managed) below.
* `none` - (Optional) Empty block. The portal is served only from its default domain.

### `acm_managed`

* `certificate_arn` - (Required) ARN of the AWS Certificate Manager certificate for the domain name. The certificate must be in the `us-east-1` Region.
* `domain_name` - (Required) Custom domain name of the portal.

### `portal_content`

* `description` - (Optional) Description of the portal.
* `display_name` - (Required) Name shown in the portal. Between 3 and 255 characters.
* `theme` - (Required) Theme of the portal. See [`theme`](#theme) below.

### `theme`

* `custom_colors` - (Required) Colors of the portal. See [`custom_colors`](#custom_colors) below.

### `custom_colors`

All colors are required.

* `accent_color` - (Required) Accent color.
* `background_color` - (Required) Background color.
* `error_validation_color` - (Required) Color of validation errors.
* `header_color` - (Required) Header color.
* `navigation_color` - (Required) Navigation color.
* `text_color` - (Required) Text color.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `last_modified` - Time the portal was last modified, in RFC3339 format.
* `last_published` - Time the portal was last published, in RFC3339 format.
* `last_published_description` - Description of the last publication.
* `portal_arn` - ARN of the portal.
* `portal_default_domain_name` - Default domain name of the portal. For a custom domain name, point a `CNAME` record at it.
* `portal_domain_hosted_zone_id` - ID of the Route 53 hosted zone of the portal's default domain name.
* `portal_id` - ID of the portal.
* `publish_status` - Publication status of the portal. Valid values: `PUBLISHED`, `PUBLISH_IN_PROGRESS`, `PUBLISH_FAILED`, `DISABLE_IN_PROGRESS`, `DISABLE_FAILED`, `DISABLED`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_apigatewayv2_portal.example
  identity = {
    portal_id = "abcdef1234"
  }
}

resource "aws_apigatewayv2_portal" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `portal_id` (String) ID of the portal.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import API Gateway portals using `portal_id`. For example:

```terraform
import {
  to = aws_apigatewayv2_portal.example
  id = "abcdef1234"
}
```

Using `terraform import`, import API Gateway portals using `portal_id`. For example:

```console
% terraform import aws_apigatewayv2_portal.example abcdef1234
```
