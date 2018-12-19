package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
)

func TestAccAWSDataPipeline_basic(t *testing.T) {
	var conf datapipeline.PipelineDescription
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataPipelineConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineExists("aws_datapipeline.foo", &conf),
					resource.TestCheckResourceAttrPair(
						"aws_datapipeline.foo", "default.role", "aws_iam_role.role", "arn"),
					resource.TestCheckResourceAttrPair(
						"aws_datapipeline.foo", "default.resource_role", "aws_iam_instance_profile.resource_role", "arn"),
				),
			},
			{
				ResourceName:      "aws_datapipeline.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDataPipeline_copyActivity(t *testing.T) {
	var conf datapipeline.PipelineDescription
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataPipelineConfig_copyActivity(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineExists("aws_datapipeline.foo", &conf),
					resource.TestCheckResourceAttrPair(
						"aws_datapipeline.foo", "default.role", "aws_iam_role.role", "arn"),
					resource.TestCheckResourceAttrPair(
						"aws_datapipeline.foo", "default.resource_role", "aws_iam_instance_profile.resource_role", "arn"),
					resource.TestCheckResourceAttrPair(
						"aws_datapipeline.foo", "copy_activity.0.runs_on", "aws_datapipeline.foo", "ec2_resource.0.id"),
					resource.TestCheckResourceAttrPair(
						"aws_datapipeline.foo", "rds_database.0.id", "aws_datapipeline.foo", "sql_data_node.0.database"),
					resource.TestCheckResourceAttrPair(
						"aws_datapipeline.foo", "copy_activity.0.input", "aws_datapipeline.foo", "sql_data_node.0.id"),
					resource.TestCheckResourceAttrPair(
						"aws_datapipeline.foo", "copy_activity.0.output", "aws_datapipeline.foo", "s3_data_node.0.id"),
					resource.TestCheckResourceAttr("aws_datapipeline.foo", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"aws_datapipeline.foo", "tags.NAME", "tf-datapipeline-test"),
					resource.TestCheckResourceAttr(
						"aws_datapipeline.foo", "tags.ENV", "test"),
				),
			},
			{
				ResourceName:      "aws_datapipeline.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDataPipeline_disappears(t *testing.T) {
	var conf datapipeline.PipelineDescription
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataPipelineConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineExists("aws_datapipeline.foo", &conf),
					testAccCheckAWSDataPipelineDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSDataPipelineDisappears(conf *datapipeline.PipelineDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).datapipelineconn
		params := &datapipeline.DeletePipelineInput{
			PipelineId: conf.PipelineId,
		}

		_, err := conn.DeletePipeline(params)
		if err != nil {
			return err
		}
		return waitForDataPipelineDeletion(conn, *conf.PipelineId)
	}
}

func testAccCheckAWSDataPipelineDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).datapipelineconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datapipeline" {
			continue
		}
		// Try to find the Pipeline
		var err error
		resp, err := conn.DescribePipelines(
			&datapipeline.DescribePipelinesInput{
				PipelineIds: []*string{aws.String(rs.Primary.ID)},
			})

		if err != nil {
			if isAWSErr(err, datapipeline.ErrCodePipelineNotFoundException, "") {
				continue
			} else if isAWSErr(err, datapipeline.ErrCodePipelineDeletedException, "") {
				continue
			}
			return err
		}
		if len(resp.PipelineDescriptionList) != 0 &&
			*resp.PipelineDescriptionList[0].PipelineId == rs.Primary.ID {
			return fmt.Errorf("Pipeline still exists")
		}
	}

	return nil
}

func testAccCheckAWSDataPipelineExists(n string, v *datapipeline.PipelineDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DataPipeline ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).datapipelineconn

		opts := &datapipeline.DescribePipelinesInput{
			PipelineIds: []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribePipelines(opts)
		if err != nil {
			return err
		}
		if len(resp.PipelineDescriptionList) != 1 ||
			*resp.PipelineDescriptionList[0].PipelineId != rs.Primary.ID {
			return fmt.Errorf("DataPipeline not found")
		}

		*v = *resp.PipelineDescriptionList[0]
		return nil
	}
}

func testAccAWSDataPipelineConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
	name = "tf-test-datapipeline-role-%s"
	  
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
	name = "tf-test-transfer-user-iam-policy-%s"
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
	name = "tf-test-datapipeline-resource-role-%s"
	  
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
	name = "tf-test-datapipeline-resource-role-profile-%s"
	role = "${aws_iam_role.resource_role.name}"
}

resource "aws_iam_role_policy" "resource_role" {
	name = "tf-test-transfer-user-iam-policy-%s"
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
}`, rName, rName, rName, rName, rName, rName)

}

func testAccAWSDataPipelineConfig_copyActivity(rName string) string {
	return fmt.Sprintf(`

resource "aws_iam_role" "role" {
	name = "tf-test-datapipeline-role-%s"
	  
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
	name = "tf-test-transfer-user-iam-policy-%s"
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
	name = "tf-test-datapipeline-resource-role-%s"
	  
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
	name = "tf-test-datapipeline-resource-role-profile-%s"
	role = "${aws_iam_role.resource_role.name}"
}

resource "aws_iam_role_policy" "resource_role" {
	name = "tf-test-transfer-user-iam-policy-%s"
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
	`, rName, rName, rName, rName, rName, rName)
}
