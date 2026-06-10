---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_connector_v2"
description: |-
  Manages a Security Hub V2 connector.
---

# Resource: aws_securityhub_connector_v2

Manages a Security Hub V2 connector.

~> **NOTE:** Connectors must be created in the aggregation (home) region. A Security Hub V2 Aggregator (`aws_securityhub_aggregator_v2`) must exist before creating connectors.

~> **NOTE:** After creation, the connector will be in `PENDING_AUTHORIZATION` status. Use the `auth_url` output to complete the OAuth authorization flow.

## Example Usage

### Jira Cloud

```terraform
resource "aws_securityhub_account_v2" "example" {}

resource "aws_securityhub_aggregator_v2" "example" {
  region_linking_mode = "ALL_REGIONS"

  depends_on = [aws_securityhub_account_v2.example]
}

resource "aws_securityhub_connector_v2" "example" {
  name = "jira-connector"

  connector_provider {
    jira_cloud {
      project_key = "SEC"
    }
  }

  depends_on = [aws_securityhub_aggregator_v2.example]
}

output "auth_url" {
  value = aws_securityhub_connector_v2.example.auth_url
}
```

### With Description and KMS Key

```terraform
resource "aws_securityhub_connector_v2" "example" {
  name        = "jira-connector"
  description = "Jira Cloud integration for security findings"
  kms_key_arn = aws_kms_key.example.arn

  connector_provider {
    jira_cloud {
      project_key = "SEC"
    }
  }

  depends_on = [aws_securityhub_aggregator_v2.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required, Forces new resource) The name of the connector.
* `description` - (Optional) A description of the connector.
* `connector_provider` - (Required, Forces new resource) Third-party provider details. See [`connector_provider`](#connector_provider) below.
* `kms_key_arn` - (Optional, Forces new resource) ARN of KMS key for connector encryption.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `connector_provider`

The `connector_provider` block supports the following:

* `jira_cloud` - (Optional) Details about a Jira Cloud integration. See [`jira_cloud`](#jira_cloud) below.
* `service_now` - (Optional) Details about a ServiceNow ITSM integration. See [`service_now`](#service_now) below.

### `jira_cloud`

The `jira_cloud` block supports the following:

* `auth_status` - (Computed) Status of the authorization between Jira Cloud and the service.
* `auth_url` - (Computed) URL to provide to customers for OAuth auth code flow.
* `cloud_id` - (Computed) Cloud ID of the Jira Cloud.
* `domain` - (Computed) URL domain of the Jira Cloud instance.
* `project_key` - (Required) Jira Cloud project key.

### `service_now`

The `service_now` block supports the following:

* `auth_status` - (Computed) Status of the authorization between ServiceNow and the service.
* `instance_name` - (Required) Instance name of ServiceNow ITSM.
* `secret_arn` - (Required) Amazon Resource Name (ARN) of the AWS Secrets Manager secret that contains the ServiceNow credentials.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the connector.
* `connector_id` - ID of the connector.
* `health` - Current health status. See [`health`](#health) below.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `health`

The `health` object supports the following:

* `connector_status` - Status of the connector.
* `last_checked_at` - Timestamp for the time the health status was checked.
* `message` - Message for the reason of `connector_status` change.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_securityhub_connector_v2.example
  identity = {
    connector_id = "8ecf045f-5a95-c24d-6769-5d52f6929563"
  }
}

resource "aws_securityhub_connector_v2" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `connector_id` (String) ID of the Security Hub V2 connector.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Hub V2 connectors using `connector_id`. For example:

```terraform
import {
  to = aws_securityhub_connector_v2.example
  id = "8ecf045f-5a95-c24d-6769-5d52f6929563"
}
```

Using `terraform import`, import Security Hub V2 connectors using `connector_id`. For example:

```console
% terraform import aws_securityhub_connector_v2.example 8ecf045f-5a95-c24d-6769-5d52f6929563
```
