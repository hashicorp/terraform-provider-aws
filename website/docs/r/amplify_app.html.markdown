---
subcategory: "Amplify"
layout: "aws"
page_title: "AWS: aws_amplify_app"
description: |-
  Provides an Amplify App resource.
---

# Resource: aws_amplify_app

Provides an Amplify App resource, a fullstack serverless app hosted on the [AWS Amplify Console](https://docs.aws.amazon.com/amplify/latest/userguide/welcome.html).

~> **Note:** When you create/update an Amplify App from Terraform, you may end up with the error "BadRequestException: You should at least provide one valid token" because of authentication issues. See the section "Repository with Tokens" below.

## Example Usage

```terraform
resource "aws_amplify_app" "example" {
  name       = "example"
  repository = "https://github.com/example/app"

  # The default build_spec added by the Amplify Console for React.
  build_spec = <<-EOT
    version: 0.1
    frontend:
      phases:
        preBuild:
          commands:
            - yarn install
        build:
          commands:
            - yarn run build
      artifacts:
        baseDirectory: build
        files:
          - '**/*'
      cache:
        paths:
          - node_modules/**/*
  EOT

  # The default rewrites and redirects added by the Amplify Console.
  custom_rule {
    source = "/<*>"
    status = "404"
    target = "/index.html"
  }

  environment_variables = {
    ENV = "test"
  }
}
```

### Repository with Tokens

If you create a new Amplify App with the `repository` argument, you also need to set `oauth_token` or `access_token` for authentication. For GitHub, get a [personal access token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line) and set `access_token` as follows:

```terraform
resource "aws_amplify_app" "example" {
  name       = "example"
  repository = "https://github.com/example/app"

  # GitHub personal access token
  access_token = "..."
}
```

You can omit `access_token` if you import an existing Amplify App created by the Amplify Console (using OAuth for authentication).

### Auto Branch Creation

```terraform
resource "aws_amplify_app" "example" {
  name = "example"

  enable_auto_branch_creation = true

  # The default patterns added by the Amplify Console.
  auto_branch_creation_patterns = [
    "*",
    "*/**",
  ]

  auto_branch_creation_config {
    # Enable auto build for the created branch.
    enable_auto_build = true
  }
}
```

### Basic Authorization

```terraform
resource "aws_amplify_app" "example" {
  name = "example"

  enable_basic_auth      = true
  basic_auth_credentials = base64encode("username1:password1")
}
```

### Rewrites and Redirects

```terraform
resource "aws_amplify_app" "example" {
  name = "example"

  # Reverse Proxy Rewrite for API requests
  # https://docs.aws.amazon.com/amplify/latest/userguide/redirects.html#reverse-proxy-rewrite
  custom_rule {
    source = "/api/<*>"
    status = "200"
    target = "https://api.example.com/api/<*>"
  }

  # Redirects for Single Page Web Apps (SPA)
  # https://docs.aws.amazon.com/amplify/latest/userguide/redirects.html#redirects-for-single-page-web-apps-spa
  custom_rule {
    source = "</^[^.]+$|\\.(?!(css|gif|ico|jpg|js|png|txt|svg|woff|ttf|map|json)$)([^.]+$)/>"
    status = "200"
    target = "/index.html"
  }
}
```

### Custom Image

```terraform
resource "aws_amplify_app" "example" {
  name = "example"

  environment_variables = {
    "_CUSTOM_IMAGE" = "node:16",
  }
}
```

### Custom Headers

```terraform
resource "aws_amplify_app" "example" {
  name = "example"

  custom_headers = <<-EOT
    customHeaders:
      - pattern: '**'
        headers:
          - key: 'Strict-Transport-Security'
            value: 'max-age=31536000; includeSubDomains'
          - key: 'X-Frame-Options'
            value: 'SAMEORIGIN'
          - key: 'X-XSS-Protection'
            value: '1; mode=block'
          - key: 'X-Content-Type-Options'
            value: 'nosniff'
          - key: 'Content-Security-Policy'
            value: "default-src 'self'"
  EOT
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Name for an Amplify app.
* `access_token` - (Optional) Personal access token for a third-party source control system for an Amplify app. This token must have write access to the relevant repo to create a webhook and a read-only deploy key for the Amplify project. The token is not stored, so after applying this attribute can be removed and the setup token deleted.
* `auto_branch_creation_config` - (Optional) Automated branch creation configuration for an Amplify app. An `auto_branch_creation_config` block is documented below.
* `auto_branch_creation_patterns` - (Optional) Automated branch creation glob patterns for an Amplify app.
* `basic_auth_credentials` - (Optional) Credentials for basic authorization for an Amplify app.
* `build_spec` - (Optional) The [build specification](https://docs.aws.amazon.com/amplify/latest/userguide/build-settings.html) (build spec) for an Amplify app.
* `custom_headers` - (Optional) The [custom HTTP headers](https://docs.aws.amazon.com/amplify/latest/userguide/custom-headers.html) for an Amplify app.
* `custom_rule` - (Optional) Custom rewrite and redirect rules for an Amplify app. A `custom_rule` block is documented below.
* `description` - (Optional) Description for an Amplify app.
* `enable_auto_branch_creation` - (Optional) Enables automated branch creation for an Amplify app.
* `enable_basic_auth` - (Optional) Enables basic authorization for an Amplify app. This will apply to all branches that are part of this app.
* `enable_branch_auto_build` - (Optional) Enables auto-building of branches for the Amplify App.
* `enable_branch_auto_deletion` - (Optional) Automatically disconnects a branch in the Amplify Console when you delete a branch from your Git repository.
* `environment_variables` - (Optional) Environment variables map for an Amplify app.
* `iam_service_role_arn` - (Optional) AWS Identity and Access Management (IAM) service role for an Amplify app.
* `oauth_token` - (Optional) OAuth token for a third-party source control system for an Amplify app. The OAuth token is used to create a webhook and a read-only deploy key. The OAuth token is not stored.
* `platform` - (Optional) Platform or framework for an Amplify app. Valid values: `WEB`, `WEB_COMPUTE`. Default value: `WEB`.
* `repository` - (Optional) Repository for an Amplify app.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

An `auto_branch_creation_config` block supports the following arguments:

* `basic_auth_credentials` - (Optional) Basic authorization credentials for the autocreated branch.
* `build_spec` - (Optional) Build specification (build spec) for the autocreated branch.
* `enable_auto_build` - (Optional) Enables auto building for the autocreated branch.
* `enable_basic_auth` - (Optional) Enables basic authorization for the autocreated branch.
* `enable_performance_mode` - (Optional) Enables performance mode for the branch.
* `enable_pull_request_preview` - (Optional) Enables pull request previews for the autocreated branch.
* `environment_variables` - (Optional) Environment variables for the autocreated branch.
* `framework` - (Optional) Framework for the autocreated branch.
* `pull_request_environment_name` - (Optional) Amplify environment name for the pull request.
* `stage` - (Optional) Describes the current stage for the autocreated branch. Valid values: `PRODUCTION`, `BETA`, `DEVELOPMENT`, `EXPERIMENTAL`, `PULL_REQUEST`.

A `custom_rule` block supports the following arguments:

* `condition` - (Optional) Condition for a URL rewrite or redirect rule, such as a country code.
* `source` - (Required) Source pattern for a URL rewrite or redirect rule.
* `status` - (Optional) Status code for a URL rewrite or redirect rule. Valid values: `200`, `301`, `302`, `404`, `404-200`.
* `target` - (Required) Target pattern for a URL rewrite or redirect rule.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Amplify app.
* `default_domain` - Default domain for the Amplify app.
* `id` - Unique ID of the Amplify app.
* `production_branch` - Describes the information about a production branch for an Amplify app. A `production_branch` block is documented below.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

A `production_branch` block supports the following attributes:

* `branch_name` - Branch name for the production branch.
* `last_deploy_time` - Last deploy time of the production branch.
* `status` - Status of the production branch.
* `thumbnail_url` - Thumbnail URL for the production branch.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amplify App using Amplify App ID (appId). For example:

```terraform
import {
  to = aws_amplify_app.example
  id = "d2ypk4k47z8u6"
}
```

Using `terraform import`, import Amplify App using Amplify App ID (appId). For example:

```console
% terraform import aws_amplify_app.example d2ypk4k47z8u6
```

App ID can be obtained from App ARN (e.g., `arn:aws:amplify:us-east-1:12345678:apps/d2ypk4k47z8u6`).
