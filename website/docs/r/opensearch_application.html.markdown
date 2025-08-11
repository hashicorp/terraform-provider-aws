---
subcategory: "OpenSearch"
layout: "aws"
page_title: "AWS: aws_opensearch_application"
description: |-
  Provides an AWS OpenSearch Application resource.
---

# Resource: aws_opensearch_application

Provides an AWS OpenSearch Application resource. OpenSearch Applications provide a user interface for interacting with OpenSearch data and managing OpenSearch resources.

## Example Usage

### Basic Usage

```terraform
resource "aws_opensearch_application" "example" {
  name = "my-opensearch-app"
}
```

### Application with Configuration

```terraform
resource "aws_opensearch_application" "example" {
  name = "my-opensearch-app"

  app_config {
    key   = "opensearchDashboards.dashboardAdmin.users"
    value = "admin-user"
  }

  app_config {
    key   = "opensearchDashboards.dashboardAdmin.groups"
    value = "admin-group"
  }

  tags = {
    Environment = "production"
    Team        = "data-platform"
  }
}
```

### Application with Data Sources

```terraform
resource "aws_opensearch_domain" "example" {
  domain_name    = "example-domain"
  engine_version = "OpenSearch_2.3"

  cluster_config {
    instance_type = "t3.small.search"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 20
  }
}

resource "aws_opensearch_application" "example" {
  name = "my-opensearch-app"

  data_source {
    data_source_arn         = aws_opensearch_domain.example.arn
    data_source_description = "Primary OpenSearch domain for analytics"
  }

  tags = {
    Environment = "production"
  }
}
```

### Application with IAM Identity Center Integration

```terraform
# Data sources for account and region information
data "aws_ssoadmin_instances" "example" {}
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# IAM Policy for OpenSearch Application Identity Center Integration
resource "aws_iam_policy" "opensearch_identity_center" {
  name        = "opensearch-identity-center-policy"
  description = "Policy for OpenSearch Application Identity Center integration"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "IdentityStoreOpenSearchDomainConnectivity"
        Effect = "Allow"
        Action = [
          "identitystore:DescribeUser",
          "identitystore:ListGroupMembershipsForMember",
          "identitystore:DescribeGroup"
        ]
        Resource = "*"
        Condition = {
          "ForAnyValue:StringEquals" = {
            "aws:CalledViaLast" = "es.amazonaws.com"
          }
        }
      },
      {
        Sid    = "OpenSearchDomain"
        Effect = "Allow"
        Action = [
          "es:ESHttp*"
        ]
        Resource = "*"
      },
      {
        Sid    = "OpenSearchServerless"
        Effect = "Allow"
        Action = [
          "aoss:APIAccessAll"
        ]
        Resource = "*"
      }
    ]
  })
}

# IAM Role for OpenSearch Application
resource "aws_iam_role" "opensearch_application" {
  name = "opensearch-application-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "application.opensearchservice.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      },
      {
        Effect = "Allow"
        Principal = {
          Service = "application.opensearchservice.amazonaws.com"
        }
        Action = "sts:SetContext"
        Condition = {
          "ForAllValues:ArnEquals" = {
            "sts:RequestContextProviders" = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/portal.sso.${data.aws_region.current.id}.amazonaws.com/apl/*"
          }
        }
      }
    ]
  })
}

# Attach policy to role
resource "aws_iam_role_policy_attachment" "opensearch_identity_center" {
  role       = aws_iam_role.opensearch_application.name
  policy_arn = aws_iam_policy.opensearch_identity_center.arn
}

resource "aws_opensearch_application" "example" {
  name = "my-opensearch-app"

  iam_identity_center_options {
    enabled                                         = true
    iam_identity_center_instance_arn                = tolist(data.aws_ssoadmin_instances.example.arns)[0]
    iam_role_for_identity_center_application_arn    = aws_iam_role.opensearch_application.arn
  }

  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The unique name of the OpenSearch application. Names must be unique within an AWS Region for each account. Must be between 3 and 30 characters, start with a lowercase letter, and contain only lowercase letters, numbers, and hyphens.
* `app_config` - (Optional) Configuration block(s) for OpenSearch application settings. See [App Config](#app-config) below.
* `client_token` - (Optional) Unique, case-sensitive identifier to ensure idempotency of the request. Must be between 1 and 64 characters.
* `data_source` - (Optional) Configuration block(s) for data sources to link to the OpenSearch application. See [Data Source](#data-source) below.
* `iam_identity_center_options` - (Optional) Configuration block for integrating AWS IAM Identity Center with the OpenSearch application. See [IAM Identity Center Options](#iam-identity-center-options) below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### App Config

The `app_config` block supports the following arguments:

* `key` - (Optional) The configuration item to set. Valid values are `opensearchDashboards.dashboardAdmin.users` and `opensearchDashboards.dashboardAdmin.groups`.
* `value` - (Optional) The value assigned to the configuration key, such as an IAM user ARN or group name. Must be between 1 and 4096 characters.

### Data Source

The `data_source` block supports the following arguments:

* `data_source_arn` - (Optional) The Amazon Resource Name (ARN) of the OpenSearch domain or collection. Must be between 20 and 2048 characters.
* `data_source_description` - (Optional) A detailed description of the data source. Must be at most 1000 characters and contain only alphanumeric characters, underscores, spaces, and the following special characters: `@#%*+=:?./!-`.

### IAM Identity Center Options

The `iam_identity_center_options` block supports the following arguments:

* `enabled` - (Optional) Specifies whether IAM Identity Center is enabled or disabled.
* `iam_identity_center_instance_arn` - (Optional) The Amazon Resource Name (ARN) of the IAM Identity Center instance. Must be between 20 and 2048 characters.
* `iam_role_for_identity_center_application_arn` - (Optional) The ARN of the IAM role associated with the IAM Identity Center application. Must be between 20 and 2048 characters and match the pattern for IAM role ARNs.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the OpenSearch application.
* `id` - The unique identifier of the OpenSearch application.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

OpenSearch Applications can be imported using the application `id`, e.g.,

```
$ terraform import aws_opensearch_application.example app-1234567890abcdef0
```

## Additional Information

For more information about OpenSearch Applications, see the [AWS OpenSearch Service Developer Guide](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/application.html).

For information about configuring IAM Identity Center with OpenSearch Applications, see [Using AWS IAM Identity Center authentication](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/application-getting-started.html#create-application).
