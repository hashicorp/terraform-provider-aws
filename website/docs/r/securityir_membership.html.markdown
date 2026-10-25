---
subcategory: "Security Incident Response"
layout: "aws"
page_title: "AWS: aws_securityir_membership"
description: |-
  Manages an AWS Security Incident Response membership.
---

# Resource: aws_securityir_membership

Manages an AWS Security Incident Response membership.

~> **Note:** An AWS account or organization can have only one Security Incident Response membership.

~> **Note:** A membership that covers an organization must be managed from the organization's management account or the Security Incident Response delegated administrator account.

!> **Note:** Destroying this resource cancels the membership.

## Example Usage

### Basic Usage

```terraform
resource "aws_securityir_membership" "example" {
  membership_name = "example-membership"

  incident_response_team {
    name      = "Jane Doe"
    job_title = "Security Engineer"
    email     = "jane.doe@example.com"
  }

  incident_response_team {
    name      = "John Doe"
    job_title = "Security Manager"
    email     = "john.doe@example.com"
  }
}
```

### Organization-Wide Membership

```terraform
resource "aws_securityir_membership" "example" {
  membership_name           = "example-membership"
  cover_entire_organization = true

  incident_response_team {
    name      = "Jane Doe"
    job_title = "Security Engineer"
    email     = "jane.doe@example.com"

    communication_preferences = ["Case Created", "Case Closed"]
  }

  incident_response_team {
    name      = "John Doe"
    job_title = "Security Manager"
    email     = "john.doe@example.com"
  }

  opt_in_feature {
    feature_name = "Triage"
    is_enabled   = true
  }
}
```

## Argument Reference

The following arguments are required:

* `incident_response_team` - (Required) Incident response team members to contact during an incident. Between 2 and 10 members are allowed. See [`incident_response_team` Block](#incident_response_team-block) below.
* `membership_name` - (Required) Name of the membership. Must be between 3 and 50 characters.

The following arguments are optional:

* `cover_entire_organization` - (Optional) Whether to cover all existing and future accounts in the AWS Organization. Default value: `false`.
* `opt_in_feature` - (Optional) Opt-in features to enable or disable. At most 2 are allowed. Only the features configured here are tracked, and removing a block does not change the feature. See [`opt_in_feature` Block](#opt_in_feature-block) below.
* `organizational_units` - (Optional) Set of AWS Organizations organizational unit IDs covered by the membership. Each ID must match `ou-[0-9a-z]{4,32}-[a-z0-9]{8,32}`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `incident_response_team` Block

The `incident_response_team` block supports:

* `communication_preferences` - (Optional) Set of notification types the team member receives. Allowed values are: `Case Created`, `Case Updated`, `Case Acknowledged`, `Case Closed`, `Case Updated To Service Managed`, `Case Status Updated`, `Case Pending Customer Action Reminder`, `Case Attachment Url Uploaded`, `Case Comment Added`, `Case Comment Updated`, `Membership Created`, `Membership Updated`, `Membership Cancelled`, `Register Delegated Administrator`, `Deregister Delegated Administrator`, `Disable AWS Service Access`.
* `email` - (Required) Email address of the team member. Must be between 6 and 254 characters.
* `job_title` - (Required) Job title of the team member. Must be between 1 and 50 characters.
* `name` - (Required) Name of the team member. Must be between 3 and 50 characters.

### `opt_in_feature` Block

The `opt_in_feature` block supports:

* `feature_name` - (Required) Name of the feature. Allowed value is: `Triage`.
* `is_enabled` - (Required) Whether to enable the feature. `Triage` is enabled by default and cannot be disabled, so this must be `true` for `Triage`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `account_id` - ID of the AWS account that owns the membership.
* `arn` - ARN of the membership.
* `customer_type` - Type of customer.
* `id` - ID of the membership.
* `membership_activation_timestamp` - Date and time the membership was activated, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `membership_status` - Status of the membership.
* `number_of_accounts_covered` - Number of accounts covered by the membership.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_securityir_membership.example
  identity = {
    id = "m-abcd1234efgh"
  }
}

resource "aws_securityir_membership" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` - ID of the membership.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Incident Response Membership using the `id`. For example:

```terraform
import {
  to = aws_securityir_membership.example
  id = "m-abcd1234efgh"
}
```

Using `terraform import`, import Security Incident Response Membership using the `id`. For example:

```console
% terraform import aws_securityir_membership.example m-abcd1234efgh
```
