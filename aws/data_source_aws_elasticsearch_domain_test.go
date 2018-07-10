package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataElasticsearchDomain_basic(t *testing.T) {
	rInt := acctest.RandInt()
	rString := acctest.RandString(10)

	domainName := fmt.Sprintf("tf-%d", rInt)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticsearchDomainConfigWithDataSource(rString, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("data.aws_elasticsearch_domain.bar", "arn",
						regexp.MustCompile(`arn:aws:es:[^:]+:[^:]+:domain/`+domainName)),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "domain_name", domainName),

					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "tags.%", "1"),
					resource.TestCheckResourceAttr("data.aws_elasticsearch_domain.bar", "tags.Domain", "somedomain"),

					resource.TestCheckResourceAttrSet("data.aws_elasticsearch_domain.bar", "endpoint"),
					resource.TestCheckResourceAttrSet("data.aws_elasticsearch_domain.bar", "domain_id"),
				),
			},
		},
	})
}

func testAccAWSElasticsearchDomainConfigWithDataSource(rString string, rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
}

resource "aws_elasticsearch_domain" "es" {
  domain_name               = "tf-%d"
  elasticsearch_version     = "5.5"
  cluster_config {
    instance_type           = "i3.large.elasticsearch"
  }
  tags { 
    Domain = "somedomain"
  }
}

data "aws_elasticsearch_domain" "bar" {
	domain_name = "${aws_elasticsearch_domain.es.domain_name}"
}
`, rInt)
}
