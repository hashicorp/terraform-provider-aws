---
subcategory: "SNS (Simple Notification)"
layout: "aws"
page_title: "AWS: aws_sns_platform_application"
description: |-
  Provides an SNS platform application resource.
---

# Resource: aws_sns_platform_application

Provides an SNS platform application resource

## Example Usage

### Apple Push Notification Service (APNS) using certificate-based authentication

```terraform
resource "aws_sns_platform_application" "apns_application" {
  name                = "apns_application"
  platform            = "APNS"
  platform_credential = "<APNS PRIVATE KEY>"
  platform_principal  = "<APNS CERTIFICATE>"
}
```

### Apple Push Notification Service (APNS) using token-based authentication

```terraform
resource "aws_sns_platform_application" "apns_application" {
  name                     = "apns_application"
  platform                 = "APNS"
  platform_credential      = "<APNS SIGNING KEY>"
  platform_principal       = "<APNS SIGNING KEY ID>"
  apple_platform_team_id   = "<APPLE TEAM ID>"
  apple_platform_bundle_id = "<APPLE BUNDLE ID>"
}
```

### Google Cloud Messaging (GCM)

```terraform
resource "aws_sns_platform_application" "gcm_application" {
  name                = "gcm_application"
  platform            = "GCM"
  platform_credential = "<GCM API KEY>"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The friendly name for the SNS platform application
* `platform` - (Required) The platform that the app is registered with. See [Platform][1] for supported platforms.
* `platform_credential` - (Required) Application Platform credential. See [Credential][1] for type of credential required for platform. The value of this attribute when stored into the Terraform state is only a hash of the real value, so therefore it is not practical to use this as an attribute for other resources.
* `event_delivery_failure_topic_arn` - (Optional) The ARN of the SNS Topic triggered when a delivery to any of the platform endpoints associated with your platform application encounters a permanent failure.
* `event_endpoint_created_topic_arn` - (Optional) The ARN of the SNS Topic triggered when a new platform endpoint is added to your platform application.
* `event_endpoint_deleted_topic_arn` - (Optional) The ARN of the SNS Topic triggered when an existing platform endpoint is deleted from your platform application.
* `event_endpoint_updated_topic_arn` - (Optional) The ARN of the SNS Topic triggered when an existing platform endpoint is changed from your platform application.
* `failure_feedback_role_arn` - (Optional) The IAM role ARN permitted to receive failure feedback for this application and give SNS write access to use CloudWatch logs on your behalf.
* `platform_principal` - (Optional) Application Platform principal. See [Principal][2] for type of principal required for platform. The value of this attribute when stored into the Terraform state is only a hash of the real value, so therefore it is not practical to use this as an attribute for other resources.
* `success_feedback_role_arn` - (Optional) The IAM role ARN permitted to receive success feedback for this application and give SNS write access to use CloudWatch logs on your behalf.
* `success_feedback_sample_rate` - (Optional) The sample rate percentage (0-100) of successfully delivered messages.

The following attributes are needed only when using APNS token credentials:

* `apple_platform_team_id` - (Required) The identifier that's assigned to your Apple developer account team. Must be 10 alphanumeric characters.
* `apple_platform_bundle_id` - (Required) The bundle identifier that's assigned to your iOS app. May only include alphanumeric characters, hyphens (-), and periods (.).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ARN of the SNS platform application
* `arn` - The ARN of the SNS platform application

[1]: http://docs.aws.amazon.com/sns/latest/dg/mobile-push-send-register.html
[2]: http://docs.aws.amazon.com/sns/latest/api/API_CreatePlatformApplication.html

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SNS platform applications using the ARN. For example:

```terraform
import {
  to = aws_sns_platform_application.gcm_application
  id = "arn:aws:sns:us-west-2:123456789012:app/GCM/gcm_application"
}
```

Using `terraform import`, import SNS platform applications using the ARN. For example:

```console
% terraform import aws_sns_platform_application.gcm_application arn:aws:sns:us-west-2:123456789012:app/GCM/gcm_application
```
