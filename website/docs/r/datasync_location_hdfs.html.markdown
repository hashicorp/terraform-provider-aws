---
subcategory: "DataSync"
layout: "aws"
page_title: "AWS: aws_datasync_location_hdfs"
description: |-
  Manages an AWS DataSync Hdfs Location
---

# Resource: aws_datasync_location_hdfs

Manages a Hdfs Location within AWS DataSync.

~> **NOTE:** The DataSync Agents must be available before creating this resource.

## Example Usage

```terraform
resource "aws_datasync_location_hdfs" "example" {
  agent_arns          = [aws_datasync_agent.example.arn]
  authentication_type = "SIMPLE"
  simple_user         = "example"

  name_node {
    hostname = aws_instance.example.private_dns
	  port     = 80
  }
}
```

## Argument Reference

The following arguments are supported:

* `agent_arns` - (Required) A list of DataSync Agent ARNs with which this location will be associated.
* `authentication_type` - (Required) The type of authentication used to determine the identity of the user. Valid values are `SIMPLE` and `KERBEROS`.
* `name_node` - (Required)  The NameNode that manages the HDFS namespace. The NameNode performs operations such as opening, closing, and renaming files and directories. The NameNode contains the information to map blocks of data to the DataNodes. You can use only one NameNode.
* `simple_user` - (Optional) The user name used to identify the client on the host operating system. If `SIMPLE` is specified for `authentication_type`, this parameter is required.
* `subdirectory` - (Optional) A subdirectory in the HDFS cluster. This subdirectory is used to read data from or write data to the HDFS cluster. If the subdirectory isn't specified, it will default to /.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Location. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### name_node Argument Reference

The following arguments are supported inside the `mount_options` configuration block:

* `hostname` - (Required) The hostname of the NameNode in the HDFS cluster. This value is the IP address or Domain Name Service (DNS) name of the NameNode. An agent that's installed on-premises uses this hostname to communicate with the NameNode in the network.
* `port` - (Required) The port that the NameNode uses to listen to client requests.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the DataSync Location.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

`aws_datasync_location_hdfs` can be imported by using the Amazon Resource Name (ARN), e.g.,

```
$ terraform import aws_datasync_location_hdfs.example arn:aws:datasync:us-east-1:123456789012:location/loc-12345678901234567
```
