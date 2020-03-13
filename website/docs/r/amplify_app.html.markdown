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

```hcl
resource "aws_amplify_app" "app" {
  name       = "app"
  repository = "https://github.com/example/app"

  // The default build_spec added by the Amplify Console for React.
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

  // The default custom_rules added by the Amplify Console.
  custom_rules {
    source = "/<*>"
    status = "404"
    target = "/index.html"
  }
}
```

### Repository with Tokens

If you create a new Amplify App with the `repository` argument, you also need to set `oauth_token` or `access_token` for authentication. For GitHub, get a [personal access token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line) and set `access_token` as follows:

```hcl
resource "aws_amplify_app" "app" {
  name       = "app"
  repository = "https://github.com/example/app"

  // GitHub personal access token
  access_token = "..."
}
```

You can omit `access_token` if you import an existing Amplify App created by the Amplify Console (using OAuth for authentication).

### Auto Branch Creation

```hcl
resource "aws_amplify_app" "app" {
  name = "app"

  auto_branch_creation_config {
    // Enable auto branch creation.
    enable_auto_branch_creation = true

    // The default patterns added by the Amplify Console.
    auto_branch_creation_patterns = [
      "*",
      "*/**",
    ]

    // Enable auto build for the created branch.
    enable_auto_build = true
  }
}
```

### Basic Authentication

```hcl
resource "aws_amplify_app" "app" {
  name = "app"

  basic_auth_config {
    // Enable basic authentication.
    enable_basic_auth = true

    username = "username"
    password = "password"
  }
}
```

### Rewrites and redirects

```hcl
resource "aws_amplify_app" "app" {
  name = "app"

  // Reverse Proxy Rewrite for API requests
  // https://docs.aws.amazon.com/amplify/latest/userguide/redirects.html#reverse-proxy-rewrite
  custom_rules {
    source = "/api/<*>"
    status = "200"
    target = "https://api.example.com/api/<*>"
  }

  // Redirects for Single Page Web Apps (SPA)
  // https://docs.aws.amazon.com/amplify/latest/userguide/redirects.html#redirects-for-single-page-web-apps-spa
  custom_rules {
    source = "</^[^.]+$|\\.(?!(css|gif|ico|jpg|js|png|txt|svg|woff|ttf|map|json)$)([^.]+$)/>"
    status = "200"
    target = "/index.html"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name for the Amplify App.
* `access_token` - (Optional) Personal Access token for 3rd party source control system for an Amplify App, used to create webhook and read-only deploy key. Token is not stored.
* `auto_branch_creation_config` - (Optional) Automated branch creation config for the Amplify App. An `auto_branch_creation_config` block is documented below.
* `basic_auth_config` - (Optional) Basic Authentication config for the Amplify App. A `basic_auth_config` block is documented below.
* `build_spec` - (Optional) BuildSpec content for Amplify App.
* `custom_rules` - (Optional) Custom redirect / rewrite rules for the Amplify App. A `custom_rules` block is documented below.
* `description` - (Optional) Description for the Amplify App.
* `enable_branch_auto_build` - (Optional) Enables auto-building of branches for the Amplify App.
* `environment_variables` - (Optional) Environment Variables for the Amplify App.
* `iam_service_role_arn` - (Optional) IAM service role ARN for the Amplify App.
* `oauth_token` - (Optional) OAuth token for 3rd party source control system for an Amplify App, used to create webhook and read-only deploy key. OAuth token is not stored.
* `platform` - (Optional) Platform for the Amplify App.
* `repository` - (Optional) Repository for the Amplify App.
* `tags` - (Optional) Key-value mapping of resource tags.

An `auto_branch_creation_config` block supports the following arguments:

* `enable_auto_branch_creation` - (Optional) Enables automated branch creation for the Amplify App.
* `auto_branch_creation_patterns` - (Optional) Automated branch creation glob patterns for the Amplify App.
* `basic_auth_config` - (Optional) Basic Authentication config for the auto created branch. A `basic_auth_config` block is documented below.
* `build_spec` - (Optional) BuildSpec for the auto created branch.
* `enable_auto_build` - (Optional) Enables auto building for the auto created branch.
* `enable_basic_auth` - (Optional) Enables Basic Auth for the auto created branch.
* `enable_pull_request_preview` - (Optional) Enables Pull Request Preview for auto created branch.
* `environment_variables` - (Optional) Environment Variables for the auto created branch.
* `framework` - (Optional) Framework for the auto created branch.
* `pull_request_environment_name` - (Optional) The Amplify Environment name for the pull request.
* `stage` - (Optional) Stage for the branch. Possible values: "PRODUCTION", "BETA", "DEVELOPMENT", "EXPERIMENTAL", or "PULL_REQUEST".

An `basic_auth_config` block supports the following arguments:

* `enable_basic_auth` - (Optional) Enables Basic Authorization.
* `username` - (Optional) Basic Authorization username.
* `password` - (Optional) Basic Authorization password.

A `custom_rules` block supports the following arguments:

* `source` - (Required) The source pattern for a URL rewrite or redirect rule.
* `target` - (Required) The target pattern for a URL rewrite or redirect rule.
* `condition` - (Optional) The condition for a URL rewrite or redirect rule, e.g. country code.
* `status` - (Optional) The status code for a URL rewrite or redirect rule.

## Attribute Reference

The following attributes are exported:

* `arn` - ARN for the Amplify App.
* `default_domain` - Default domain for the Amplify App.

## Import

Amplify App can be imported using Amplify App ID (appId), e.g.

```
$ terraform import aws_amplify_app.app d2ypk4k47z8u6
```

App ID can be obtained from App ARN (e.g. `arn:aws:amplify:us-east-1:12345678:apps/d2ypk4k47z8u6`).
