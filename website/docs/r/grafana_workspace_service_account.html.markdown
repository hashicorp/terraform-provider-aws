---
subcategory: "Managed Grafana"
layout: "aws"
page_title: "AWS: aws_grafana_workspace_service_account"
description: |-
  Terraform resource for managing an Amazon Managed Grafana Workspace Service Account.
---

# Resource: aws_grafana_workspace_service_account

-> **Note:** You cannot update a service account. If you change any attribute, Terraform
will delete the current and create a new one.

Read about Service Accounts in the [Amazon Managed Grafana user guide](https://docs.aws.amazon.com/grafana/latest/userguide/service-accounts.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_grafana_workspace_service_account" "example" {
  name         = "example-admin"
  grafana_role = "ADMIN"
  workspace_id = aws_grafana_workspace.example.id
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) A name for the service account. The name must be unique within the workspace, as it determines the ID associated with the service account.
* `grafana_role` - (Required) The permission level to use for this service account. For more information about the roles and the permissions each has, see the [User roles](https://docs.aws.amazon.com/grafana/latest/userguide/Grafana-user-roles.html) documentation.
* `workspace_id` - (Required) The Grafana workspace with which the service account is associated.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `service_account_id` - Identifier of the service account in the given Grafana workspace

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Managed Grafana Workspace Service Account using the `workspace_id` and `service_account_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_grafana_workspace_service_account.example
  id = "g-abc12345,1"
}
```

Using `terraform import`, import Managed Grafana Workspace Service Account using the `workspace_id` and `service_account_id` separated by a comma (`,`). For example:

```console
% terraform import aws_grafana_workspace_service_account.example g-abc12345,1
```
