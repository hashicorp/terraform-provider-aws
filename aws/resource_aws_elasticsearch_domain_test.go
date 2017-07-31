package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSElasticSearchDomain_basic(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &domain),
					resource.TestCheckResourceAttr(
						"aws_elasticsearch_domain.example", "elasticsearch_version", "1.5"),
				),
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_importBasic(t *testing.T) {
	resourceName := "aws_elasticsearch_domain.example"
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig(ri),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     resourceId,
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_v23(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfigV23(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &domain),
					resource.TestCheckResourceAttr(
						"aws_elasticsearch_domain.example", "elasticsearch_version", "2.3"),
				),
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_complex(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig_complex(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &domain),
				),
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_policy(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfigWithPolicy(acctest.RandInt(), acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &domain),
				),
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_tags(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	var td elasticsearch.ListTagsOutput
	ri := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &domain),
				),
			},

			{
				Config: testAccESDomainConfig_TagUpdate(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &domain),
					testAccLoadESTags(&domain, &td),
					testAccCheckElasticsearchServiceTags(&td.TagList, "foo", "bar"),
					testAccCheckElasticsearchServiceTags(&td.TagList, "new", "type"),
				),
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_update(t *testing.T) {
	var input elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig_ClusterUpdate(ri, 2, 22),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &input),
					testAccCheckESNumberOfInstances(2, &input),
					testAccCheckESSnapshotHour(22, &input),
				),
			},
			{
				Config: testAccESDomainConfig_ClusterUpdate(ri, 4, 23),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &input),
					testAccCheckESNumberOfInstances(4, &input),
					testAccCheckESSnapshotHour(23, &input),
				),
			},
		}})
}

func testAccCheckESSnapshotHour(snapshotHour int, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.SnapshotOptions
		if *conf.AutomatedSnapshotStartHour != int64(snapshotHour) {
			return fmt.Errorf("Snapshots start hour differ. Given: %d, Expected: %d", *conf.AutomatedSnapshotStartHour, snapshotHour)
		}
		return nil
	}
}

func testAccCheckESNumberOfInstances(numberOfInstances int, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.ElasticsearchClusterConfig
		if *conf.InstanceCount != int64(numberOfInstances) {
			return fmt.Errorf("Number of instances differ. Given: %d, Expected: %d", *conf.InstanceCount, numberOfInstances)
		}
		return nil
	}
}

func testAccLoadESTags(conf *elasticsearch.ElasticsearchDomainStatus, td *elasticsearch.ListTagsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).esconn

		describe, err := conn.ListTags(&elasticsearch.ListTagsInput{
			ARN: conf.ARN,
		})

		if err != nil {
			return err
		}
		if len(describe.TagList) > 0 {
			*td = *describe
		}
		return nil
	}
}

func testAccCheckESDomainExists(n string, domain *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ES Domain ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).esconn
		opts := &elasticsearch.DescribeElasticsearchDomainInput{
			DomainName: aws.String(rs.Primary.Attributes["domain_name"]),
		}

		resp, err := conn.DescribeElasticsearchDomain(opts)
		if err != nil {
			return fmt.Errorf("Error describing domain: %s", err.Error())
		}

		*domain = *resp.DomainStatus

		return nil
	}
}

func testAccCheckESDomainDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticsearch_domain" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).esconn
		opts := &elasticsearch.DescribeElasticsearchDomainInput{
			DomainName: aws.String(rs.Primary.Attributes["domain_name"]),
		}

		_, err := conn.DescribeElasticsearchDomain(opts)
		// Verify the error is what we want
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				continue
			}
			return err
		}
	}
	return nil
}

func testAccESDomainConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "example" {
  domain_name = "tf-test-%d"
  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, randInt)
}

func testAccESDomainConfig_ClusterUpdate(randInt, instanceInt, snapshotInt int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "example" {
  domain_name = "tf-test-%d"

  advanced_options {
    "indices.fielddata.cache.size" = 80
  }

  ebs_options {
    ebs_enabled = true
		volume_size = 10
		
  }

  cluster_config {
    instance_count = %d
    zone_awareness_enabled = true
    instance_type = "t2.micro.elasticsearch"
  }

  snapshot_options {
    automated_snapshot_start_hour = %d
  }
}
`, randInt, instanceInt, snapshotInt)
}

func testAccESDomainConfig_TagUpdate(randInt int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "example" {
  domain_name = "tf-test-%d"
  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  tags {
    foo = "bar"
    new = "type"
  }
}
`, randInt)
}

func testAccESDomainConfigWithPolicy(randESId int, randRoleId int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "example" {
  domain_name = "tf-test-%d"
   ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
  access_policies = <<CONFIG
  {
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
	"AWS": "${aws_iam_role.example_role.arn}"
      },
      "Action": "es:*",
      "Resource": "arn:aws:es:*"
    }
  ]
  }
CONFIG
}
resource "aws_iam_role" "example_role" {
  name = "es-domain-role-%d"
  assume_role_policy = "${data.aws_iam_policy_document.instance-assume-role-policy.json}"
}
data "aws_iam_policy_document" "instance-assume-role-policy" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}
`, randESId, randRoleId)
}

func testAccESDomainConfig_complex(randInt int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "example" {
  domain_name = "tf-test-%d"

  advanced_options {
    "indices.fielddata.cache.size" = 80
  }

  ebs_options {
    ebs_enabled = false
  }

  cluster_config {
    instance_count = 2
    zone_awareness_enabled = true
    instance_type = "r3.large.elasticsearch"
  }

  snapshot_options {
    automated_snapshot_start_hour = 23
  }

  tags {
    bar = "complex"
  }
}
`, randInt)
}

func testAccESDomainConfigV23(randInt int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "example" {
  domain_name = "tf-test-%d"
  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
  elasticsearch_version = "2.3"
}
`, randInt)
}
