package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_msk_cluster", &resource.Sweeper{
		Name: "aws_msk_cluster",
		F:    testSweepMskClusters,
	})
}

func testSweepMskClusters(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).kafkaconn

	out, err := conn.ListClusters(&kafka.ListClustersInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] skipping msk cluster domain sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving MSK clusters: %s", err)
	}

	for _, cluster := range out.ClusterInfoList {
		log.Printf("[INFO] Deleting Msk cluster: %s", *cluster.ClusterName)
		_, err := conn.DeleteCluster(&kafka.DeleteClusterInput{
			ClusterArn: cluster.ClusterArn,
		})
		if err != nil {
			log.Printf("[ERROR] Failed to delete MSK cluster %s: %s", *cluster.ClusterName, err)
			continue
		}
		err = resourceAwsMskClusterDeleteWaiter(conn, *cluster.ClusterArn)
		if err != nil {
			log.Printf("[ERROR] failed to wait for deletion of MSK cluster %s: %s", *cluster.ClusterName, err)
		}
	}
	return nil
}

func TestAccAWSMskCluster_basic(t *testing.T) {
	var cluster kafka.ClusterInfo
	var td kafka.ListTagsForResourceOutput
	ri := acctest.RandInt()
	resourceName := "aws_msk_cluster.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfig_basic(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "kafka", regexp.MustCompile(`cluster/.+`)),
					testAccMatchResourceAttrRegionalARN(resourceName, "encryption_info.0.encryption_at_rest_kms_key_arn", "kms", regexp.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cluster_name", fmt.Sprintf("tf-test-%d", ri)),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.az_distribution", kafka.BrokerAZDistributionDefault),
					resource.TestCheckResourceAttr(resourceName, "kafka_version", "1.1.1"),
					resource.TestCheckResourceAttr(resourceName, "number_of_broker_nodes", "3"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.instance_type", "kafka.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.ebs_volume_size", "10"),
					resource.TestMatchResourceAttr(resourceName, "zookeeper_connect_string", regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+:\d+,\d+\.\d+\.\d+\.\d+:\d+,\d+\.\d+\.\d+\.\d+:\d+$`)),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers", regexp.MustCompile(`^(([-\w]+\.){1,}[\w]+:\d+,){2,}([-\w]+\.){1,}[\w]+:\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "enhanced_monitoring", kafka.EnhancedMonitoringDefault),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.client_subnets.#", "3"),
					resource.TestCheckResourceAttrPair(resourceName, "broker_node_group_info.0.client_subnets.0", "aws_subnet.example_subnet_az1", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "broker_node_group_info.0.client_subnets.1", "aws_subnet.example_subnet_az2", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "broker_node_group_info.0.client_subnets.2", "aws_subnet.example_subnet_az3", "id"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.security_groups.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "broker_node_group_info.0.security_groups.0", "aws_security_group.example_sg", "id"),
					testAccLoadMskTags(&cluster, &td),
					testAccCheckMskClusterTags(&td, "foo", "bar"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bootstrap_brokers", // API may mutate ordering and selection of brokers to return
				},
			},
		},
	})
}
func TestAccAWSMskCluster_kms(t *testing.T) {
	var cluster kafka.ClusterInfo
	ri := acctest.RandInt()
	resourceName := "aws_msk_cluster.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfig_kms(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_info.0.encryption_at_rest_kms_key_arn", "aws_kms_key.example_key", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bootstrap_brokers", // API may mutate ordering and selection of brokers to return
				},
			},
		},
	})
}

func TestAccAWSMskCluster_enhancedMonitoring(t *testing.T) {
	var cluster kafka.ClusterInfo
	ri := acctest.RandInt()
	resourceName := "aws_msk_cluster.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfig_enhancedMonitoring(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "enhanced_monitoring", kafka.EnhancedMonitoringPerBroker),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bootstrap_brokers",
				},
			},
		},
	})
}
func TestAccAWSMskCluster_tagsUpdate(t *testing.T) {
	var cluster kafka.ClusterInfo
	var td kafka.ListTagsForResourceOutput
	ri := acctest.RandInt()
	resourceName := "aws_msk_cluster.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfig_basic(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster),
					testAccLoadMskTags(&cluster, &td),
					testAccCheckMskClusterTags(&td, "foo", "bar"),
				),
			},
			{
				Config: testAccMskClusterConfig_tagsUpdate(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster),
					testAccLoadMskTags(&cluster, &td),
					testAccCheckMskClusterTags(&td, "foo", "baz"),
					testAccCheckMskClusterTags(&td, "new", "type"),
				),
			},
		},
	})
}

func TestAccAWSMskCluster_brokerNodes(t *testing.T) {
	var cluster kafka.ClusterInfo
	ri := acctest.RandInt()
	resourceName := "aws_msk_cluster.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfig_brokerNodes(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "number_of_broker_nodes", "6"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.instance_type", "kafka.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.ebs_volume_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.client_subnets.#", "3"),
					resource.TestCheckResourceAttrPair(resourceName, "broker_node_group_info.0.client_subnets.0", "aws_subnet.example_subnet_az1", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "broker_node_group_info.0.client_subnets.1", "aws_subnet.example_subnet_az2", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "broker_node_group_info.0.client_subnets.2", "aws_subnet.example_subnet_az3", "id"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.security_groups.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "broker_node_group_info.0.security_groups.0", "aws_security_group.example_sg", "id"),
					resource.TestCheckResourceAttr(resourceName, "kafka_version", "2.1.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bootstrap_brokers", // API may mutate ordering and selection of brokers to return
				},
			},
		},
	})
}

func testAccCheckMskClusterDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_msk_cluster" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).kafkaconn
		opts := &kafka.DescribeClusterInput{
			ClusterArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeCluster(opts)
		if err != nil {
			if isAWSErr(err, kafka.ErrCodeNotFoundException, "") {
				continue
			}
			return err
		}
	}
	return nil
}

func testAccCheckMskClusterExists(n string, cluster *kafka.ClusterInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cluster arn is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kafkaconn
		resp, err := conn.DescribeCluster(&kafka.DescribeClusterInput{
			ClusterArn: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("Error describing cluster: %s", err.Error())
		}

		*cluster = *resp.ClusterInfo
		return nil
	}
}

func testAccLoadMskTags(cluster *kafka.ClusterInfo, td *kafka.ListTagsForResourceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).kafkaconn

		tagOut, err := conn.ListTagsForResource(&kafka.ListTagsForResourceInput{
			ResourceArn: cluster.ClusterArn,
		})
		if err != nil {
			return err
		}
		if tagOut != nil {
			*td = *tagOut
			log.Printf("[DEBUG] loaded acceptance test tags: %v (from %v)", td, tagOut)
		}
		return nil
	}
}

func testAccCheckMskClusterTags(td *kafka.ListTagsForResourceOutput, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		m := tagsToMapMskCluster(td.Tags)
		v, ok := m[key]
		if value != "" && !ok {
			return fmt.Errorf("Missing tag: %s - (found tags %v)", key, m)
		} else if value == "" && ok {
			return fmt.Errorf("Extra tag: %s", key)
		}
		if value == "" {
			return nil
		}
		if v != value {
			return fmt.Errorf("%s: bad value: %s", key, v)
		}
		return nil
	}
}

func testAccMskClusterBaseConfig() string {
	return fmt.Sprintf(`
resource "aws_vpc" "example_vpc" {
	cidr_block = "192.168.0.0/22"
	tags = {
		Name = "tf-testacc-msk-cluster-vpc"
	}
}

data "aws_availability_zones" "available" {
	state = "available"
}

resource "aws_subnet" "example_subnet_az1" {
	vpc_id = "${aws_vpc.example_vpc.id}"
	cidr_block = "192.168.0.0/24"
	availability_zone = "${data.aws_availability_zones.available.names[0]}"
	tags = {
		Name = "tf-testacc-msk-cluster-subnet-az1"
	}
}
resource "aws_subnet" "example_subnet_az2" {
	vpc_id = "${aws_vpc.example_vpc.id}"
	cidr_block = "192.168.1.0/24"
	availability_zone = "${data.aws_availability_zones.available.names[1]}"
	tags = {
		Name = "tf-testacc-msk-cluster-subnet-az2"
	}
}
resource "aws_subnet" "example_subnet_az3" {
	vpc_id = "${aws_vpc.example_vpc.id}"
	cidr_block = "192.168.2.0/24"
	availability_zone = "${data.aws_availability_zones.available.names[2]}"
	tags = {
		Name = "tf-testacc-msk-cluster-subnet-az3"
	}
}

resource "aws_security_group" "example_sg" {
	vpc_id = "${aws_vpc.example_vpc.id}"
}
`)

}
func testAccMskClusterConfig_basic(randInt int) string {
	return testAccMskClusterBaseConfig() + fmt.Sprintf(`

resource "aws_msk_cluster" "example" {
	cluster_name = "tf-test-%d"
	kafka_version = "1.1.1"
	number_of_broker_nodes = 3
	broker_node_group_info {
		instance_type = "kafka.m5.large"
		ebs_volume_size = 10
		client_subnets = [ "${aws_subnet.example_subnet_az1.id}", "${aws_subnet.example_subnet_az2.id}", "${aws_subnet.example_subnet_az3.id}" ] 
		security_groups = [ "${aws_security_group.example_sg.id}" ]
	}
	tags = {
		foo = "bar"
	}
}
`, randInt)
}

func testAccMskClusterConfig_kms(randInt int) string {
	return testAccMskClusterBaseConfig() + fmt.Sprintf(`

resource "aws_kms_key" "example_key" {
	description = "tf-testacc-msk-cluster-kms"
	tags = {
		Name = "tf-testacc-msk-cluster-kms"
	}
}

resource "aws_msk_cluster" "example" {
	cluster_name = "tf-test-%d"
	kafka_version = "1.1.1"
	number_of_broker_nodes = 3
	encryption_info {
		encryption_at_rest_kms_key_arn = "${aws_kms_key.example_key.arn}"
	}
	broker_node_group_info {
		instance_type = "kafka.m5.large"
		ebs_volume_size = 10
		client_subnets = [ "${aws_subnet.example_subnet_az1.id}", "${aws_subnet.example_subnet_az2.id}", "${aws_subnet.example_subnet_az3.id}" ] 
		security_groups = [ "${aws_security_group.example_sg.id}" ]
	}
}
`, randInt)

}

func testAccMskClusterConfig_enhancedMonitoring(randInt int) string {
	return testAccMskClusterBaseConfig() + fmt.Sprintf(`

resource "aws_msk_cluster" "example" {
	cluster_name = "tf-test-%d"
	kafka_version = "1.1.1"
	number_of_broker_nodes = 3
	broker_node_group_info {
		instance_type = "kafka.m5.large"
		ebs_volume_size = 10
		client_subnets = [ "${aws_subnet.example_subnet_az1.id}", "${aws_subnet.example_subnet_az2.id}", "${aws_subnet.example_subnet_az3.id}" ] 
		security_groups = [ "${aws_security_group.example_sg.id}" ]
	}
	enhanced_monitoring = "PER_BROKER"
}
`, randInt)

}

func testAccMskClusterConfig_brokerNodes(randInt int) string {
	return testAccMskClusterBaseConfig() + fmt.Sprintf(`

resource "aws_msk_cluster" "example" {
	cluster_name = "tf-test-%d"
	kafka_version = "2.1.0"
	number_of_broker_nodes = 6
	broker_node_group_info {
		instance_type = "kafka.m5.large"
		ebs_volume_size = 1
		client_subnets = [ "${aws_subnet.example_subnet_az1.id}", "${aws_subnet.example_subnet_az2.id}", "${aws_subnet.example_subnet_az3.id}" ] 
		security_groups = [ "${aws_security_group.example_sg.id}" ]
	}
}
`, randInt)

}

func testAccMskClusterConfig_tagsUpdate(randInt int) string {
	return testAccMskClusterBaseConfig() + fmt.Sprintf(`

resource "aws_msk_cluster" "example" {
	cluster_name = "tf-test-%d"
	kafka_version = "1.1.1"
	number_of_broker_nodes = 3
	broker_node_group_info {
		instance_type = "kafka.m5.large"
		ebs_volume_size = 10
		client_subnets = [ "${aws_subnet.example_subnet_az1.id}", "${aws_subnet.example_subnet_az2.id}", "${aws_subnet.example_subnet_az3.id}" ] 
		security_groups = [ "${aws_security_group.example_sg.id}" ]
	}
	tags = {
		foo = "baz"
		new = "type"
	}
}
`, randInt)

}
