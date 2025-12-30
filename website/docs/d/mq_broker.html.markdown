---
subcategory: "MQ"
layout: "aws"
page_title: "AWS: aws_mq_broker"
description: |-
  Provides details about an existing Amazon MQ broker.
---

# Data Source: aws_mq_broker

Provides details about an existing Amazon MQ broker. Use this data source to retrieve configuration and metadata for an Amazon MQ broker by ID or name.

## Example Usage

```terraform
# Get broker by ID
data "aws_mq_broker" "example" {
  broker_id = "b-1234a5b6-78cd-901e-2fgh-3i45j6k178l9"
}

# Get broker by name
data "aws_mq_broker" "example" {
  broker_name = "example"
}
```

## Argument Reference

The following arguments are optional:

* `broker_id` - (Optional) Unique ID of the MQ broker.
* `broker_name` - (Optional) Unique name of the MQ broker.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
~> **Note:** Either `broker_id` or `broker_name` must be specified.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the broker.
* `authentication_strategy` - Authentication strategy used to secure the broker.
* `auto_minor_version_upgrade` - Whether to automatically upgrade to new minor versions of brokers as Amazon MQ makes releases available.
* `configuration` - Configuration block for broker configuration. See [Configuration](#configuration) below.
* `deployment_mode` - Deployment mode of the broker.
* `encryption_options` - Configuration block containing encryption options. See [Encryption Options](#encryption-options) below.
* `engine_type` - Type of broker engine.
* `engine_version` - Version of the broker engine.
* `host_instance_type` - Broker's instance type.
* `instances` - List of information about allocated brokers (both active & standby). See [Instances](#instances) below.
* `ldap_server_metadata` - Configuration block for the LDAP server used to authenticate and authorize connections to the broker. See [LDAP Server Metadata](#ldap-server-metadata) below.
* `logs` - Configuration block for the logging configuration of the broker. See [Logs](#logs) below.
* `maintenance_window_start_time` - Configuration block for the maintenance window start time. See [Maintenance Window Start Time](#maintenance-window-start-time) below.
* `publicly_accessible` - Whether to enable connections from applications outside of the VPC that hosts the broker's subnets.
* `security_groups` - List of security group IDs assigned to the broker.
* `storage_type` - Storage type of the broker.
* `subnet_ids` - List of subnet IDs in which to launch the broker.
* `tags` - Map of tags assigned to the broker.
* `user` - Configuration block for broker users. See [User](#user) below.

### Configuration

* `id` - Configuration ID.
* `revision` - Revision of the Configuration.

### Encryption Options

* `kms_key_id` - Amazon Resource Name (ARN) of Key Management Service (KMS) Customer Master Key (CMK) to use for encryption at rest.
* `use_aws_owned_key` - Whether to enable an AWS-owned KMS CMK that is not in your account.

### Instances

* `console_url` - URL of the ActiveMQ Web Console or the RabbitMQ Management UI depending on `engine_type`.
* `endpoints` - Broker's wire-level protocol endpoints.
* `ip_address` - IP Address of the broker.

### LDAP Server Metadata

* `hosts` - List of a fully qualified domain name of the LDAP server and an optional failover server.
* `role_base` - Fully qualified name of the directory to search for a user's groups.
* `role_name` - LDAP attribute that identifies the group name attribute in the object returned from the group membership query.
* `role_search_matching` - Search criteria for groups.
* `role_search_subtree` - Whether the directory search scope is the entire sub-tree.
* `service_account_password` - Service account password.
* `service_account_username` - Service account username.
* `user_base` - Fully qualified name of the directory where you want to search for users.
* `user_role_name` - Name of the LDAP attribute for the user group membership.
* `user_search_matching` - Search criteria for users.
* `user_search_subtree` - Whether the directory search scope is the entire sub-tree.

### Logs

* `audit` - Whether audit logging is enabled.
* `general` - Whether general logging is enabled.

### Maintenance Window Start Time

* `day_of_week` - Day of the week.
* `time_of_day` - Time, in 24-hour format.
* `time_zone` - Time zone in either the Country/City format or the UTC offset format.

### User

* `console_access` - Whether to enable access to the ActiveMQ Web Console for the user.
* `groups` - List of groups to which the ActiveMQ user belongs.
* `replication_user` - Whether to set replication user.
* `username` - Username of the user.
