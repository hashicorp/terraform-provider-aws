package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataElasticsearchDomain_basic(t *testing.T) {
	rInt := acctest.RandInt()
	datasourceName := "data.aws_elasticsearch_domain.test"
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticsearchDomainConfigWithDataSource(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch_version", resourceName, "elasticsearch_version"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.#", resourceName, "cluster_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.0.instance_type", resourceName, "cluster_config.0.instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.0.instance_count", resourceName, "cluster_config.0.instance_count"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.0.dedicated_master_enabled", resourceName, "cluster_config.0.dedicated_master_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.0.zone_awareness_enabled", resourceName, "cluster_config.0.zone_awareness_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_options.#", resourceName, "ebs_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_options.0.ebs_enabled", resourceName, "ebs_options.0.ebs_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_options.0.volume_type", resourceName, "ebs_options.0.volume_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_options.0.volume_size", resourceName, "ebs_options.0.volume_size"),
					resource.TestCheckResourceAttrPair(datasourceName, "snapshot_options.#", resourceName, "snapshot_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "snapshot_options.0.automated_snapshot_start_hour", resourceName, "snapshot_options.0.automated_snapshot_start_hour"),
				),
			},
		},
	})
}

func TestAccAWSDataElasticsearchDomain_advanced(t *testing.T) {
	rInt := acctest.RandInt()
	datasourceName := "data.aws_elasticsearch_domain.test"
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticsearchDomainConfigAdvancedWithDataSource(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "elasticsearch_version", resourceName, "elasticsearch_version"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.#", resourceName, "cluster_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.0.instance_type", resourceName, "cluster_config.0.instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.0.instance_count", resourceName, "cluster_config.0.instance_count"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.0.dedicated_master_enabled", resourceName, "cluster_config.0.dedicated_master_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "cluster_config.0.zone_awareness_enabled", resourceName, "cluster_config.0.zone_awareness_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_options.#", resourceName, "ebs_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_options.0.ebs_enabled", resourceName, "ebs_options.0.ebs_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_options.0.volume_type", resourceName, "ebs_options.0.volume_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_options.0.volume_size", resourceName, "ebs_options.0.volume_size"),
					resource.TestCheckResourceAttrPair(datasourceName, "snapshot_options.#", resourceName, "snapshot_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "snapshot_options.0.automated_snapshot_start_hour", resourceName, "snapshot_options.0.automated_snapshot_start_hour"),
					resource.TestCheckResourceAttrPair(datasourceName, "log_publishing_options.#", resourceName, "log_publishing_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_options.#", resourceName, "vpc_options.#"),
				),
			},
		},
	})
}

func testAccAWSElasticsearchDomainConfigWithDataSource(rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
}

locals {
	random_name = "test-es-%d"
}
		
data "aws_region" "current" {}
		
data "aws_caller_identity" "current" {}
					
resource "aws_elasticsearch_domain" "test" {
	domain_name = "${local.random_name}"
	elasticsearch_version = "1.5"
	
	access_policies = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "es:*",
			"Principal": "*",
			"Effect": "Allow",
			"Resource": "arn:aws:es:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:domain/${local.random_name}/*",
			"Condition": {
				"IpAddress": {"aws:SourceIp": ["127.0.0.0/8"]}
			}
		}
	]
}
POLICY

  cluster_config {
    instance_type = "t2.micro.elasticsearch"
    instance_count = 2
	dedicated_master_enabled = false
	zone_awareness_config {
		availability_zone_count = 2
	}
    zone_awareness_enabled = true
  }
  ebs_options {
	ebs_enabled = true
	volume_type = "gp2"
	volume_size = 20
  }
  snapshot_options {
    automated_snapshot_start_hour = 23
  }
}

data "aws_elasticsearch_domain" "test" {
  domain_name = "${aws_elasticsearch_domain.test.domain_name}"
}
		`, rInt)
}

func testAccAWSElasticsearchDomainConfigAdvancedWithDataSource(rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
}
		
data "aws_region" "current" {}
		
data "aws_caller_identity" "current" {}

locals {
	random_name = "test-es-%d"
}

resource "aws_cloudwatch_log_group" "test" {
	name = "${local.random_name}"
}

resource "aws_cloudwatch_log_resource_policy" "test" {
	policy_name = "${local.random_name}"
	policy_document = <<CONFIG
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Principal": {
				"Service": "es.amazonaws.com"
			},
			"Action": [
				"logs:PutLogEvents",
				"logs:PutLogEventsBatch",
				"logs:CreateLogStream"
			],
			"Resource": "arn:aws:logs:*"
		}
	]
}
CONFIG
}

resource "aws_vpc" "test" {
	cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.0.0.0/24"
}

resource "aws_subnet" "test2" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.0.1.0/24"
}

resource "aws_security_group" "test" {
	name = "${local.random_name}"
	vpc_id = "${aws_vpc.test.id}"
}

resource "aws_security_group_rule" "test" {
	type = "ingress"
	from_port = 443
	to_port = 443
	protocol = "tcp"
	cidr_blocks = [ "0.0.0.0/0" ]

	security_group_id = "${aws_security_group.test.id}"
}

resource "aws_iam_service_linked_role" "test" {
	aws_service_name = "es.amazonaws.com"
}

resource "aws_elasticsearch_domain" "test" {
	domain_name = "${local.random_name}"
	elasticsearch_version = "1.5"
	
	access_policies = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "es:*",
			"Principal": "*",
			"Effect": "Allow",
			"Resource": "arn:aws:es:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:domain/${local.random_name}/*"
		}
	]
}
POLICY

  cluster_config {
    instance_type = "t2.micro.elasticsearch"
    instance_count = 2
    dedicated_master_enabled = false
	zone_awareness_config {
		availability_zone_count = 2
	}
    zone_awareness_enabled = true
  }
  ebs_options {
	ebs_enabled = true
	volume_type = "gp2"
	volume_size = 20
  }
  snapshot_options {
    automated_snapshot_start_hour = 23
  }
  log_publishing_options {
    cloudwatch_log_group_arn = "${aws_cloudwatch_log_group.test.arn}"
    log_type = "INDEX_SLOW_LOGS"
  }
  vpc_options {
		security_group_ids = [
			"${aws_security_group.test.id}"
		]
		subnet_ids = [
			"${aws_subnet.test.id}",
			"${aws_subnet.test2.id}"
		]
  }

  tags = {
	Domain = "TestDomain"
  }
	
  depends_on = [
    "aws_iam_service_linked_role.test",
  ]
}

data "aws_elasticsearch_domain" "test" {
  domain_name = "${aws_elasticsearch_domain.test.domain_name}"
}
				`, rInt)
}
