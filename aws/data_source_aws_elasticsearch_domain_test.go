package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataElasticsearchDomain_basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticsearchDomainConfigWithDataSource(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "elasticsearch_version", "1.5"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "cluster_config.#", "1"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "cluster_config.0.instance_type", "t2.micro.elasticsearch"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "cluster_config.0.instance_count", "2"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "cluster_config.0.dedicated_master_enabled", "false"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "cluster_config.0.zone_awareness_enabled", "true"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "ebs_options.#", "1"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "ebs_options.0.ebs_enabled", "true"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "ebs_options.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "ebs_options.0.volume_size", "20"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "snapshot_options.#", "1"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "snapshot_options.0.automated_snapshot_start_hour", "23"),
				),
			},
		},
	})
}

func TestAccAWSDataElasticsearchDomain_advanced(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticsearchDomainConfigAdvancedWithDataSource(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "elasticsearch_version", "1.5"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "cluster_config.#", "1"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "cluster_config.0.instance_type", "t2.micro.elasticsearch"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "cluster_config.0.instance_count", "2"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "cluster_config.0.dedicated_master_enabled", "false"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "cluster_config.0.zone_awareness_enabled", "true"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "ebs_options.#", "1"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "ebs_options.0.ebs_enabled", "true"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "ebs_options.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "ebs_options.0.volume_size", "20"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "snapshot_options.#", "1"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "snapshot_options.0.automated_snapshot_start_hour", "23"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "log_publishing_options.#", "1"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "log_publishing_options.0.log_type", "INDEX_SLOW_LOGS"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "vpc_options.#", "1"),
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
					
resource "aws_elasticsearch_domain" "bar" {
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
				"IpAddress": {"aws:SourceIp": ["66.193.100.22/32"]}
			}
		}
	]
}
POLICY

  cluster_config {
    instance_type = "t2.micro.elasticsearch"
    instance_count = 2
    dedicated_master_enabled = false
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

data "aws_elasticsearch_domain" "bar" {
  domain_name = "${aws_elasticsearch_domain.bar.domain_name}"
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

resource "aws_cloudwatch_log_group" "bar" {
	name = "${local.random_name}"
}

resource "aws_cloudwatch_log_resource_policy" "bar" {
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

resource "aws_vpc" "bar" {
	cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "bar" {
	vpc_id = "${aws_vpc.bar.id}"
	cidr_block = "10.0.0.0/24"
}

resource "aws_subnet" "baz" {
	vpc_id = "${aws_vpc.bar.id}"
	cidr_block = "10.0.1.0/24"
}

resource "aws_security_group" "bar" {
	name = "${local.random_name}"
	vpc_id = "${aws_vpc.bar.id}"
}

resource "aws_security_group_rule" "bar" {
	type = "ingress"
	from_port = 443
	to_port = 443
	protocol = "tcp"
	cidr_blocks = [ "0.0.0.0/0" ]

	security_group_id = "${aws_security_group.bar.id}"
}

resource "aws_iam_service_linked_role" "bar" {
	aws_service_name = "es.amazonaws.com"
}

resource "aws_elasticsearch_domain" "bar" {
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
    cloudwatch_log_group_arn = "${aws_cloudwatch_log_group.bar.arn}"
    log_type = "INDEX_SLOW_LOGS"
  }
  vpc_options {
		security_group_ids = [
			"${aws_security_group.bar.id}"
		]
		subnet_ids = [
			"${aws_subnet.bar.id}",
			"${aws_subnet.baz.id}"
		]
  }

  tags {
	  Domain = "TestDomain"
	}
	
	depends_on = [
    "aws_iam_service_linked_role.bar",
  ]
}

data "aws_elasticsearch_domain" "bar" {
  domain_name = "${aws_elasticsearch_domain.bar.domain_name}"
}
				`, rInt)
}
