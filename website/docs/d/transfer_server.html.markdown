---
subcategory: "Transfer Family"
layout: "aws"
page_title: "AWS: aws_transfer_server"
description: |-
  Get information on an AWS Transfer Server resource
---

# Data Source: aws_transfer_server

Use this data source to get the ARN of an AWS Transfer Server for use in other
resources.

## Example Usage

```terraform
data "aws_transfer_server" "example" {
  server_id = "s-1234567"
}
```

## Argument Reference

* `server_id` - (Required) ID for an SFTP server.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of Transfer Server.
* `certificate` - ARN of any certificate.
* `domain` -  The domain of the storage system that is used for file transfers.
* `endpoint` - Endpoint of the Transfer Server (e.g., `s-12345678.server.transfer.REGION.amazonaws.com`).
* `endpoint_type` - Type of endpoint that the server is connected to.
* `identity_provider_type` - The mode of authentication enabled for this service. The default value is `SERVICE_MANAGED`, which allows you to store and access SFTP user credentials within the service. `API_GATEWAY` indicates that user authentication requires a call to an API Gateway endpoint URL provided by you to integrate an identity provider of your choice.
* `invocation_role` - ARN of the IAM role used to authenticate the user account with an `identity_provider_type` of `API_GATEWAY`.
* `logging_role` - ARN of an IAM role that allows the service to write your SFTP users’ activity to your Amazon CloudWatch logs for monitoring and auditing purposes.
* `protocols` - File transfer protocol or protocols over which your file transfer protocol client can connect to your server's endpoint.
* `security_policy_name` - The name of the security policy that is attached to the server.
* `structured_log_destinations` - A set of ARNs of destinations that will receive structured logs from the transfer server such as CloudWatch Log Group ARNs.
* `tags` - Map of tags assigned to the resource.
* `url` - URL of the service endpoint used to authenticate users with an `identity_provider_type` of `API_GATEWAY`.
