---
layout: "aws"
page_title: "AWS: aws_datapipeline"
sidebar_current: "docs-aws-resource-datapipeline"
description: |-
  Provides a AWS DataPipeline.
---

# aws_datapipeline

Provides a Data Pipeline resource.

## Example Usage

### Create Pipeline Only

```hcl
resource "aws_iam_role" "role" {
	name = "tf-test-datapipeline-role"
	  
	assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Principal": {
				"Service": [
					"elasticmapreduce.amazonaws.com",
					"datapipeline.amazonaws.com"
				]
			},
			"Action": "sts:AssumeRole"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "role" {
	name = "tf-test-transfer-user-iam-policy"
	role = "${aws_iam_role.role.id}"
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AllowFullAccesstoS3",
			"Effect": "Allow",
			"Action": [
				"cloudwatch:*",
				"datapipeline:DescribeObjects",
				"datapipeline:EvaluateExpression",
				"dynamodb:BatchGetItem",
				"dynamodb:DescribeTable",
				"dynamodb:GetItem",
				"dynamodb:Query",
				"dynamodb:Scan",
				"dynamodb:UpdateTable",
				"ec2:AuthorizeSecurityGroupIngress",
				"ec2:CancelSpotInstanceRequests",
				"ec2:CreateSecurityGroup",
				"ec2:CreateTags",
				"ec2:DeleteTags",
				"ec2:Describe*",
				"ec2:ModifyImageAttribute",
				"ec2:ModifyInstanceAttribute",
				"ec2:RequestSpotInstances",
				"ec2:RunInstances",
				"ec2:StartInstances",
				"ec2:StopInstances",
				"ec2:TerminateInstances",
				"ec2:AuthorizeSecurityGroupEgress", 
				"ec2:DeleteSecurityGroup", 
				"ec2:RevokeSecurityGroupEgress", 
				"ec2:DescribeNetworkInterfaces", 
				"ec2:CreateNetworkInterface", 
				"ec2:DeleteNetworkInterface", 
				"ec2:DetachNetworkInterface",
				"elasticmapreduce:*",
				"iam:GetInstanceProfile",
				"iam:GetRole",
				"iam:GetRolePolicy",
				"iam:ListAttachedRolePolicies",
				"iam:ListRolePolicies",
				"iam:ListInstanceProfiles",
				"iam:PassRole",
				"rds:DescribeDBInstances",
				"rds:DescribeDBSecurityGroups",
				"redshift:DescribeClusters",
				"redshift:DescribeClusterSecurityGroups",
				"s3:CreateBucket",
				"s3:DeleteObject",
				"s3:Get*",
				"s3:List*",
				"s3:Put*",
				"sdb:BatchPutAttributes",
				"sdb:Select*",
				"sns:GetTopicAttributes",
				"sns:ListTopics",
				"sns:Publish",
				"sns:Subscribe",
				"sns:Unsubscribe",
				"sqs:CreateQueue", 
				"sqs:Delete*", 
				"sqs:GetQueue*", 
				"sqs:PurgeQueue", 
				"sqs:ReceiveMessage" 
			],
			"Resource": "*"
		},
		{
			"Effect": "Allow",
			"Action": "iam:CreateServiceLinkedRole",
			"Resource": "*",
			"Condition": {
			  "StringLike": {
				  "iam:AWSServiceName": ["elasticmapreduce.amazonaws.com","spot.amazonaws.com"]
			  }
			}
		}
	]
}
POLICY
}



resource "aws_iam_role" "resource_role" {
	name = "tf-test-datapipeline-resource-role"
	  
	assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Principal": {
				"Service": [
					"ec2.amazonaws.com"
				]
			},
			"Action": "sts:AssumeRole"
		}
	]
}
EOF
}

resource "aws_iam_instance_profile" "resource_role" {
	name = "tf-test-datapipeline-resource-role-profile"
	role = "${aws_iam_role.resource_role.name}"
}

resource "aws_iam_role_policy" "resource_role" {
	name = "tf-test-transfer-user-iam-policy"
	role = "${aws_iam_role.resource_role.id}"
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AllowFullAccesstoS3",
			"Effect": "Allow",
			"Action": [
				"cloudwatch:*",
				"datapipeline:*",
				"dynamodb:*",
				"ec2:Describe*",
				"elasticmapreduce:AddJobFlowSteps",
				"elasticmapreduce:Describe*",
				"elasticmapreduce:ListInstance*",
				"rds:Describe*",
				"redshift:DescribeClusters",
				"redshift:DescribeClusterSecurityGroups",
				"s3:*",
				"sdb:*",
				"sns:*",
				"sqs:*"
			],
			"Resource": "*"
		}
	]
}
POLICY
}


resource "aws_datapipeline" "foo" {
	name      = "tf-datapipeline-%s"

	default {
		schedule_type          = "ondemand"
		failure_and_rerun_mode = "cascade"
		role                   = "${aws_iam_role.role.arn}"
		resource_role          = "${aws_iam_instance_profile.resource_role.arn}"
	}
}
```

### Create Copy Activity Pipeline

```hcl
resource "aws_iam_role" "role" {
	name = "tf-test-datapipeline-role"
	  
	assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Principal": {
				"Service":[
					"elasticmapreduce.amazonaws.com",
					"datapipeline.amazonaws.com"
				]
			},
			"Action": "sts:AssumeRole"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "role" {
	name = "tf-test-transfer-user-iam-policy"
	role = "${aws_iam_role.role.id}"
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AllowFullAccesstoS3",
			"Effect": "Allow",
			"Action": [
				"cloudwatch:*",
				"datapipeline:DescribeObjects",
				"datapipeline:EvaluateExpression",
				"dynamodb:BatchGetItem",
				"dynamodb:DescribeTable",
				"dynamodb:GetItem",
				"dynamodb:Query",
				"dynamodb:Scan",
				"dynamodb:UpdateTable",
				"ec2:AuthorizeSecurityGroupIngress",
				"ec2:CancelSpotInstanceRequests",
				"ec2:CreateSecurityGroup",
				"ec2:CreateTags",
				"ec2:DeleteTags",
				"ec2:Describe*",
				"ec2:ModifyImageAttribute",
				"ec2:ModifyInstanceAttribute",
				"ec2:RequestSpotInstances",
				"ec2:RunInstances",
				"ec2:StartInstances",
				"ec2:StopInstances",
				"ec2:TerminateInstances",
				"ec2:AuthorizeSecurityGroupEgress", 
				"ec2:DeleteSecurityGroup", 
				"ec2:RevokeSecurityGroupEgress", 
				"ec2:DescribeNetworkInterfaces", 
				"ec2:CreateNetworkInterface", 
				"ec2:DeleteNetworkInterface", 
				"ec2:DetachNetworkInterface",
				"elasticmapreduce:*",
				"iam:GetInstanceProfile",
				"iam:GetRole",
				"iam:GetRolePolicy",
				"iam:ListAttachedRolePolicies",
				"iam:ListRolePolicies",
				"iam:ListInstanceProfiles",
				"iam:PassRole",
				"rds:DescribeDBInstances",
				"rds:DescribeDBSecurityGroups",
				"redshift:DescribeClusters",
				"redshift:DescribeClusterSecurityGroups",
				"s3:CreateBucket",
				"s3:DeleteObject",
				"s3:Get*",
				"s3:List*",
				"s3:Put*",
				"sdb:BatchPutAttributes",
				"sdb:Select*",
				"sns:GetTopicAttributes",
				"sns:ListTopics",
				"sns:Publish",
				"sns:Subscribe",
				"sns:Unsubscribe",
				"sqs:CreateQueue", 
				"sqs:Delete*", 
				"sqs:GetQueue*", 
				"sqs:PurgeQueue", 
				"sqs:ReceiveMessage" 
			],
			"Resource": "*"
		},
		{
			"Effect": "Allow",
			"Action": "iam:CreateServiceLinkedRole",
			"Resource": "*",
			"Condition": {
			  "StringLike": {
				  "iam:AWSServiceName": ["elasticmapreduce.amazonaws.com","spot.amazonaws.com"]
			  }
			}
		}
	]
}
POLICY
}



resource "aws_iam_role" "resource_role" {
	name = "tf-test-datapipeline-resource-role"
	  
	assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Principal": {
				"Service": [
					"ec2.amazonaws.com"
				]
			},
			"Action": "sts:AssumeRole"
		}
	]
}
EOF
}

resource "aws_iam_instance_profile" "resource_role" {
	name = "tf-test-datapipeline-resource-role-profile"
	role = "${aws_iam_role.resource_role.name}"
}

resource "aws_iam_role_policy" "resource_role" {
	name = "tf-test-transfer-user-iam-policy"
	role = "${aws_iam_role.resource_role.id}"
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AllowFullAccesstoS3",
			"Effect": "Allow",
			"Action": [
				"cloudwatch:*",
				"datapipeline:*",
				"dynamodb:*",
				"ec2:Describe*",
				"elasticmapreduce:AddJobFlowSteps",
				"elasticmapreduce:Describe*",
				"elasticmapreduce:ListInstance*",
				"rds:Describe*",
				"redshift:DescribeClusters",
				"redshift:DescribeClusterSecurityGroups",
				"s3:*",
				"sdb:*",
				"sns:*",
				"sqs:*"
			],
			"Resource": "*"
		}
	]
}
POLICY
}


resource "aws_datapipeline" "foo" {
	name      = "tf-datapipeline"

	default {
		schedule_type          = "cron"
		schedule			   = "TestHourlySchedule"
		failure_and_rerun_mode = "cascade"
		role                   = "${aws_iam_role.role.arn}"
		resource_role          = "${aws_iam_instance_profile.resource_role.arn}"
	}

	copy_activity {
		id = "TestCopyActivity"
		name = "TestCopyActivity"
		schedule = "TestDailySchedule"
		runs_on = "TestEC2Resource"
		depends_on = "TestEC2Resource"
		input = "TestInputSqlDataNode"
		output = "TestOutputS3DataNode"
	}

	ec2_resource {
		id            = "TestEC2Resource"
		name          = "TestEC2Resource"
		instance_type = "t1.small"

		associate_public_ip_address = true
	}

	rds_database {
		id = "TestInputRdsDatabase"
		name = "TestInputRdsDatabase"
		rds_instance_id = "my_db_instance_identifier"
		username = "test"
		password = "test"
		database_name = "test"
	}

	sql_data_node {
		id = "TestInputSqlDataNode"
		name = "TestInputSqlDataNode"
		database = "TestInputRdsDatabase"
		table = "test"
		select_query = "SELECT * FROM #{table}"
	}

	s3_data_node {
		id = "TestOutputS3DataNode"
		name = "TestOutputS3DataNode"
		compression = "gzip"
		file_path = "s3://my-bucket/output/my-key-for-file"
	}

	schedule {
		id 				= "TestDailySchedule"
		name			= "TestDailySchedule"
		period			= "1 Day"
		start_date_time = "2019-01-01T00:00:00"
		end_date_time 	= "2019-09-01T00:00:00"
	}

	tags {
		NAME = "tf-datapipeline-test"
		ENV  = "test"
	}
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of Pipeline.
* `description` - (Optional) The description of Pipeline.
* `default` - (Required) The default configuration (documented below).
* `copy_activity` - (Optional) The activity configuration to CopyActivity (documented below).
* `ec2_resource` - (Optional) The resource configuration to EC2 (documented below).
* `rds_database` - (Optional) The database configuration to RDS (documented below).
* `s3_data_node` - (Optional) The data node configuration to S3 (documented below).
* `sql_data_node` - (Optional) The data node configuration to SQL (documented below).
* `schedule` - (Optional) The schedule configuration (documented below).
* `parameter_object` - (Optional) The Parameter Object configuration (documented below).
* `parameter_value` - (Optional) The Parameter Value configuration (documented below).
* `tags` - (Optional) A mapping of tags to assign to the resource.

The `default` resource configuration supports the following:

* `schedule_type` - (Required) The string of schedule type. Supported values are `cron`, `ondemand` and `timeseries`.
* `failure_and_rerun_mode` - (Optional) Describes consumer node behavior when dependencies fail or are rerun. Supported values are `cascade` and `none`. Default value is `none`.
* `pipeline_log_uri` - (Optional) The s3 URI for uploading logs for the pipeline.  (such as `s3://BucketName/Key/`) 
* `role` - (Required) The IAM role that AWS Data Pipeline uses to create the EC2 instance.
* `resource_role` - (Required) The IAM role that controls the resources that the Amazon EC2 instance can access.	
* `schedule` - (Optional) This object is invoked within the execution of a schedule interval.

The `copy_activity` activity configuration supports the following:
For more information, see the [AWS Data Pipeline CopyActivity Guide](https://docs.aws.amazon.com/datapipeline/latest/DeveloperGuide/dp-object-copyactivity.html).

* `id` - (Required) Specifies unique identifier for each Pipeline Object.
* `name` - (Required) The name of pipeline object.
* `schedule` - (Optional) This object is invoked within the execution of a schedule interval.
* `runs_on` - (Required) The computational resource to run the activity or command. For example, an Amazon EC2 instance or Amazon EMR cluster.
* `worker_group` - (Optional) The worker group. This is used for routing tasks. If you provide a `runs_on` value and `worker_group` exists, `worker_group` is ignored.
* `attempt_status` - (Optional) Most recently reported status from the remote activity.
* `attempt_timeout` - (Optional) Timeout for remote work completion. If set then a remote activity that does not complete within the set time of starting may be retried.
* `depends_on` - (Optional) Specify dependency on another runnable object.	
* `failure_and_rerun_mode` - (Optional) Describes consumer node behavior when dependencies fail or are rerun.
* `input` - (Optional) The input data source.
* `late_after_timeout` - (Optional) The elapsed time after pipeline start within which the object must start. It is triggered only when the schedule type is not set to `ondemand`.
* `max_active_instances` - (Optional) The maximum number of concurrent active instances of a component. Re-runs do not count toward the number of active instances.
* `maximum_retries` - (Optional) Maximum number attempt retries on failure.
* `on_fail` - (Optional) An action to run when current object fails.
* `on_late_action` - (Optional) Actions that should be triggered if an object has not yet been scheduled or still not completed.
* `on_success` - (Optional) An action to run when current object succeeds.
* `output` - (Optional) The output data source.
* `parent` - (Optional) Parent of the current object from which slots will be inherited.
* `pipeline_log_uri` - (Optional) The s3 URI for uploading logs for the pipeline.  (such as `s3://BucketName/Key/`) 
* `precondition` - (Optional) A data node is not marked `READY` until all preconditions have been met.
* `report_progress_timeout` - (Optional) Timeout for remote work successive calls to reportProgress. If set, then remote activities that do not report progress for the specified period may be considered stalled and so retried.
* `retry_delay` - (Optional) The timeout duration between two retry attempts.
* `schedule_type` - (Optional) The string of schedule type. Supported values are `cron`, `ondemand` and `timeseries`.

The `ec2_resource` resource configuration supports the following:
For more information, see the [AWS Data Pipeline Ec2Resource Guide](https://docs.aws.amazon.com/datapipeline/latest/DeveloperGuide/dp-object-ec2resource.html).

* `id` - (Required) Specifies unique identifier for each Pipeline Object.
* `name` - (Required) The name of pipeline object.
* `action_on_resource_failure` - (Optional) The action taken after a resource failure for this resource. Valid values are `retryall` and `retrynone`.
* `action_on_task_failure` - (Optional) The action taken after a task failure for this resource. Valid values are `continue` or `terminate`.	
* `associate_public_ip_address` - (Optional) ndicates whether to assign a public IP address to the instance. If the instance is in Amazon EC2 or Amazon VPC, the default value is `true`. Otherwise, the default value is `false`.
* `attempt_status` - (Optional) Most recently reported status from the remote activity.
* `attempt_timeout` - (Optional) Timeout for remote work completion. If set then a remote activity that does not complete within the set time of starting may be retried.
* `availability_zone` - (Optional) The Availability Zone in which to launch the Amazon EC2 instance.
* `image_id` - (Optional) The ID of the AMI to use for the instance. By default, AWS Data Pipeline uses the HVM AMI virtualization type. [The specific AMI IDs used are based on a Region.](https://docs.aws.amazon.com/datapipeline/latest/DeveloperGuide/dp-object-ec2resource.html)
* `init_timeout` - (Optional) The amount of time to wait for the resource to start.	
* `instance_type` - (Optional)  For more information, see the [Supported Instance Types for Pipeline Work Activities](https://docs.aws.amazon.com/datapipeline/latest/DeveloperGuide/dp-supported-instance-types.html) 
* `key_pair` - (Optional) The name of the key pair. If you launch an Amazon EC2 instance without specifying a key pair, you cannot log on to it.	
* `late_after_timeout` - (Optional) The elapsed time after pipeline start within which the object must start. It is triggered only when the schedule type is not set to `ondemand`.	
* `max_active_instances` - (Optional) The maximum number of concurrent active instances of a component. Re-runs do not count toward the number of active instances.
* `maximum_retries` - (Optional) The maximum number of attempt retries on failure.	
* `on_fail` - (Optional) An action to run when current object fails.
* `on_late_action` - (Optional) Actions that should be triggered if an object has not yet been scheduled or still not completed.
* `on_success` - (Optional) An action to run when current object succeeds.
* `pipeline_log_uri` - (Optional) The s3 URI for uploading logs for the pipeline.  (such as `s3://BucketName/Key/`) 
* `region` - (Optional) The code for the Region in which the Amazon EC2 instance should run. By default, the instance runs in the same Region as the pipeline.
* `schedule_type` - (Optional) The string of schedule type. Supported values are `cron`, `ondemand` and `timeseries`.
* `security_group_ids` - (Optional) The IDs of one or more Amazon EC2 security groups to use for the instances in the resource pool.
* `security_groups` - (Optional) One or more Amazon EC2 security groups to use for the instances in the resource pool.
* `spot_bid_price` - (Optional) The maximum amount per hour for your Spot Instance in dollars, which is a decimal value between `0` and `20.00`, exclusive.
* `subnet_id` - (Optional) The ID of the Amazon EC2 subnet in which to start the instance.	
* `terminate_after` - (Optional) The number of hours after which to terminate the resource.
* `use_on_demand_on_last_attempt` - (Optional) On the last attempt to request a Spot Instance, make a request for On-Demand Instances rather than a Spot Instance.

The `rds_database` resource configuration supports the following:
For more information, see the [AWS Data Pipeline RdsDatabase Guide](https://docs.aws.amazon.com/datapipeline/latest/DeveloperGuide/dp-object-rdsdatabase.html).

* `id` - (Required) Specifies unique identifier for each Pipeline Object.
* `name` - (Required) The name of pipeline object.
* `username` - (Required) The name of database user to access.
* `password` - (Required) The password of database user to access.
* `rds_instance_id` - (Required) The identifier of the DB instance.
* `database_name` - (Optional) The name of logical database to attach.
* `jdbc_driver_jar_uri` - (Optional) The location in Amazon S3 of the JDBC driver JAR file used to connect to the database. For the `MySQL` and `PostgreSQL` engines, the default driver is used if this field is not specified, but you can override the default using this field. For the `Oracle` and `SQL Server` engines, this field is required.
* `jdbc_properties` - (Optional) Pairs of the form A=B that will be set as properties on JDBC connections for this database.
* `parent` - (Optional) Parent of the current object from which slots will be inherited.
* `region` - (Optional) The code for the region where the database exists. 

The `s3_data_node` data node configuration supports the following:
For more information, see the [AWS Data Pipeline S3DataNode Guide](https://docs.aws.amazon.com/datapipeline/latest/DeveloperGuide/dp-object-s3datanode.html).

* `id` - (Required) Specifies unique identifier for each Pipeline Object.
* `name` - (Required) The name of pipeline object.
* `compression` - (Optional) The type of compression for the data described by the S3DataNode. `none` is no compression and `gzip` is compressed with the gzip algorithm. This field is only supported for use with Amazon Redshift and when you use S3DataNode with CopyActivity.
* `data_format` - (Optional) DataFormat for the data described by this S3DataNode.
* `depends_on` - (Optional) Specify dependency on another runnable object.
* `directory_path` - (Optional) Amazon S3 directory path as a URI: s3://my-bucket/my-key-for-directory. You must provide either a filePath or directoryPath value.
* `failure_and_rerun_mode` - (Optional) Describes consumer node behavior when dependencies fail or are rerun. Supported values are `cascade` and `none`. Default value is `none`.
* `file_path` - (Optional) The path to the object in Amazon S3 as a URI, for example: s3://my-bucket/my-key-for-file. You must provide either a filePath or directoryPath value. These represent a folder and a file name. Use the directoryPath value to accommodate multiple files in a directory.
* `late_after_timeout` - (Optional) The elapsed time after pipeline start within which the object must start. It is triggered only when the schedule type is not set to `ondemand`.
* `manifest_file_path` - (Optional) The Amazon S3 path to a manifest file in the format supported by Amazon Redshift. AWS Data Pipeline uses the manifest file to copy the specified Amazon S3 files into the table. This field is valid only when a RedShiftCopyActivity references the S3DataNode.
* `max_active_instances` - (Optional) The maximum number of concurrent active instances of a component. Re-runs do not count toward the number of active instances.
* `maximum_retries` - (Optional) Maximum number attempt retries on failure.
* `on_fail` - (Optional) An action to run when current object fails.
* `on_late_action` - (Optional) Actions that should be triggered if an object has not yet been scheduled or still not completed.
* `on_success` - (Optional) An action to run when current object succeeds.
* `parent` - (Optional) Parent of the current object from which slots will be inherited.
* `pipeline_log_uri` - (Optional) The s3 URI for uploading logs for the pipeline.  (such as `s3://BucketName/Key/`) 
* `precondition` - (Optional) A data node is not marked `READY` until all preconditions have been met.
* `report_progress_timeout` - (Optional) Timeout for remote work successive calls to reportProgress. If set, then remote activities that do not report progress for the specified period may be considered stalled and so retried.
* `retry_delay` - (Optional) The timeout duration between two retry attempts.
* `runs_on` - (Required) The computational resource to run the activity or command. For example, an Amazon EC2 instance or Amazon EMR cluster.
* `s3_encryption_type` - (Optional) Overrides the Amazon S3 encryption type. Values are `SERVER_SIDE_ENCRYPTION` or `NONE`. Server-side encryption is enabled by default.
* `schedule_type` - (Optional) The string of schedule type. Supported values are `cron`, `ondemand` and `timeseries`.
* `worker_group` - (Optional) The worker group. This is used for routing tasks. If you provide a `runs_on` value and `worker_group` exists, `worker_group` is ignored.

The `sql_data_node` data node configuration supports the following:
For more information, see the [AWS Data Pipeline SqlDataNode Guide](https://docs.aws.amazon.com/datapipeline/latest/DeveloperGuide/dp-object-sqldatanode.html).

* `id` - (Required) Specifies unique identifier for each Pipeline Object.
* `name` - (Required) The name of pipeline object.
* `table` - (Required) The name of the table in the SQL database.
* `create_table_sql` - (Optional) An SQL create table expression that creates the table.
* `database` - (Optional) The name of the database.
* `depends_on` - (Optional) Specify dependency on another runnable object.
* `failure_and_rerun_mode` - (Optional) Describes consumer node behavior when dependencies fail or are rerun. Supported values are `cascade` and `none`. Default value is `none`.
* `insert_query` - (Optional) An SQL statement to insert data into the table.
* `late_after_timeout` - (Optional) The elapsed time after pipeline start within which the object must start. It is triggered only when the schedule type is not set to `ondemand`.
* `max_active_instances` - (Optional) The maximum number of concurrent active instances of a component. Re-runs do not count toward the number of active instances.
* `maximum_retries` - (Optional) Maximum number attempt retries on failure.
* `on_fail` - (Optional) An action to run when current object fails.
* `on_late_action` - (Optional) Actions that should be triggered if an object has not yet been scheduled or still not completed.
* `on_success` - (Optional) An action to run when current object succeeds.
* `parent` - (Optional) Parent of the current object from which slots will be inherited.
* `pipeline_log_uri` - (Optional) The s3 URI for uploading logs for the pipeline.  (such as `s3://BucketName/Key/`) 
* `precondition` - (Optional) A data node is not marked `READY` until all preconditions have been met.
* `report_progress_timeout` - (Optional) Timeout for remote work successive calls to reportProgress. If set, then remote activities that do not report progress for the specified period may be considered stalled and so retried.
* `retry_delay` - (Optional) The timeout duration between two retry attempts.
* `runs_on` - (Required) The computational resource to run the activity or command. For example, an Amazon EC2 instance or Amazon EMR cluster.
* `schedule_type` - (Optional) The string of schedule type. Supported values are `cron`, `ondemand` and `timeseries`.
* `schema_name` - (Optional) The name of the schema holding the table.
* `select_query` - (Optional) A SQL statement to fetch data from the table.
* `worker_group` - (Optional) The worker group. This is used for routing tasks. If you provide a `runs_on` value and `worker_group` exists, `worker_group` is ignored.

The `schedule` configuration supports the following:
For more information, see the [AWS Data Pipeline Schedule Guide](https://docs.aws.amazon.com/datapipeline/latest/DeveloperGuide/dp-object-schedule.html).

* `id` - (Required) Specifies unique identifier for each Pipeline Object.
* `name` - (Required) The name of schedule pipeline object.
* `period` - (Required) Specifies how often the pipeline should run. The format is `N [minutes|hours|days|weeks|months]` The minimum period is 15 minutes and the maximum period is 3 years.
* `start_at` - (Optional) The date and time at which to start the scheduled pipeline runs. Valid value is FIRST_ACTIVATION_DATE_TIME, which is deprecated in favor of creating an on-demand pipeline.
* `start_date_time` - (Optional) The date and time to start the scheduled runs. You must use either startDateTime or startAt but not both.
* `end_date_time` - (Optional) The date and time to end the scheduled runs. Must be a date and time later than the value of startDateTime or startAt. The default behavior is to schedule runs until the pipeline is shut down.
* `occurrences` - (Optional) The number of times to execute the pipeline after it's activated. You can't use occurrences with endDateTime.
* `parent` - (Optional) Parent of the current object from which slots will be inherited.

The `parameter_object` configuration supports the following:
For more information, see the [AWS Data Pipeline Pipeline ParameterObjects](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-datapipeline-pipeline-parameterobjects.html).

* `id` - (Required) The identifier of the parameter object.
* `description` - (Optional) A description of the parameter.
* `type` - (Optional) The parameter type that defines the allowed range of input values and validation rules. The default is `String`.
* `optional` - (Optional) Indicates whether the parameter is optional or required. The default is `false`.
* `allowed_values` - (Optional) Enumerates all permitted values for the parameter.
* `default` - (Optional) The default value for the parameter. If you specify a value for this parameter using parameter values, it overrides the default value.
* `is_array` - (Optional) Indicates whether the parameter is an array.

The `parameter_value` configuration supports the following:
For more information, see the [AWS Data Pipeline User Guide](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-datapipeline-pipeline-parametervalues.html).

* `id` - (Required) The ID of a parameter object.
* `string_value` - (Required) A value to associate with the parameter object.