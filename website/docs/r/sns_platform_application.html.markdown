---
subcategory: "SNS (Simple Notification)"
layout: "aws"
page_title: "AWS: aws_sns_platform_application"
description: |-
  Provides an SNS platform application resource.
---

# Resource: aws_sns_platform_application

Provides an SNS platform application resource

-> **Note:** Write-Only arguments `platform_credential_wo` and `platform_principal_wo` are available to use in place of `platform_credential` and `platform_principal`. Write-Only arguments are supported in HashiCorp Terraform 1.11.0 and later. [Learn more](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments).

## Example Usage

### Apple Push Notification Service (APNS) using certificate-based authentication

~> **NOTE:** For certificate-based APNS, both `platform_credential` (private key) and `platform_principal` (certificate) must be PEM-encoded strings. Terraform string values must be valid UTF-8, so do not pass a binary Apple `.p12` via `base64decode()` — that fails with `the result of decoding the provided string is not valid UTF-8`. Convert the `.p12` to PEM (for example with `openssl`) and load the PEM files instead.

```terraform
resource "aws_sns_platform_application" "apns_application" {
  name                = "apns_application"
  platform            = "APNS"
  platform_credential = file("apns-private-key.pem")
  platform_principal  = file("apns-certificate.pem")
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

### Google Cloud Messaging (GCM) using a write-only credential

In the following example, we use the write-only variant of `platform_credential`, `platform_credential_wo`. In this case, the credential
is read from an ephemeral resource and never ends up in the Terraform plan or state at any point. If an update is needed, incrementing the
`platform_credential_wo_version` will force re-evaluation of the secret (e.g, for rotation).

```terraform
ephemeral "aws_ssm_parameter" "fcm_service_account" {
  arn = "<ARN OF SSM PARAMETER CONTAINING FIREBASE SERVICE ACCOUNT JSON>"
}

resource "aws_sns_platform_application" "gcm_application" {
  name                           = "gcm_application"
  platform                       = "GCM"
  platform_credential_wo         = ephemeral.aws_ssm_parameter.fcm_service_account.value
  platform_credential_wo_version = 1
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The friendly name for the SNS platform application
* `platform` - (Required) The platform that the app is registered with. See [Platform](http://docs.aws.amazon.com/sns/latest/dg/mobile-push-send-register.html) for supported platforms.
* `platform_credential` - (Optional) Application Platform credential. Exactly one of `platform_credential` or `platform_credential_wo` must be set. See [Credential](http://docs.aws.amazon.com/sns/latest/dg/mobile-push-send-register.html) for type of credential required for platform. The value of this attribute when stored into the Terraform state is only a hash of the real value, so therefore it is not practical to use this as an attribute for other resources.
* `platform_credential_wo` - (Optional, Write-Only) Application Platform credential. Write-only values are never stored to state. Exactly one of `platform_credential` or `platform_credential_wo` must be set. Requires `platform_credential_wo_version`.
* `platform_credential_wo_version` - (Optional) Used together with `platform_credential_wo` to trigger an update. Increment this value when an update to `platform_credential_wo` is required.
* `event_delivery_failure_topic_arn` - (Optional) The ARN of the SNS Topic triggered when a delivery to any of the platform endpoints associated with your platform application encounters a permanent failure.
* `event_endpoint_created_topic_arn` - (Optional) The ARN of the SNS Topic triggered when a new platform endpoint is added to your platform application.
* `event_endpoint_deleted_topic_arn` - (Optional) The ARN of the SNS Topic triggered when an existing platform endpoint is deleted from your platform application.
* `event_endpoint_updated_topic_arn` - (Optional) The ARN of the SNS Topic triggered when an existing platform endpoint is changed from your platform application.
* `failure_feedback_role_arn` - (Optional) The IAM role ARN permitted to receive failure feedback for this application and give SNS write access to use CloudWatch logs on your behalf.
* `platform_principal` - (Optional) Application Platform principal. See [Principal](http://docs.aws.amazon.com/sns/latest/api/API_CreatePlatformApplication.html) for type of principal required for platform. The value of this attribute when stored into the Terraform state is only a hash of the real value, so therefore it is not practical to use this as an attribute for other resources.
* `platform_principal_wo` - (Optional, Write-Only) Application Platform principal. Write-only values are never stored to state. Conflicts with `platform_principal`. Requires `platform_principal_wo_version`.
* `platform_principal_wo_version` - (Optional) Used together with `platform_principal_wo` to trigger an update. Increment this value when an update to `platform_principal_wo` is required.
* `success_feedback_role_arn` - (Optional) The IAM role ARN permitted to receive success feedback for this application and give SNS write access to use CloudWatch logs on your behalf.
* `success_feedback_sample_rate` - (Optional) The sample rate percentage (0-100) of successfully delivered messages.

The following attributes are needed only when using APNS token credentials:

* `apple_platform_team_id` - (Required) The identifier that's assigned to your Apple developer account team. Must be 10 alphanumeric characters.
* `apple_platform_bundle_id` - (Required) The bundle identifier that's assigned to your iOS app. May only include alphanumeric characters, hyphens (-), and periods (.).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ARN of the SNS platform application
* `arn` - The ARN of the SNS platform application

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
