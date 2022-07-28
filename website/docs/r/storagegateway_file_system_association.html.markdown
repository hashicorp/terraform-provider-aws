---
subcategory: "Storage Gateway"
layout: "aws"
page_title: "AWS: aws_storagegateway_file_system_association"
description: |-
  Mananges an association between an Amazon FSx file system and an Amazon FSx File Gateway.
---

# Resource: aws_storagegateway_file_system_association

Associate an Amazon FSx file system with the FSx File Gateway. After the association process is complete, the file shares on the Amazon FSx file system are available for access through the gateway. This operation only supports the FSx File Gateway type.

[FSx File Gateway requirements](https://docs.aws.amazon.com/filegateway/latest/filefsxw/Requirements.html).

## Example Usage

```terraform
resource "aws_storagegateway_file_system_association" "example" {
  gateway_arn           = aws_storagegateway_gateway.example.arn
  location_arn          = aws_fsx_windows_file_system.example.arn
  username              = "Admin"
  password              = "avoid-plaintext-passwords"
  audit_destination_arn = aws_s3_bucket.example.arn
}
```

## Required Services Example

```terraform
data "aws_ssm_parameter" "aws_service_storagegateway_ami_FILE_S3_latest" {
  name = "/aws/service/storagegateway/ami/FILE_S3/latest"
}

resource "aws_instance" "test" {
  # If using a single root module to build full gateway stack
  # you must include the dependencies below
  depends_on = [aws_route.test, aws_vpc_dhcp_options_association.test]

  ami                         = data.aws_ssm_parameter.aws_service_storagegateway_ami_FILE_S3_latest.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test[0].id
}

resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = "test-sgw"
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_FSX_SMB"

  smb_active_directory_settings {
    domain_name = aws_directory_service_directory.test.name
    password    = aws_directory_service_directory.test.password
    username    = "Admin"
  }
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  security_group_ids  = [aws_security_group.test.id]
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8
}

resource "aws_storagegateway_file_system_association" "fsx" {

  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_fsx_windows_file_system.test.arn
  username     = "Admin"
  password     = aws_directory_service_directory.test.password
  cache_attributes {
    cache_stale_timeout_in_seconds = 400
  }
  audit_destination_arn = aws_cloudwatch_log_group.test.arn
}
```

## Argument Reference

The following arguments are supported:

* `gateway_arn` - (Required) The Amazon Resource Name (ARN) of the gateway.
* `location_arn` - (Required) The Amazon Resource Name (ARN) of the Amazon FSx file system to associate with the FSx File Gateway.
* `username` - (Required) The user name of the user credential that has permission to access the root share of the Amazon FSx file system. The user account must belong to the Amazon FSx delegated admin user group.
* `password` - (Required, sensitive) The password of the user credential.
* `audit_destination_arn` - (Optional) The Amazon Resource Name (ARN) of the storage used for the audit logs.
* `cache_attributes` - (Optional) Refresh cache information. see [Cache Attributes](#cache_attributes) for more details.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### cache_attributes

* `cache_stale_timeout_in_seconds` - (Optional) Refreshes a file share's cache by using Time To Live (TTL).
 TTL is the length of time since the last refresh after which access to the directory would cause the file gateway
  to first refresh that directory's contents from the Amazon S3 bucket. Valid Values: `0` or `300` to `2592000` seconds (5 minutes to 30 days). Defaults to `0`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the FSx file system association
* `arn` - Amazon Resource Name (ARN) of the newly created file system association.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

`aws_storagegateway_file_system_association` can be imported by using the FSx file system association Amazon Resource Name (ARN), e.g.,

```
$ terraform import aws_storagegateway_file_system_association.example arn:aws:storagegateway:us-east-1:123456789012:fs-association/fsa-0DA347732FDB40125
```
