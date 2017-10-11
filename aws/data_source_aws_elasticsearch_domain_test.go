package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataElasticsearchDomain_basic(t *testing.T) {
	rInt := acctest.RandInt()

	fmt.Println("Inside TestAccAWSDataElasticsearchDomain_basic")

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

func testAccAWSElasticsearchDomainConfigWithDataSource(rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
}

resource "aws_elasticsearch_domain" "bar" {
  domain_name = "test-es-%d"
  elasticsearch_version = "1.5"

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

  tags {
    Domain = "TestDomain"
  }
}

data "aws_elasticsearch_domain" "bar" {
	domain_name = "${aws_elasticsearch_domain.bar.domain_name}"
}

`, rInt)
}
