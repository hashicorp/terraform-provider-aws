---
layout: "aws"
page_title: "AWS: aws_transfer_user"
sidebar_current: "docs-aws-resource-transfer-user"
description: |-
  Provides a AWS Transfer User resource.
---

# aws_transfer_server

Provides a AWS Transfer User resource.


```hcl
resource "aws_transfer_server" "foo" {
	identity_provider_type = "SERVICE_MANAGED"

	tags {
		NAME     = "tf-acc-test-transfer-server"
	}
}


resource "aws_iam_role" "foo" {
	name = "tf-test-transfer-user-iam-role-%s"

	assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Effect": "Allow",
		"Principal": {
			"Service": "transfer.amazonaws.com"
		},
		"Action": "sts:AssumeRole"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "foo" {
	name = "tf-test-transfer-user-iam-policy-%s"
	role = "${aws_iam_role.foo.id}"
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AllowFullAccesstoS3",
			"Effect": "Allow",
			"Action": [
				"s3:*"
			],
			"Resource": "*"
		}
	]
}
POLICY
}


resource "aws_transfer_user" "foo" {
	server_id      = "${aws_transfer_server.foo.id}"
	user_name      = "tftestuser"
	role           = "${aws_iam_role.foo.arn}"

	tags {
		NAME = "tftestuser"
	}
}

```

## Argument Reference

The following arguments are supported:

* `server_id` - (Requirement) The Server ID of the Transfer Server (e.g. `s-12345678`)
* `user_name` - (Requirement) The name used for log in to your SFTP server.
* `home_directory` - (Optional) The landing directory (folder) for a user when they log in to the server using their SFTP client.
* `policy` - (Optional) The policy scopes down user access to portions of their Amazon S3 bucket. Variables you can use inside this policy include ${Transfer:UserName}, ${Transfer:HomeDirectory}, and ${Transfer:HomeBucket}.
* `role` - (Optional) Amazon Resource Name (ARN) of an IAM role that allows the service to controls your user’s access to your Amazon S3 bucket.
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of Transfer User

## Import

Transfer user can be imported using the `user_name` and `server_id` separated by `/`.

```
$ terraform import aws_transfer_user.bar test-username/s-12345678
```