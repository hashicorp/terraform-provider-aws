---
subcategory: "EMR"
layout: "aws"
page_title: "AWS: aws_emr_cluster"
description: |-
  Provides an Elastic MapReduce Cluster
---

# Resource: aws_emr_cluster

Provides an Elastic MapReduce Cluster, a web service that makes it easy to process large amounts of data efficiently. See [Amazon Elastic MapReduce Documentation](https://aws.amazon.com/documentation/elastic-mapreduce/) for more information.

To configure [Instance Groups](https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-instance-group-configuration.html#emr-plan-instance-groups) for [task nodes](https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-master-core-task-nodes.html#emr-plan-task), see the [`aws_emr_instance_group` resource](/docs/providers/aws/r/emr_instance_group.html).

## Example Usage

```terraform
resource "aws_emr_cluster" "cluster" {
  name          = "emr-test-arn"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  additional_info = <<EOF
{
  "instanceAwsClientConfiguration": {
    "proxyPort": 8099,
    "proxyHost": "myproxy.example.com"
  }
}
EOF

  termination_protection            = false
  keep_job_flow_alive_when_no_steps = true

  ec2_attributes {
    subnet_id                         = aws_subnet.main.id
    emr_managed_master_security_group = aws_security_group.sg.id
    emr_managed_slave_security_group  = aws_security_group.sg.id
    instance_profile                  = aws_iam_instance_profile.emr_profile.arn
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_type  = "c4.large"
    instance_count = 1

    ebs_config {
      size                 = "40"
      type                 = "gp2"
      volumes_per_instance = 1
    }

    bid_price = "0.30"

    autoscaling_policy = <<EOF
{
"Constraints": {
  "MinCapacity": 1,
  "MaxCapacity": 2
},
"Rules": [
  {
    "Name": "ScaleOutMemoryPercentage",
    "Description": "Scale out if YARNMemoryAvailablePercentage is less than 15",
    "Action": {
      "SimpleScalingPolicyConfiguration": {
        "AdjustmentType": "CHANGE_IN_CAPACITY",
        "ScalingAdjustment": 1,
        "CoolDown": 300
      }
    },
    "Trigger": {
      "CloudWatchAlarmDefinition": {
        "ComparisonOperator": "LESS_THAN",
        "EvaluationPeriods": 1,
        "MetricName": "YARNMemoryAvailablePercentage",
        "Namespace": "AWS/ElasticMapReduce",
        "Period": 300,
        "Statistic": "AVERAGE",
        "Threshold": 15.0,
        "Unit": "PERCENT"
      }
    }
  }
]
}
EOF
  }

  ebs_root_volume_size = 100

  tags = {
    role = "rolename"
    env  = "env"
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations_json = <<EOF
  [
    {
      "Classification": "hadoop-env",
      "Configurations": [
        {
          "Classification": "export",
          "Properties": {
            "JAVA_HOME": "/usr/lib/jvm/java-1.8.0"
          }
        }
      ],
      "Properties": {}
    },
    {
      "Classification": "spark-env",
      "Configurations": [
        {
          "Classification": "export",
          "Properties": {
            "JAVA_HOME": "/usr/lib/jvm/java-1.8.0"
          }
        }
      ],
      "Properties": {}
    }
  ]
EOF

  service_role = aws_iam_role.iam_emr_service_role.arn
}
```

The `aws_emr_cluster` resource typically requires two IAM roles, one for the EMR Cluster to use as a service role, and another is assigned to every EC2 instance in a cluster and each application process that runs on a cluster assumes this role for permissions to interact with other AWS services. An additional role, the Auto Scaling role, is required if your cluster uses automatic scaling in Amazon EMR.

The default AWS managed EMR service role is called `EMR_DefaultRole` with Amazon managed policy `AmazonEMRServicePolicy_v2` attached. The name of default instance profile role is `EMR_EC2_DefaultRole` with default managed policy `AmazonElasticMapReduceforEC2Role` attached, but it is on the path to deprecation and will not be replaced with another default managed policy. You'll need to create and specify an instance profile to replace the deprecated role and default policy. See the [Configure IAM service roles for Amazon EMR](https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-iam-roles.html) guide for more information on these IAM roles. There is also a fully-bootable example Terraform configuration at the bottom of this page.

### Instance Fleet

```terraform
resource "aws_emr_cluster" "example" {
  # ... other configuration ...
  master_instance_fleet {
    instance_type_configs {
      instance_type = "m4.xlarge"
    }
    target_on_demand_capacity = 1
  }
  core_instance_fleet {
    instance_type_configs {
      bid_price_as_percentage_of_on_demand_price = 80
      ebs_config {
        size                 = 100
        type                 = "gp2"
        volumes_per_instance = 1
      }
      instance_type     = "m3.xlarge"
      weighted_capacity = 1
    }
    instance_type_configs {
      bid_price_as_percentage_of_on_demand_price = 100
      ebs_config {
        size                 = 100
        type                 = "gp2"
        volumes_per_instance = 1
      }
      instance_type     = "m4.xlarge"
      weighted_capacity = 1
    }
    instance_type_configs {
      bid_price_as_percentage_of_on_demand_price = 100
      ebs_config {
        size                 = 100
        type                 = "gp2"
        volumes_per_instance = 1
      }
      instance_type     = "m4.2xlarge"
      weighted_capacity = 2
    }
    launch_specifications {
      spot_specification {
        allocation_strategy      = "capacity-optimized"
        block_duration_minutes   = 0
        timeout_action           = "SWITCH_TO_ON_DEMAND"
        timeout_duration_minutes = 10
      }
    }
    name                      = "core fleet"
    target_on_demand_capacity = 2
    target_spot_capacity      = 2
  }
}

resource "aws_emr_instance_fleet" "task" {
  cluster_id = aws_emr_cluster.example.id
  instance_type_configs {
    bid_price_as_percentage_of_on_demand_price = 100
    ebs_config {
      size                 = 100
      type                 = "gp2"
      volumes_per_instance = 1
    }
    instance_type     = "m4.xlarge"
    weighted_capacity = 1
  }
  instance_type_configs {
    bid_price_as_percentage_of_on_demand_price = 100
    ebs_config {
      size                 = 100
      type                 = "gp2"
      volumes_per_instance = 1
    }
    instance_type     = "m4.2xlarge"
    weighted_capacity = 2
  }
  launch_specifications {
    spot_specification {
      allocation_strategy      = "capacity-optimized"
      block_duration_minutes   = 0
      timeout_action           = "TERMINATE_CLUSTER"
      timeout_duration_minutes = 10
    }
  }
  name                      = "task fleet"
  target_on_demand_capacity = 1
  target_spot_capacity      = 1
}
```

### Enable Debug Logging

[Debug logging in EMR](https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-plan-debugging.html) is implemented as a step. It is highly recommended that you utilize the [lifecycle configuration block](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html) with `ignore_changes` if other steps are being managed outside of Terraform.

```terraform
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  step {
    action_on_failure = "TERMINATE_CLUSTER"
    name              = "Setup Hadoop Debugging"

    hadoop_jar_step {
      jar  = "command-runner.jar"
      args = ["state-pusher-script"]
    }
  }

  # Optional: ignore outside changes to running cluster steps
  lifecycle {
    ignore_changes = [step]
  }
}
```

### Multiple Node Master Instance Group

Available in EMR version 5.23.0 and later, an EMR Cluster can be launched with three master nodes for high availability. Additional information about this functionality and its requirements can be found in the [EMR Management Guide](https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-plan-ha.html).

```terraform
# This configuration is for illustrative purposes and highlights
# only relevant configurations for working with this functionality.

# Map public IP on launch must be enabled for public (Internet accessible) subnets
resource "aws_subnet" "example" {
  # ... other configuration ...

  map_public_ip_on_launch = true
}

resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  # EMR version must be 5.23.0 or later
  release_label = "emr-5.24.1"

  # Termination protection is automatically enabled for multiple masters
  # To destroy the cluster, this must be configured to false and applied first
  termination_protection = true

  ec2_attributes {
    # ... other configuration ...

    subnet_id = aws_subnet.example.id
  }

  master_instance_group {
    # ... other configuration ...

    # Master instance count must be set to 3
    instance_count = 3
  }

  # core_instance_group must be configured
  core_instance_group {
    # ... other configuration ...
  }
}
```

### Bootable Cluster

**NOTE:** This configuration demonstrates a minimal configuration needed to boot an example EMR Cluster. It is not meant to display best practices. As with all examples, use at your own risk.

```terraform
resource "aws_emr_cluster" "cluster" {
  name          = "emr-test-arn"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = aws_subnet.main.id
    emr_managed_master_security_group = aws_security_group.allow_access.id
    emr_managed_slave_security_group  = aws_security_group.allow_access.id
    instance_profile                  = aws_iam_instance_profile.emr_profile.arn
  }

  master_instance_group {
    instance_type = "m5.xlarge"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "m5.xlarge"
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations_json = <<EOF
  [
    {
      "Classification": "hadoop-env",
      "Configurations": [
        {
          "Classification": "export",
          "Properties": {
            "JAVA_HOME": "/usr/lib/jvm/java-1.8.0"
          }
        }
      ],
      "Properties": {}
    },
    {
      "Classification": "spark-env",
      "Configurations": [
        {
          "Classification": "export",
          "Properties": {
            "JAVA_HOME": "/usr/lib/jvm/java-1.8.0"
          }
        }
      ],
      "Properties": {}
    }
  ]
EOF

  service_role = aws_iam_role.iam_emr_service_role.arn
}

resource "aws_security_group" "allow_access" {
  name        = "allow_access"
  description = "Allow inbound traffic"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [aws_vpc.main.cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = [aws_subnet.main]

  lifecycle {
    ignore_changes = [
      ingress,
      egress,
    ]
  }

  tags = {
    name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    name = "emr_test"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "168.31.0.0/20"

  tags = {
    name = "emr_test"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.main.id
}

resource "aws_route_table" "r" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.gw.id
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = aws_vpc.main.id
  route_table_id = aws_route_table.r.id
}

###

# IAM Role setups

###

# IAM role for EMR Service
data "aws_iam_policy_document" "emr_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.amazonaws.com"]
    }

    actions = "sts:AssumeRole"
  }
}

resource "aws_iam_role" "iam_emr_service_role" {
  name               = "iam_emr_service_role"
  assume_role_policy = data.aws_iam_policy_document.emr_assume_role.json
}

data "aws_iam_policy_document" "iam_emr_service_policy" {
  statement {
    effect = "Allow"

    actions = [
      "ec2:AuthorizeSecurityGroupEgress",
      "ec2:AuthorizeSecurityGroupIngress",
      "ec2:CancelSpotInstanceRequests",
      "ec2:CreateNetworkInterface",
      "ec2:CreateSecurityGroup",
      "ec2:CreateTags",
      "ec2:DeleteNetworkInterface",
      "ec2:DeleteSecurityGroup",
      "ec2:DeleteTags",
      "ec2:DescribeAvailabilityZones",
      "ec2:DescribeAccountAttributes",
      "ec2:DescribeDhcpOptions",
      "ec2:DescribeInstanceStatus",
      "ec2:DescribeInstances",
      "ec2:DescribeKeyPairs",
      "ec2:DescribeNetworkAcls",
      "ec2:DescribeNetworkInterfaces",
      "ec2:DescribePrefixLists",
      "ec2:DescribeRouteTables",
      "ec2:DescribeSecurityGroups",
      "ec2:DescribeSpotInstanceRequests",
      "ec2:DescribeSpotPriceHistory",
      "ec2:DescribeSubnets",
      "ec2:DescribeVpcAttribute",
      "ec2:DescribeVpcEndpoints",
      "ec2:DescribeVpcEndpointServices",
      "ec2:DescribeVpcs",
      "ec2:DetachNetworkInterface",
      "ec2:ModifyImageAttribute",
      "ec2:ModifyInstanceAttribute",
      "ec2:RequestSpotInstances",
      "ec2:RevokeSecurityGroupEgress",
      "ec2:RunInstances",
      "ec2:TerminateInstances",
      "ec2:DeleteVolume",
      "ec2:DescribeVolumeStatus",
      "ec2:DescribeVolumes",
      "ec2:DetachVolume",
      "iam:GetRole",
      "iam:GetRolePolicy",
      "iam:ListInstanceProfiles",
      "iam:ListRolePolicies",
      "iam:PassRole",
      "s3:CreateBucket",
      "s3:Get*",
      "s3:List*",
      "sdb:BatchPutAttributes",
      "sdb:Select",
      "sqs:CreateQueue",
      "sqs:Delete*",
      "sqs:GetQueue*",
      "sqs:PurgeQueue",
      "sqs:ReceiveMessage",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "iam_emr_service_policy" {
  name   = "iam_emr_service_policy"
  role   = aws_iam_role.iam_emr_service_role.id
  policy = data.aws_iam_policy_document.iam_emr_service_policy.json
}

# IAM Role for EC2 Instance Profile
data "aws_iam_policy_document" "ec2_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }

    actions = "sts:AssumeRole"
  }
}

resource "aws_iam_role" "iam_emr_profile_role" {
  name               = "iam_emr_profile_role"
  assume_role_policy = data.aws_iam_policy_document.ec2_assume_role.json
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile"
  role = aws_iam_role.iam_emr_profile_role.name
}

data "aws_iam_policy_document" "iam_emr_profile_policy" {
  statement {
    effect = "Allow"

    actions = [
      "cloudwatch:*",
      "dynamodb:*",
      "ec2:Describe*",
      "elasticmapreduce:Describe*",
      "elasticmapreduce:ListBootstrapActions",
      "elasticmapreduce:ListClusters",
      "elasticmapreduce:ListInstanceGroups",
      "elasticmapreduce:ListInstances",
      "elasticmapreduce:ListSteps",
      "kinesis:CreateStream",
      "kinesis:DeleteStream",
      "kinesis:DescribeStream",
      "kinesis:GetRecords",
      "kinesis:GetShardIterator",
      "kinesis:MergeShards",
      "kinesis:PutRecord",
      "kinesis:SplitShard",
      "rds:Describe*",
      "s3:*",
      "sdb:*",
      "sns:*",
      "sqs:*",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "iam_emr_profile_policy" {
  name   = "iam_emr_profile_policy"
  role   = aws_iam_role.iam_emr_profile_role.id
  policy = data.aws_iam_policy_document.iam_emr_profile_policy.json
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the job flow.
* `release_label` - (Required) Release label for the Amazon EMR release.
* `service_role` - (Required) IAM role that will be assumed by the Amazon EMR service to access AWS resources.

The following arguments are optional:

* `additional_info` - (Optional) JSON string for selecting additional features such as adding proxy information. Note: Currently there is no API to retrieve the value of this argument after EMR cluster creation from provider, therefore Terraform cannot detect drift from the actual EMR cluster if its value is changed outside Terraform.
* `applications` - (Optional) A case-insensitive list of applications for Amazon EMR to install and configure when launching the cluster. For a list of applications available for each Amazon EMR release version, see the [Amazon EMR Release Guide](https://docs.aws.amazon.com/emr/latest/ReleaseGuide/emr-release-components.html).
* `autoscaling_role` - (Optional) IAM role for automatic scaling policies. The IAM role provides permissions that the automatic scaling feature requires to launch and terminate EC2 instances in an instance group.
* `auto_termination_policy` - (Optional) An auto-termination policy for an Amazon EMR cluster. An auto-termination policy defines the amount of idle time in seconds after which a cluster automatically terminates. See [Auto Termination Policy](#auto_termination_policy) Below.
* `bootstrap_action` - (Optional) Ordered list of bootstrap actions that will be run before Hadoop is started on the cluster nodes. See below.
* `configurations` - (Optional) List of configurations supplied for the EMR cluster you are creating. Supply a configuration object for applications to override their default configuration. See [AWS Documentation](https://docs.aws.amazon.com/emr/latest/ReleaseGuide/emr-configure-apps.html) for more information.
* `configurations_json` - (Optional) JSON string for supplying list of configurations for the EMR cluster.

~> **NOTE on `configurations_json`:** If the `Configurations` value is empty then you should skip the `Configurations` field instead of providing an empty list as a value, `"Configurations": []`.

```terraform
resource "aws_emr_cluster" "cluster" {
  # ... other configuration ...

  configurations_json = <<EOF
  [
    {
      "Classification": "hadoop-env",
      "Configurations": [
        {
          "Classification": "export",
          "Properties": {
            "JAVA_HOME": "/usr/lib/jvm/java-1.8.0"
          }
        }
      ],
      "Properties": {}
    }
  ]
EOF
}
```

* `core_instance_fleet` - (Optional) Configuration block to use an [Instance Fleet](https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-instance-fleet.html) for the core node type. Cannot be specified if any `core_instance_group` configuration blocks are set. Detailed below.
* `core_instance_group` - (Optional) Configuration block to use an [Instance Group](https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-instance-group-configuration.html#emr-plan-instance-groups) for the [core node type](https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-master-core-task-nodes.html#emr-plan-core).
* `custom_ami_id` - (Optional) Custom Amazon Linux AMI for the cluster (instead of an EMR-owned AMI). Available in Amazon EMR version 5.7.0 and later.
* `ebs_root_volume_size` - (Optional) Size in GiB of the EBS root device volume of the Linux AMI that is used for each EC2 instance. Available in Amazon EMR version 4.x and later.
* `ec2_attributes` - (Optional) Attributes for the EC2 instances running the job flow. See below.
* `keep_job_flow_alive_when_no_steps` - (Optional) Switch on/off run cluster with no steps or when all steps are complete (default is on)
* `kerberos_attributes` - (Optional) Kerberos configuration for the cluster. See below.
* `list_steps_states` - (Optional) List of [step states](https://docs.aws.amazon.com/emr/latest/APIReference/API_StepStatus.html) used to filter returned steps
* `log_encryption_kms_key_id` - (Optional) AWS KMS customer master key (CMK) key ID or arn used for encrypting log files. This attribute is only available with EMR version 5.30.0 and later, excluding EMR 6.0.0.
* `log_uri` - (Optional) S3 bucket to write the log files of the job flow. If a value is not provided, logs are not created.
* `master_instance_fleet` - (Optional) Configuration block to use an [Instance Fleet](https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-instance-fleet.html) for the master node type. Cannot be specified if any `master_instance_group` configuration blocks are set. Detailed below.
* `master_instance_group` - (Optional) Configuration block to use an [Instance Group](https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-instance-group-configuration.html#emr-plan-instance-groups) for the [master node type](https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-master-core-task-nodes.html#emr-plan-master).
* `placement_group_config` - (Optional) The specified placement group configuration for an Amazon EMR cluster.
* `scale_down_behavior` - (Optional) Way that individual Amazon EC2 instances terminate when an automatic scale-in activity occurs or an `instance group` is resized.
* `security_configuration` - (Optional) Security configuration name to attach to the EMR cluster. Only valid for EMR clusters with `release_label` 4.8.0 or greater.
* `step` - (Optional) List of steps to run when creating the cluster. See below. It is highly recommended to utilize the [lifecycle configuration block](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html) with `ignore_changes` if other steps are being managed outside of Terraform. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).
* `step_concurrency_level` - (Optional) Number of steps that can be executed concurrently. You can specify a maximum of 256 steps. Only valid for EMR clusters with `release_label` 5.28.0 or greater (default is 1).
* `tags` - (Optional) list of tags to apply to the EMR Cluster. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `termination_protection` - (Optional) Switch on/off termination protection (default is `false`, except when using multiple master nodes). Before attempting to destroy the resource when termination protection is enabled, this configuration must be applied with its value set to `false`.
* `unhealthy_node_replacement` - (Optional) Whether whether Amazon EMR should gracefully replace core nodes that have degraded within the cluster. Default value is `false`.
* `visible_to_all_users` - (Optional) Whether the job flow is visible to all IAM users of the AWS account associated with the job flow. Default value is `true`.

### bootstrap_action

* `args` - (Optional) List of command line arguments to pass to the bootstrap action script.
* `name` - (Required) Name of the bootstrap action.
* `path` - (Required) Location of the script to run during a bootstrap action. Can be either a location in Amazon S3 or on a local file system.

### auto_termination_policy

* `idle_timeout` - (Optional) Specifies the amount of idle time in seconds after which the cluster automatically terminates. You can specify a minimum of `60` seconds and a maximum of `604800` seconds (seven days).

### configurations

A configuration classification that applies when provisioning cluster instances, which can include configurations for applications and software that run on the cluster. See [Configuring Applications](https://docs.aws.amazon.com/emr/latest/ReleaseGuide/emr-configure-apps.html).

* `classification` - (Optional) Classification within a configuration.
* `properties` - (Optional) Map of properties specified within a configuration classification.

### core_instance_fleet

* `instance_type_configs` - (Optional) Configuration block for instance fleet.
* `launch_specifications` - (Optional) Configuration block for launch specification.
* `name` - (Optional) Friendly name given to the instance fleet.
* `target_on_demand_capacity` - (Optional)  The target capacity of On-Demand units for the instance fleet, which determines how many On-Demand instances to provision.
* `target_spot_capacity` - (Optional) Target capacity of Spot units for the instance fleet, which determines how many Spot instances to provision.

#### instance_type_configs

* `bid_price` - (Optional) Bid price for each EC2 Spot instance type as defined by `instance_type`. Expressed in USD. If neither `bid_price` nor `bid_price_as_percentage_of_on_demand_price` is provided, `bid_price_as_percentage_of_on_demand_price` defaults to 100%.
* `bid_price_as_percentage_of_on_demand_price` - (Optional) Bid price, as a percentage of On-Demand price, for each EC2 Spot instance as defined by `instance_type`. Expressed as a number (for example, 20 specifies 20%). If neither `bid_price` nor `bid_price_as_percentage_of_on_demand_price` is provided, `bid_price_as_percentage_of_on_demand_price` defaults to 100%.
* `configurations` - (Optional) Configuration classification that applies when provisioning cluster instances, which can include configurations for applications and software that run on the cluster. List of `configuration` blocks.
* `ebs_config` - (Optional) Configuration block(s) for EBS volumes attached to each instance in the instance group. Detailed below.
* `instance_type` - (Required) EC2 instance type, such as m4.xlarge.
* `weighted_capacity` - (Optional) Number of units that a provisioned instance of this type provides toward fulfilling the target capacities defined in `aws_emr_instance_fleet`.

#### launch_specifications

* `on_demand_specification` - (Optional) Configuration block for on demand instances launch specifications.
* `spot_specification` - (Optional) Configuration block for spot instances launch specifications.

##### on_demand_specification

The launch specification for On-Demand instances in the instance fleet, which determines the allocation strategy.
The instance fleet configuration is available only in Amazon EMR versions 4.8.0 and later, excluding 5.0.x versions. On-Demand instances allocation strategy is available in Amazon EMR version 5.12.1 and later.

* `allocation_strategy` - (Required) Specifies the strategy to use in launching On-Demand instance fleets. Currently, the only option is `lowest-price` (the default), which launches the lowest price first.

##### spot_specification

The launch specification for Spot instances in the fleet, which determines the defined duration, provisioning timeout behavior, and allocation strategy.

* `allocation_strategy` - (Required) Specifies the strategy to use in launching Spot instance fleets. Valid values include `capacity-optimized`, `diversified`, `lowest-price`, `price-capacity-optimized`. See the [AWS documentation](https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-instance-fleet.html#emr-instance-fleet-allocation-strategy) for details on each strategy type.
* `block_duration_minutes` - (Optional) Defined duration for Spot instances (also known as Spot blocks) in minutes. When specified, the Spot instance does not terminate before the defined duration expires, and defined duration pricing for Spot instances applies. Valid values are 60, 120, 180, 240, 300, or 360. The duration period starts as soon as a Spot instance receives its instance ID. At the end of the duration, Amazon EC2 marks the Spot instance for termination and provides a Spot instance termination notice, which gives the instance a two-minute warning before it terminates.
* `timeout_action` - (Required) Action to take when TargetSpotCapacity has not been fulfilled when the TimeoutDurationMinutes has expired; that is, when all Spot instances could not be provisioned within the Spot provisioning timeout. Valid values are `TERMINATE_CLUSTER` and `SWITCH_TO_ON_DEMAND`. SWITCH_TO_ON_DEMAND specifies that if no Spot instances are available, On-Demand Instances should be provisioned to fulfill any remaining Spot capacity.
* `timeout_duration_minutes` - (Required) Spot provisioning timeout period in minutes. If Spot instances are not provisioned within this time period, the TimeOutAction is taken. Minimum value is 5 and maximum value is 1440. The timeout applies only during initial provisioning, when the cluster is first created.

### core_instance_group

* `autoscaling_policy` - (Optional) String containing the [EMR Auto Scaling Policy](https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-automatic-scaling.html) JSON.
* `bid_price` - (Optional) Bid price for each EC2 instance in the instance group, expressed in USD. By setting this attribute, the instance group is being declared as a Spot Instance, and will implicitly create a Spot request. Leave this blank to use On-Demand Instances.
* `ebs_config` - (Optional) Configuration block(s) for EBS volumes attached to each instance in the instance group. Detailed below.
* `instance_count` - (Optional) Target number of instances for the instance group. Must be at least 1. Defaults to 1.
* `instance_type` - (Required) EC2 instance type for all instances in the instance group.
* `name` - (Optional) Friendly name given to the instance group.

#### ebs_config

* `iops` - (Optional) Number of I/O operations per second (IOPS) that the volume supports.
* `size` - (Required) Volume size, in gibibytes (GiB).
* `type` - (Required) Volume type. Valid options are `gp3`, `gp2`, `io1`, `standard`, `st1` and `sc1`. See [EBS Volume Types](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumeTypes.html).
* `throughput` - (Optional) The throughput, in mebibyte per second (MiB/s).
* `volumes_per_instance` - (Optional) Number of EBS volumes with this configuration to attach to each EC2 instance in the instance group (default is 1).

### ec2_attributes

Attributes for the Amazon EC2 instances running the job flow:

* `additional_master_security_groups` - (Optional) String containing a comma separated list of additional Amazon EC2 security group IDs for the master node.
* `additional_slave_security_groups` - (Optional) String containing a comma separated list of additional Amazon EC2 security group IDs for the slave nodes as a comma separated string.
* `emr_managed_master_security_group` - (Optional) Identifier of the Amazon EC2 EMR-Managed security group for the master node.
* `emr_managed_slave_security_group` - (Optional) Identifier of the Amazon EC2 EMR-Managed security group for the slave nodes.
* `instance_profile` - (Required) Instance Profile for EC2 instances of the cluster assume this role.
* `key_name` - (Optional) Amazon EC2 key pair that can be used to ssh to the master node as the user called `hadoop`.
* `service_access_security_group` - (Optional) Identifier of the Amazon EC2 service-access security group - required when the cluster runs on a private subnet.
* `subnet_id` - (Optional) VPC subnet id where you want the job flow to launch. Cannot specify the `cc1.4xlarge` instance type for nodes of a job flow launched in an Amazon VPC.
* `subnet_ids` - (Optional) List of VPC subnet id-s where you want the job flow to launch.  Amazon EMR identifies the best Availability Zone to launch instances according to your fleet specifications.

~> **NOTE on EMR-Managed security groups:** These security groups will have any missing inbound or outbound access rules added and maintained by AWS, to ensure proper communication between instances in a cluster. The EMR service will maintain these rules for groups provided in `emr_managed_master_security_group` and `emr_managed_slave_security_group`; attempts to remove the required rules may succeed, only for the EMR service to re-add them in a matter of minutes. This may cause Terraform to fail to destroy an environment that contains an EMR cluster, because the EMR service does not revoke rules added on deletion, leaving a cyclic dependency between the security groups that prevents their deletion. To avoid this, use the `revoke_rules_on_delete` optional attribute for any Security Group used in `emr_managed_master_security_group` and `emr_managed_slave_security_group`. See [Amazon EMR-Managed Security Groups](http://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-man-sec-groups.html) for more information about the EMR-managed security group rules.

### kerberos_attributes

* `ad_domain_join_password` - (Optional) Active Directory password for `ad_domain_join_user`. Terraform cannot perform drift detection of this configuration.
* `ad_domain_join_user` - (Optional) Required only when establishing a cross-realm trust with an Active Directory domain. A user with sufficient privileges to join resources to the domain. Terraform cannot perform drift detection of this configuration.
* `cross_realm_trust_principal_password` - (Optional) Required only when establishing a cross-realm trust with a KDC in a different realm. The cross-realm principal password, which must be identical across realms. Terraform cannot perform drift detection of this configuration.
* `kdc_admin_password` - (Required) Password used within the cluster for the kadmin service on the cluster-dedicated KDC, which maintains Kerberos principals, password policies, and keytabs for the cluster. Terraform cannot perform drift detection of this configuration.
* `realm` - (Required) Name of the Kerberos realm to which all nodes in a cluster belong. For example, `EC2.INTERNAL`

### master_instance_fleet

* `instance_type_configs` - (Optional) Configuration block for instance fleet.
* `launch_specifications` - (Optional) Configuration block for launch specification.
* `name` - (Optional) Friendly name given to the instance fleet.
* `target_on_demand_capacity` - (Optional) Target capacity of On-Demand units for the instance fleet, which determines how many On-Demand instances to provision.
* `target_spot_capacity` - (Optional) Target capacity of Spot units for the instance fleet, which determines how many Spot instances to provision.

#### instance_type_configs

See `instance_type_configs` above, under `core_instance_fleet`.

#### launch_specifications

See `launch_specifications` above, under `core_instance_fleet`.

### master_instance_group

Supported nested arguments for the `master_instance_group` configuration block:

* `bid_price` - (Optional) Bid price for each EC2 instance in the instance group, expressed in USD. By setting this attribute, the instance group is being declared as a Spot Instance, and will implicitly create a Spot request. Leave this blank to use On-Demand Instances.
* `ebs_config` - (Optional) Configuration block(s) for EBS volumes attached to each instance in the instance group. Detailed below.
* `instance_count` - (Optional) Target number of instances for the instance group. Must be 1 or 3. Defaults to 1. Launching with multiple master nodes is only supported in EMR version 5.23.0+, and requires this resource's `core_instance_group` to be configured. Public (Internet accessible) instances must be created in VPC subnets that have [map public IP on launch](/docs/providers/aws/r/subnet.html#map_public_ip_on_launch) enabled. Termination protection is automatically enabled when launched with multiple master nodes and Terraform must have the `termination_protection = false` configuration applied before destroying this resource.
* `instance_type` - (Required) EC2 instance type for all instances in the instance group.
* `name` - (Optional) Friendly name given to the instance group.

#### ebs_config

See `ebs_config` under `core_instance_group` above.

### step

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

* `action_on_failure` - (Required) Action to take if the step fails. Valid values: `TERMINATE_JOB_FLOW`, `TERMINATE_CLUSTER`, `CANCEL_AND_WAIT`, and `CONTINUE`
* `hadoop_jar_step` - (Required) JAR file used for the step. See below.
* `name` - (Required) Name of the step.

#### hadoop_jar_step

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

* `args` - (Optional) List of command line arguments passed to the JAR file's main function when executed.
* `jar` - (Required) Path to a JAR file run during the step.
* `main_class` - (Optional) Name of the main class in the specified Java file. If not specified, the JAR file should specify a Main-Class in its manifest file.
* `properties` - (Optional) Key-Value map of Java properties that are set when the step runs. You can use these properties to pass key value pairs to your main function.

### placement_group_config

* `instance_role` - (Required) Role of the instance in the cluster. Valid Values: `MASTER`, `CORE`, `TASK`.
* `placement_strategy` - (Optional) EC2 Placement Group strategy associated with instance role. Valid Values: `SPREAD`, `PARTITION`, `CLUSTER`, `NONE`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `applications` - Applications installed on this cluster.
* `arn`- ARN of the cluster.
* `bootstrap_action` - List of bootstrap actions that will be run before Hadoop is started on the cluster nodes.
* `configurations` - List of Configurations supplied to the EMR cluster.
* `core_instance_group.0.id` - Core node type Instance Group ID, if using Instance Group for this node type.
* `ec2_attributes` - Provides information about the EC2 instances in a cluster grouped by category: key name, subnet ID, IAM instance profile, and so on.
* `id` - ID of the cluster.
* `log_uri` - Path to the Amazon S3 location where logs for this cluster are stored.
* `master_instance_group.0.id` - Master node type Instance Group ID, if using Instance Group for this node type.
* `master_public_dns` - The DNS name of the master node. If the cluster is on a private subnet, this is the private DNS name. On a public subnet, this is the public DNS name.
* `name` - Name of the cluster.
* `release_label` - Release label for the Amazon EMR release.
* `service_role` - IAM role that will be assumed by the Amazon EMR service to access AWS resources on your behalf.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `visible_to_all_users` - Indicates whether the job flow is visible to all IAM users of the AWS account associated with the job flow.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EMR clusters using the `id`. For example:

```terraform
import {
  to = aws_emr_cluster.cluster
  id = "j-123456ABCDEF"
}
```

Using `terraform import`, import EMR clusters using the `id`. For example:

```console
% terraform import aws_emr_cluster.cluster j-123456ABCDEF
```

Since the API does not return the actual values for Kerberos configurations, environments with those Terraform configurations will need to use the [`lifecycle` configuration block `ignore_changes` argument](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) available to all Terraform resources to prevent perpetual differences. For example:

```terraform
resource "aws_emr_cluster" "example" {
  # ... other configuration ...

  lifecycle {
    ignore_changes = [kerberos_attributes]
  }
}
```
