---
subcategory: "Kendra"
layout: "aws"
page_title: "AWS: aws_kendra_experience"
description: |-
  Provides details about a specific Amazon Kendra Experience.
---

# Data Source: aws_kendra_experience

Provides details about a specific Amazon Kendra Experience.

## Example Usage

```hcl
data "aws_kendra_experience" "example" {
  experience_id = "87654321-1234-4321-4321-321987654321"
  index_id      = "12345678-1234-1234-1234-123456789123"
}
```

## Argument Reference

This data source supports the following arguments:

* `experience_id` - (Required) Identifier of the Experience.
* `index_id` - (Required) Identifier of the index that contains the Experience.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Experience.
* `configuration` - Block that specifies the configuration information for your Amazon Kendra Experience. This includes `content_source_configuration`, which specifies the data source IDs and/or FAQ IDs, and `user_identity_configuration`, which specifies the user or group information to grant access to your Amazon Kendra Experience. Documented below.
* `created_at` - Unix datetime that the Experience was created.
* `description` - Description of the Experience.
* `endpoints` - Shows the endpoint URLs for your Amazon Kendra Experiences. The URLs are unique and fully hosted by AWS. Documented below.
* `error_message` - Reason your Amazon Kendra Experience could not properly process.
* `id` - Unique identifiers of the Experience and index separated by a slash (`/`).
* `name` - Name of the Experience.
* `role_arn` - Shows the ARN of a role with permission to access `Query` API, `QuerySuggestions` API, `SubmitFeedback` API, and AWS SSO that stores your user and group information.
* `status` - Current processing status of your Amazon Kendra Experience. When the status is `ACTIVE`, your Amazon Kendra Experience is ready to use. When the status is `FAILED`, the `error_message` field contains the reason that this failed.
* `updated_at` - Date and time that the Experience was last updated.

The `configuration` block supports the following attributes:

* `content_source_configuration` - The identifiers of your data sources and FAQs. This is the content you want to use for your Amazon Kendra Experience. Documented below.
* `user_identity_configuration` - The AWS SSO field name that contains the identifiers of your users, such as their emails. Documented below.

The `content_source_configuration` block supports the following attributes:

* `data_source_ids` - Identifiers of the data sources you want to use for your Amazon Kendra Experience.
* `direct_put_content` - Whether to use documents you indexed directly using the `BatchPutDocument API`.
* `faq_ids` - Identifier of the FAQs that you want to use for your Amazon Kendra Experience.

The `user_identity_configuration` block supports the following attributes:

* `identity_attribute_name` - The AWS SSO field name that contains the identifiers of your users, such as their emails.

The `endpoints` block supports the following attributes:

* `endpoint` - Endpoint of your Amazon Kendra Experience.
* `endpoint_type` - Type of endpoint for your Amazon Kendra Experience.
