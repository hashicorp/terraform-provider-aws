---
subcategory: "DataSync"
layout: "aws"
page_title: "AWS: aws_datasync_location_object_storage"
description: |-
  Manages an AWS DataSync Object Storage Location
---

# Resource: aws_datasync_location_object_storage

Manages a Object Storage Location within AWS DataSync.

~> **NOTE:** The DataSync Agents must be available before creating this resource.

## Example Usage

```terraform
resource "aws_datasync_location_object_storage" "example" {
  agent_arns      = [aws_datasync_agent.example.arn]
  server_hostname = "example"
  bucket_name     = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `agent_arns` - (Required) A list of DataSync Agent ARNs with which this location will be associated.
* `access_key` - (Optional) The access key is used if credentials are required to access the self-managed object storage server. If your object storage requires a user name and password to authenticate, use `access_key` and `secret_key` to provide the user name and password, respectively.
* `bucket_name` - (Required) The bucket on the self-managed object storage server that is used to read data from.
* `secret_key` - (Optional) The secret key is used if credentials are required to access the self-managed object storage server. If your object storage requires a user name and password to authenticate, use `access_key` and `secret_key` to provide the user name and password, respectively.
* `server_certificate` - (Optional) Specifies a certificate to authenticate with an object storage system that uses a private or self-signed certificate authority (CA). You must specify a Base64-encoded .pem string. The certificate can be up to 32768 bytes (before Base64 encoding).
* `server_hostname` - (Required) The name of the self-managed object storage server. This value is the IP address or Domain Name Service (DNS) name of the object storage server. An agent uses this host name to mount the object storage server in a network.
* `server_protocol` - (Optional) The protocol that the object storage server uses to communicate. Valid values are `HTTP` or `HTTPS`.
* `server_port` - (Optional) The port that your self-managed object storage server accepts inbound network traffic on. The server port is set by default to TCP 80 (`HTTP`) or TCP 443 (`HTTPS`). You can specify a custom port if your self-managed object storage server requires one.
* `subdirectory` - (Optional) A subdirectory in the HDFS cluster. This subdirectory is used to read data from or write data to the HDFS cluster. If the subdirectory isn't specified, it will default to /.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Location. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the DataSync Location.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `uri` - The URL of the Object Storage location that was described.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_datasync_location_object_storage` using the Amazon Resource Name (ARN). For example:

```terraform
import {
  to = aws_datasync_location_object_storage.example
  id = "arn:aws:datasync:us-east-1:123456789012:location/loc-12345678901234567"
}
```

Using `terraform import`, import `aws_datasync_location_object_storage` using the Amazon Resource Name (ARN). For example:

```console
% terraform import aws_datasync_location_object_storage.example arn:aws:datasync:us-east-1:123456789012:location/loc-12345678901234567
```
