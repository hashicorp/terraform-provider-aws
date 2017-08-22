package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSEMRInstanceFleet_basic(t *testing.T) {
	var fleet emr.InstanceFleet
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceFleetConfig(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckAWSEmrInstanceFleetExists("aws_emr_instance_fleet.task", &fleet),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "instance_fleet_type", "TASK"),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "instance_type_configs.#", "1"),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "target_on_demand_capacity", "0"),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "target_spot_capacity", "1"),
				),
			},
		},
	})
}

func TestAccAWSEMRInstanceFleet_zero_count(t *testing.T) {
	var fleet emr.InstanceFleet
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceFleetConfig(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckAWSEmrInstanceFleetExists("aws_emr_instance_fleet.task", &fleet),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "instance_fleet_type", "TASK"),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "instance_type_configs.#", "1"),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "target_on_demand_capacity", "0"),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "target_spot_capacity", "1"),
				),
			},
			{
				Config: testAccAWSEmrInstanceFleetConfigZeroCount(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckAWSEmrInstanceFleetExists("aws_emr_instance_fleet.task", &fleet),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "instance_fleet_type", "TASK"),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "instance_type_configs.#", "1"),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "target_on_demand_capacity", "0"),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "target_spot_capacity", "0"),
				),
			},
		},
	})
}

func TestAccAWSEMRInstanceFleet_ebsBasic(t *testing.T) {
	var fleet emr.InstanceFleet
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceFleetConfigEbsBasic(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckAWSEmrInstanceFleetExists("aws_emr_instance_fleet.task", &fleet),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "instance_fleet_type", "TASK"),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "instance_type_configs.#", "1"),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "target_on_demand_capacity", "0"),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "target_spot_capacity", "1"),
				),
			},
		},
	})
}

func TestAccAWSEMRInstanceFleet_full(t *testing.T) {
	var fleet emr.InstanceFleet
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrInstanceFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrInstanceFleetConfigFull(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckAWSEmrInstanceFleetExists("aws_emr_instance_fleet.task", &fleet),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "instance_fleet_type", "TASK"),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "instance_type_configs.#", "1"),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "target_on_demand_capacity", "1"),
					resource.TestCheckResourceAttr("aws_emr_instance_fleet.task", "target_spot_capacity", "1"),
				),
			},
		},
	})
}

func testAccCheckAWSEmrInstanceFleetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).emrconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_emr_cluster" {
			continue
		}

		params := &emr.DescribeClusterInput{
			ClusterId: aws.String(rs.Primary.ID),
		}

		describe, err := conn.DescribeCluster(params)

		if err == nil {
			if describe.Cluster != nil &&
				*describe.Cluster.Status.State == "WAITING" {
				return fmt.Errorf("EMR Cluster still exists")
			}
		}

		providerErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}

		log.Printf("[ERROR] %v", providerErr)
	}

	return nil
}

func testAccCheckAWSEmrInstanceFleetExists(n string, v *emr.InstanceFleet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No task fleet id set")
		}
		meta := testAccProvider.Meta()
		conn := meta.(*AWSClient).emrconn
		f, err := fetchEMRInstanceFleet(conn, rs.Primary.Attributes["cluster_id"], rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("EMR error: %v", err)
		}

		if f == nil {
			return fmt.Errorf("No match found for (%s)", n)
		}

		v = f
		return nil
	}
}

const testAccAWSEmrInstanceFleetBase = `
provider "aws" {
	region = "us-west-2"
}

resource "aws_emr_cluster" "tf-test-cluster" {
	name          = "emr-test-%d"
	release_label = "emr-5.10.0"
	applications  = ["Spark"]

	ec2_attributes {
		subnet_id                         = "${aws_subnet.main.id}"
		emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
		emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
		instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
	}

	instance_fleet = [
		{
			instance_fleet_type = "MASTER"
			instance_type_configs = [
				{
					instance_type = "m3.xlarge"
				}
			]
			target_on_demand_capacity = 1
		},
		{
			instance_fleet_type = "CORE"
			instance_type_configs = [
				{
					bid_price_as_percentage_of_on_demand_price = 80
					ebs_optimized = true
					ebs_config = [
						{
							size = 100
							type = "gp2"
							volumes_per_instance = 1
						}
					]
					instance_type = "m3.xlarge"
					weighted_capacity = 1
				}
			]
			launch_specifications {
				spot_specification {
					block_duration_minutes   = 60
					timeout_action           = "SWITCH_TO_ON_DEMAND"
					timeout_duration_minutes = 10
				}
			}
			name                      = "instance-fleet-test"
			target_on_demand_capacity = 0
			target_spot_capacity      = 1
		}
	]

	tags {
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

	configurations = "test-fixtures/emr_configurations.json"
	service_role = "${aws_iam_role.iam_emr_default_role.arn}"

	depends_on = ["aws_internet_gateway.gw"]
}

resource "aws_security_group" "allow_all" {
	name        = "allow_all"
	description = "Allow all inbound traffic"
	vpc_id      = "${aws_vpc.main.id}"

	ingress {
		from_port   = 0
		to_port     = 0
		protocol    = "-1"
		cidr_blocks = ["0.0.0.0/0"]
	}

	egress {
		from_port   = 0
		to_port     = 0
		protocol    = "-1"
		cidr_blocks = ["0.0.0.0/0"]
	}

	depends_on = ["aws_subnet.main"]

	lifecycle {
		ignore_changes = ["ingress", "egress"]
	}
}

resource "aws_vpc" "main" {
	cidr_block           = "168.31.0.0/16"
	enable_dns_hostnames = true

	tags {
		Name = "tf_acc_emr_tests"
	}
}

resource "aws_subnet" "main" {
	vpc_id     = "${aws_vpc.main.id}"
	cidr_block = "168.31.0.0/20"

	#  map_public_ip_on_launch = true
}

resource "aws_internet_gateway" "gw" {
	vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
	vpc_id = "${aws_vpc.main.id}"

	route {
		cidr_block = "0.0.0.0/0"
		gateway_id = "${aws_internet_gateway.gw.id}"
	}
}

resource "aws_main_route_table_association" "a" {
	vpc_id         = "${aws_vpc.main.id}"
	route_table_id = "${aws_route_table.r.id}"
}

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
	name = "iam_emr_default_role_%d"

	assume_role_policy = <<EOT
{
	"Version": "2008-10-17",
	"Statement": [
		{
			"Sid": "",
			"Effect": "Allow",
			"Principal": {
				"Service": "elasticmapreduce.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		}
	]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
	role       = "${aws_iam_role.iam_emr_default_role.id}"
	policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
	name = "iam_emr_default_policy_%d"

	policy = <<EOT
{
	"Version": "2012-10-17",
	"Statement": [{
		"Effect": "Allow",
		"Resource": "*",
		"Action": [
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
			"sqs:ReceiveMessage"
		]
	}]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
	name = "iam_emr_profile_role_%d"

	assume_role_policy = <<EOT
{
	"Version": "2008-10-17",
	"Statement": [
		{
			"Sid": "",
			"Effect": "Allow",
			"Principal": {
				"Service": "ec2.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		}
	]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
	name = "emr_profile_%d"
	role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
	role       = "${aws_iam_role.iam_emr_profile_role.id}"
	policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
	name = "iam_emr_profile_policy_%d"

	policy = <<EOT
{
	"Version": "2012-10-17",
	"Statement": [{
		"Effect": "Allow",
		"Resource": "*",
		"Action": [
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
			"sqs:*"
		]
	}]
}
EOT
}
`

func testAccAWSEmrInstanceFleetConfig(r int) string {
	return fmt.Sprintf(testAccAWSEmrInstanceFleetBase+`
	resource "aws_emr_instance_fleet" "task" {
		cluster_id            = "${aws_emr_cluster.tf-test-cluster.id}"
		instance_fleet_type   = "TASK"
		instance_type_configs = [
			{
				instance_type = "m3.xlarge"
			}
		]
		launch_specifications {
			spot_specification {
				timeout_action           = "TERMINATE_CLUSTER"
				timeout_duration_minutes = 10
			}
		}
		name                      = "emr_instance_fleet_%d"
		target_on_demand_capacity = 0
		target_spot_capacity      = 1
	}
	`, r, r, r, r, r, r, r)
}

func testAccAWSEmrInstanceFleetConfigZeroCount(r int) string {
	return fmt.Sprintf(testAccAWSEmrInstanceFleetBase+`
	resource "aws_emr_instance_fleet" "task" {
		cluster_id            = "${aws_emr_cluster.tf-test-cluster.id}"
		instance_fleet_type   = "TASK"
		instance_type_configs = [
			{
				instance_type = "m3.xlarge"
			}
		]
		launch_specifications {
			spot_specification {
				timeout_action           = "TERMINATE_CLUSTER"
				timeout_duration_minutes = 10
			}
		}
		name                      = "emr_instance_fleet_%d"
		target_on_demand_capacity = 0
		target_spot_capacity      = 0
	}
	`, r, r, r, r, r, r, r)
}

func testAccAWSEmrInstanceFleetConfigEbsBasic(r int) string {
	return fmt.Sprintf(testAccAWSEmrInstanceFleetBase+`
	resource "aws_emr_instance_fleet" "task" {
		cluster_id            = "${aws_emr_cluster.tf-test-cluster.id}"
		instance_fleet_type   = "TASK"
		instance_type_configs = [
			{
				ebs_optimized = true
				ebs_config = [
					{
						size = 10
						type = "gp2"
						volumes_per_instance = 1
					}
				]
				instance_type = "m3.xlarge"
			}
		]
		launch_specifications {
			spot_specification {
				timeout_action           = "TERMINATE_CLUSTER"
				timeout_duration_minutes = 10
			}
		}
		name                      = "emr_instance_fleet_%d"
		target_on_demand_capacity = 0
		target_spot_capacity      = 1
	}
	`, r, r, r, r, r, r, r)
}

func testAccAWSEmrInstanceFleetConfigFull(r int) string {
	return fmt.Sprintf(testAccAWSEmrInstanceFleetBase+`
	resource "aws_emr_instance_fleet" "task" {
		cluster_id            = "${aws_emr_cluster.tf-test-cluster.id}"
		instance_fleet_type   = "TASK"
		instance_type_configs = [
			{
				bid_price_as_percentage_of_on_demand_price = 100
				configurations = [
					{
						classification = "core-site"
						properties {
							"hadoop.security.groups.cache.secs" = "250"
						}
					}
				]

				ebs_optimized = true
				ebs_config = [
					{
						size = 10
						type = "gp2"
						volumes_per_instance = 1
					}
				]

				instance_type     = "m3.xlarge"
				weighted_capacity = 8
			}
		]
		launch_specifications {
			spot_specification {
				block_duration_minutes   = 60
				timeout_action           = "TERMINATE_CLUSTER"
				timeout_duration_minutes = 10
			}
		}
		name                      = "emr_instance_fleet_%d"
		target_on_demand_capacity = 1
		target_spot_capacity      = 1
	}
	`, r, r, r, r, r, r, r)
}
