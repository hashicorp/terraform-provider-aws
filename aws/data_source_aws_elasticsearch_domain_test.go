package aws

import (
	"fmt"
	"testing"

	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAWSElasticSearchDomain_basic(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceName := "aws_elasticsearch_domain.example"
	dataSourceName := "data.aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceESDomainConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &domain),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "elasticsearch_version", resourceName, "elasticsearch_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domain_id", resourceName, "domain_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint", resourceName, "endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kibana_endpoint", resourceName, "kibana_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ebs_options.0.ebs_enabled", resourceName, "ebs_options.0.ebs_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ebs_options.0.iops", resourceName, "ebs_options.0.iops"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ebs_options.0.volume_size", resourceName, "ebs_options.0.volume_size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ebs_options.0.volume_type", resourceName, "ebs_options.0.volume_type"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSElasticSearchDomain_withDedicatedMaster(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceName := "aws_elasticsearch_domain.example"
	dataSourceName := "data.aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceESDomainConfigWithDedicatedClusterMaster(ri, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &domain),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_config.0.instance_type", resourceName, "cluster_config.0.instance_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_config.0.instance_count", resourceName, "cluster_config.0.instance_count"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_config.0.dedicated_master_enabled", resourceName, "cluster_config.0.dedicated_master_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_config.0.dedicated_master_count", resourceName, "cluster_config.0.dedicated_master_count"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_config.0.dedicated_master_type", resourceName, "cluster_config.0.dedicated_master_type"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSElasticSearchDomain_withPolicy(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceName := "aws_elasticsearch_domain.example"
	dataSourceName := "data.aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceESDomainConfigWithPolicy(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &domain),
					resource.TestCheckResourceAttrPair(dataSourceName, "access_policies", resourceName, "access_policies"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSElasticSearchDomain_withEncryptAtRestDefaultKey(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceName := "aws_elasticsearch_domain.example"
	dataSourceName := "data.aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceESDomainConfigWithEncryptAtRestDefaultKey(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &domain),
					resource.TestCheckResourceAttrPair(dataSourceName, "encrypt_at_rest.0.enabled", resourceName, "encrypt_at_rest.0.enabled"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSElasticSearchDomain_withNodeToNodeEncryption(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceName := "aws_elasticsearch_domain.example"
	dataSourceName := "data.aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceESDomainConfigWithNodeToNodeEncryption(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &domain),
					resource.TestCheckResourceAttrPair(dataSourceName, "node_to_node_encryption.0.enabled", resourceName, "node_to_node_encryption.0.enabled"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSElasticSearchDomain_complex(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceName := "aws_elasticsearch_domain.example"
	dataSourceName := "data.aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceESDomainConfigComplex(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &domain),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_options.0.automated_snapshot_start_hour", resourceName, "snapshot_options.0.automated_snapshot_start_hour"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.bar", resourceName, "tags.bar"),
				),
			},
		},
	})
}

func testAccDataSourceESDomainConfig(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_elasticsearch_domain" "example" {
	domain_name = "${aws_elasticsearch_domain.example.domain_name}"
}
`, testAccESDomainConfig(randInt))
}

func testAccDataSourceESDomainConfigWithDedicatedClusterMaster(randInt int, enabled bool) string {
	return fmt.Sprintf(`
%s

data "aws_elasticsearch_domain" "example" {
	domain_name = "${aws_elasticsearch_domain.example.domain_name}"
}
`, testAccESDomainConfig_WithDedicatedClusterMaster(randInt, enabled))
}

func testAccDataSourceESDomainConfigWithPolicy(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_elasticsearch_domain" "example" {
	domain_name = "${aws_elasticsearch_domain.example.domain_name}"
}
`, testAccESDomainConfigWithPolicy(randInt, randInt))
}

func testAccDataSourceESDomainConfigWithEncryptAtRestDefaultKey(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_elasticsearch_domain" "example" {
	domain_name = "${aws_elasticsearch_domain.example.domain_name}"
}
`, testAccESDomainConfigWithEncryptAtRestDefaultKey(randInt))
}

func testAccDataSourceESDomainConfigWithNodeToNodeEncryption(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_elasticsearch_domain" "example" {
	domain_name = "${aws_elasticsearch_domain.example.domain_name}"
}
`, testAccESDomainConfigwithNodeToNodeEncryption(randInt))
}

func testAccDataSourceESDomainConfigComplex(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_elasticsearch_domain" "example" {
	domain_name = "${aws_elasticsearch_domain.example.domain_name}"
}
`, testAccESDomainConfig_complex(randInt))
}
