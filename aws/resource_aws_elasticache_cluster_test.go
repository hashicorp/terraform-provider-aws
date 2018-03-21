package aws

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSElasticacheCluster_basic(t *testing.T) {
	var ec elasticache.CacheCluster
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSecurityGroupExists("aws_elasticache_security_group.bar"),
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.bar", &ec),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttrSet("aws_elasticache_cluster.bar", "configuration_endpoint"),
					resource.TestCheckResourceAttrSet("aws_elasticache_cluster.bar", "cluster_address"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_snapshotsWithUpdates(t *testing.T) {
	var ec elasticache.CacheCluster

	ri := acctest.RandInt()
	preConfig := fmt.Sprintf(testAccAWSElasticacheClusterConfig_snapshots, ri, ri, acctest.RandString(10))
	postConfig := fmt.Sprintf(testAccAWSElasticacheClusterConfig_snapshotsUpdated, ri, ri, acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSecurityGroupExists("aws_elasticache_security_group.bar"),
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.bar", &ec),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "snapshot_window", "05:00-09:00"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "snapshot_retention_limit", "3"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSecurityGroupExists("aws_elasticache_security_group.bar"),
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.bar", &ec),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "snapshot_window", "07:00-09:00"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "snapshot_retention_limit", "7"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_decreasingCacheNodes(t *testing.T) {
	var ec elasticache.CacheCluster

	ri := acctest.RandInt()
	preConfig := fmt.Sprintf(testAccAWSElasticacheClusterConfigDecreasingNodes, ri, ri, acctest.RandString(10))
	postConfig := fmt.Sprintf(testAccAWSElasticacheClusterConfigDecreasingNodes_update, ri, ri, acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSecurityGroupExists("aws_elasticache_security_group.bar"),
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.bar", &ec),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "num_cache_nodes", "3"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSecurityGroupExists("aws_elasticache_security_group.bar"),
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.bar", &ec),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "num_cache_nodes", "1"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_vpc(t *testing.T) {
	var csg elasticache.CacheSubnetGroup
	var ec elasticache.CacheCluster
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterInVPCConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSubnetGroupExists("aws_elasticache_subnet_group.bar", &csg),
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.bar", &ec),
					testAccCheckAWSElasticacheClusterAttributes(&ec),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "availability_zone", "us-west-2a"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_multiAZInVpc(t *testing.T) {
	var csg elasticache.CacheSubnetGroup
	var ec elasticache.CacheCluster
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterMultiAZInVPCConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSubnetGroupExists("aws_elasticache_subnet_group.bar", &csg),
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.bar", &ec),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "availability_zone", "Multiple"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_AZMode_Memcached_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var cluster elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheClusterConfig_AZMode_Memcached_Ec2Classic(rName, "unknown"),
				ExpectError: regexp.MustCompile(`expected az_mode to be one of .*, got unknown`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_AZMode_Memcached_Ec2Classic(rName, "cross-az"),
				ExpectError: regexp.MustCompile(`az_mode "cross-az" is not supported with num_cache_nodes = 1`),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_AZMode_Memcached_Ec2Classic(rName, "single-az"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "az_mode", "single-az"),
				),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_AZMode_Memcached_Ec2Classic(rName, "cross-az"),
				ExpectError: regexp.MustCompile(`az_mode "cross-az" is not supported with num_cache_nodes = 1`),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_AZMode_Redis_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var cluster elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheClusterConfig_AZMode_Redis_Ec2Classic(rName, "unknown"),
				ExpectError: regexp.MustCompile(`expected az_mode to be one of .*, got unknown`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_AZMode_Redis_Ec2Classic(rName, "cross-az"),
				ExpectError: regexp.MustCompile(`az_mode "cross-az" is not supported with num_cache_nodes = 1`),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_AZMode_Redis_Ec2Classic(rName, "single-az"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "az_mode", "single-az"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_EngineVersion_Memcached_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var pre, mid, post elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Memcached_Ec2Classic(rName, "1.4.33"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.4.33"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Memcached_Ec2Classic(rName, "1.4.24"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &mid),
					testAccCheckAWSElasticacheClusterRecreated(&pre, &mid),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.4.24"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Memcached_Ec2Classic(rName, "1.4.34"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &post),
					testAccCheckAWSElasticacheClusterNotRecreated(&mid, &post),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.4.34"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_EngineVersion_Redis_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var pre, mid, post elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Redis_Ec2Classic(rName, "3.2.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.2.6"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Redis_Ec2Classic(rName, "3.2.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &mid),
					testAccCheckAWSElasticacheClusterRecreated(&pre, &mid),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.2.4"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Redis_Ec2Classic(rName, "3.2.10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &post),
					testAccCheckAWSElasticacheClusterNotRecreated(&mid, &post),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.2.10"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_NodeTypeResize_Memcached_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var pre, post elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheClusterConfig_NodeType_Memcached_Ec2Classic(rName, "cache.t2.micro"),
				ExpectError: regexp.MustCompile(`node_type "cache.t2.micro" can only be created in a VPC`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_NodeType_Memcached_Ec2Classic(rName, "cache.t2.small"),
				ExpectError: regexp.MustCompile(`node_type "cache.t2.small" can only be created in a VPC`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_NodeType_Memcached_Ec2Classic(rName, "cache.t2.medium"),
				ExpectError: regexp.MustCompile(`node_type "cache.t2.medium" can only be created in a VPC`),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_NodeType_Memcached_Ec2Classic(rName, "cache.m3.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.m3.medium"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_NodeType_Memcached_Ec2Classic(rName, "cache.m3.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &post),
					testAccCheckAWSElasticacheClusterRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.m3.large"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_NodeTypeResize_Redis_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var pre, post elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheClusterConfig_NodeType_Redis_Ec2Classic(rName, "cache.t2.micro"),
				ExpectError: regexp.MustCompile(`node_type "cache.t2.micro" can only be created in a VPC`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_NodeType_Redis_Ec2Classic(rName, "cache.t2.small"),
				ExpectError: regexp.MustCompile(`node_type "cache.t2.small" can only be created in a VPC`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_NodeType_Redis_Ec2Classic(rName, "cache.t2.medium"),
				ExpectError: regexp.MustCompile(`node_type "cache.t2.medium" can only be created in a VPC`),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_NodeType_Redis_Ec2Classic(rName, "cache.m3.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.m3.medium"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_NodeType_Redis_Ec2Classic(rName, "cache.m3.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &post),
					testAccCheckAWSElasticacheClusterNotRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.m3.large"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_NumCacheNodes_Redis_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheClusterConfig_NumCacheNodes_Redis_Ec2Classic(rName, 2),
				ExpectError: regexp.MustCompile(`engine "redis" does not support num_cache_nodes > 1`),
			},
		},
	})
}

func testAccCheckAWSElasticacheClusterAttributes(v *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v.NotificationConfiguration == nil {
			return fmt.Errorf("Expected NotificationConfiguration for ElastiCache Cluster (%s)", *v.CacheClusterId)
		}

		if strings.ToLower(*v.NotificationConfiguration.TopicStatus) != "active" {
			return fmt.Errorf("Expected NotificationConfiguration status to be 'active', got (%s)", *v.NotificationConfiguration.TopicStatus)
		}

		return nil
	}
}

func testAccCheckAWSElasticacheClusterNotRecreated(i, j *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CacheClusterCreateTime) != aws.TimeValue(j.CacheClusterCreateTime) {
			return errors.New("Elasticache Cluster was recreated")
		}

		return nil
	}
}

func testAccCheckAWSElasticacheClusterRecreated(i, j *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CacheClusterCreateTime) == aws.TimeValue(j.CacheClusterCreateTime) {
			return errors.New("Elasticache Cluster was not recreated")
		}

		return nil
	}
}

func testAccCheckAWSElasticacheClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_cluster" {
			continue
		}
		res, err := conn.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{
			CacheClusterId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			// Verify the error is what we want
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "CacheClusterNotFound" {
				continue
			}
			return err
		}
		if len(res.CacheClusters) > 0 {
			return fmt.Errorf("still exist.")
		}
	}
	return nil
}

func testAccCheckAWSElasticacheClusterExists(n string, v *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No cache cluster ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elasticacheconn
		resp, err := conn.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{
			CacheClusterId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("Elasticache error: %v", err)
		}

		for _, c := range resp.CacheClusters {
			if *c.CacheClusterId == rs.Primary.ID {
				*v = *c
			}
		}

		return nil
	}
}

func testAccAWSElasticacheClusterConfigBasic(clusterId string) string {
	return fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
}

resource "aws_elasticache_cluster" "bar" {
    cluster_id = "tf-%s"
    engine = "memcached"
    node_type = "cache.m1.small"
    num_cache_nodes = 1
    port = 11211
    parameter_group_name = "default.memcached1.4"
}
`, clusterId)
}

var testAccAWSElasticacheClusterConfig = fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
}
resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }

		tags {
			Name = "TestAccAWSElasticacheCluster_basic"
		}
}

resource "aws_elasticache_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    security_group_names = ["${aws_security_group.bar.name}"]
}

resource "aws_elasticache_cluster" "bar" {
    cluster_id = "tf-%s"
    engine = "memcached"
    node_type = "cache.m1.small"
    num_cache_nodes = 1
    port = 11211
    parameter_group_name = "default.memcached1.4"
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
}
`, acctest.RandInt(), acctest.RandInt(), acctest.RandString(10))

var testAccAWSElasticacheClusterConfig_snapshots = `
provider "aws" {
	region = "us-east-1"
}
resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    security_group_names = ["${aws_security_group.bar.name}"]
}

resource "aws_elasticache_cluster" "bar" {
    cluster_id = "tf-%s"
    engine = "redis"
    node_type = "cache.m1.small"
    num_cache_nodes = 1
    port = 6379
  	parameter_group_name = "default.redis3.2"
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
    snapshot_window = "05:00-09:00"
    snapshot_retention_limit = 3
}
`

var testAccAWSElasticacheClusterConfig_snapshotsUpdated = `
provider "aws" {
	region = "us-east-1"
}
resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    security_group_names = ["${aws_security_group.bar.name}"]
}

resource "aws_elasticache_cluster" "bar" {
    cluster_id = "tf-%s"
    engine = "redis"
    node_type = "cache.m1.small"
    num_cache_nodes = 1
    port = 6379
  	parameter_group_name = "default.redis3.2"
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
    snapshot_window = "07:00-09:00"
    snapshot_retention_limit = 7
    apply_immediately = true
}
`

var testAccAWSElasticacheClusterConfigDecreasingNodes = `
provider "aws" {
	region = "us-east-1"
}
resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    security_group_names = ["${aws_security_group.bar.name}"]
}

resource "aws_elasticache_cluster" "bar" {
    cluster_id = "tf-%s"
    engine = "memcached"
    node_type = "cache.m1.small"
    num_cache_nodes = 3
    port = 11211
    parameter_group_name = "default.memcached1.4"
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
}
`

var testAccAWSElasticacheClusterConfigDecreasingNodes_update = `
provider "aws" {
	region = "us-east-1"
}
resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    security_group_names = ["${aws_security_group.bar.name}"]
}

resource "aws_elasticache_cluster" "bar" {
    cluster_id = "tf-%s"
    engine = "memcached"
    node_type = "cache.m1.small"
    num_cache_nodes = 1
    port = 11211
    parameter_group_name = "default.memcached1.4"
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
    apply_immediately = true
}
`

var testAccAWSElasticacheClusterInVPCConfig = fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "192.168.0.0/16"
    tags {
        Name = "terraform-testacc-elasticache-cluster-in-vpc"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
    tags {
        Name = "tf-acc-elasticache-cluster-in-vpc"
    }
}

resource "aws_elasticache_subnet_group" "bar" {
    name = "tf-test-cache-subnet-%03d"
    description = "tf-test-cache-subnet-group-descr"
    subnet_ids = ["${aws_subnet.foo.id}"]
}

resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    vpc_id = "${aws_vpc.foo.id}"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_cluster" "bar" {
    // Including uppercase letters in this name to ensure
    // that we correctly handle the fact that the API
    // normalizes names to lowercase.
    cluster_id = "tf-%s"
    node_type = "cache.m1.small"
    num_cache_nodes = 1
    engine = "redis"
    engine_version = "2.8.19"
    port = 6379
    subnet_group_name = "${aws_elasticache_subnet_group.bar.name}"
    security_group_ids = ["${aws_security_group.bar.id}"]
    parameter_group_name = "default.redis2.8"
    notification_topic_arn      = "${aws_sns_topic.topic_example.arn}"
    availability_zone = "us-west-2a"
}

resource "aws_sns_topic" "topic_example" {
  name = "tf-ecache-cluster-test"
}
`, acctest.RandInt(), acctest.RandInt(), acctest.RandString(10))

var testAccAWSElasticacheClusterMultiAZInVPCConfig = fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "192.168.0.0/16"
    tags {
        Name = "terraform-testacc-elasticache-cluster-multi-az-in-vpc"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
    tags {
        Name = "tf-acc-elasticache-cluster-multi-az-in-vpc-foo"
    }
}

resource "aws_subnet" "bar" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.16.0/20"
    availability_zone = "us-west-2b"
    tags {
        Name = "tf-acc-elasticache-cluster-multi-az-in-vpc-bar"
    }
}

resource "aws_elasticache_subnet_group" "bar" {
    name = "tf-test-cache-subnet-%03d"
    description = "tf-test-cache-subnet-group-descr"
    subnet_ids = [
        "${aws_subnet.foo.id}",
        "${aws_subnet.bar.id}"
    ]
}

resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    vpc_id = "${aws_vpc.foo.id}"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_cluster" "bar" {
    cluster_id = "tf-%s"
    engine = "memcached"
    node_type = "cache.m1.small"
    num_cache_nodes = 2
    port = 11211
    subnet_group_name = "${aws_elasticache_subnet_group.bar.name}"
    security_group_ids = ["${aws_security_group.bar.id}"]
    parameter_group_name = "default.memcached1.4"
    az_mode = "cross-az"
    availability_zones = [
        "us-west-2a",
        "us-west-2b"
    ]
}
`, acctest.RandInt(), acctest.RandInt(), acctest.RandString(10))

func testAccAWSElasticacheClusterConfig_AZMode_Memcached_Ec2Classic(rName, azMode string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  apply_immediately    = true
  az_mode              = "%[2]s"
  cluster_id           = "%[1]s"
  engine               = "memcached"
  node_type            = "cache.m3.medium"
  num_cache_nodes      = 1
  parameter_group_name = "default.memcached1.4"
  port                 = 11211
}
`, rName, azMode)
}

func testAccAWSElasticacheClusterConfig_AZMode_Redis_Ec2Classic(rName, azMode string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  apply_immediately    = true
  az_mode              = "%[2]s"
  cluster_id           = "%[1]s"
  engine               = "redis"
  node_type            = "cache.m3.medium"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis3.2"
  port                 = 6379
}
`, rName, azMode)
}

func testAccAWSElasticacheClusterConfig_EngineVersion_Memcached_Ec2Classic(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  apply_immediately    = true
  cluster_id           = "%[1]s"
  engine               = "memcached"
  engine_version       = "%[2]s"
  node_type            = "cache.m3.medium"
  num_cache_nodes      = 1
  parameter_group_name = "default.memcached1.4"
  port                 = 11211
}
`, rName, engineVersion)
}

func testAccAWSElasticacheClusterConfig_EngineVersion_Redis_Ec2Classic(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  apply_immediately    = true
  cluster_id           = "%[1]s"
  engine               = "redis"
  engine_version       = "%[2]s"
  node_type            = "cache.m3.medium"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis3.2"
  port                 = 6379
}
`, rName, engineVersion)
}

func testAccAWSElasticacheClusterConfig_NodeType_Memcached_Ec2Classic(rName, nodeType string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  apply_immediately    = true
  cluster_id           = "%[1]s"
  engine               = "memcached"
  node_type            = "%[2]s"
  num_cache_nodes      = 1
  parameter_group_name = "default.memcached1.4"
  port                 = 11211
}
`, rName, nodeType)
}

func testAccAWSElasticacheClusterConfig_NodeType_Redis_Ec2Classic(rName, nodeType string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  apply_immediately    = true
  cluster_id           = "%[1]s"
  engine               = "redis"
  node_type            = "%[2]s"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis3.2"
  port                 = 6379
}
`, rName, nodeType)
}

func testAccAWSElasticacheClusterConfig_NumCacheNodes_Redis_Ec2Classic(rName string, numCacheNodes int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  apply_immediately    = true
  cluster_id           = "%[1]s"
  engine               = "redis"
  node_type            = "cache.m3.medium"
  num_cache_nodes      = %[2]d
  parameter_group_name = "default.redis3.2"
  port                 = 6379
}
`, rName, numCacheNodes)
}
