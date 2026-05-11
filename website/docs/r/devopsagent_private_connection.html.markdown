---
subcategory: "DevOps Agent"
layout: "aws"
page_title: "AWS: aws_devopsagent_private_connection"
description: |-
  Manages an AWS DevOps Agent Private Connection.
---

# Resource: aws_devopsagent_private_connection

Manages an AWS DevOps Agent Private Connection.

A Private Connection enables AWS DevOps Agent to securely connect to resources in your VPC or on-premises environment.

## Example Usage

### Self Managed

```terraform
resource "aws_devopsagent_private_connection" "example" {
  name                      = "example-connection"
  mode                      = "SELF_MANAGED"
  resource_configuration_id = aws_vpc_lattice_resource_configuration.example.id
  certificate               = var.certificate_pem
}
```

### Service Managed

```terraform
resource "aws_devopsagent_private_connection" "example" {
  name         = "example-connection"
  mode         = "SERVICE_MANAGED"
  host_address = "10.0.0.1"
  vpc_id       = aws_vpc.example.id
  subnet_ids   = aws_subnet.example[*].id
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required, Forces new resource) The name of the Private Connection. Must be unique within the account.
* `mode` - (Required, Forces new resource) The mode of the Private Connection. Valid values: `SELF_MANAGED`, `SERVICE_MANAGED`.

The following arguments are applicable for `SELF_MANAGED` type:

* `resource_configuration_id` - (Optional, Forces new resource) The ID or ARN of the VPC Lattice resource configuration.

The following arguments are applicable for `SERVICE_MANAGED` type:

* `host_address` - (Optional, Forces new resource) IP address or DNS name of the target resource.
* `subnet_ids` - (Optional, Forces new resource) Subnets that the service-managed Resource Gateway will span.
* `vpc_id` - (Optional, Forces new resource) VPC to create the service-managed Resource Gateway in.

The following arguments are optional for both types:

* `certificate` - (Optional, Sensitive) The certificate to associate with the Private Connection. This is the only field that can be updated in-place.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the Private Connection.
* `status` - The current status of the Private Connection. Values include `ACTIVE`, `CREATE_IN_PROGRESS`, `CREATE_FAILED`, `DELETE_IN_PROGRESS`, `DELETE_FAILED`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

DevOps Agent Private Connection can be imported using the `name`, e.g.,

```
$ terraform import aws_devopsagent_private_connection.example example-connection
```
