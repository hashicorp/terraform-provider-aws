package kafka_test

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/kafka"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	clusterPortPlaintext     = 9092
	clusterPortSASLIAM       = 9098
	clusterPortSASLIAMPublic = 9198
	clusterPortSASLScram     = 9096
	clusterPortTLS           = 9094

	clusterPortZookeeper = 2181
)

const (
	clusterBrokerRegexpFormat = `^(([-\w]+\.){1,}[\w]+:%[1]d,){2,}([-\w]+\.){1,}[\w]+:%[1]d+$`
)

var (
	clusterBoostrapBrokersRegexp              = regexp.MustCompile(fmt.Sprintf(clusterBrokerRegexpFormat, clusterPortPlaintext))
	clusterBoostrapBrokersSASLIAMRegexp       = regexp.MustCompile(fmt.Sprintf(clusterBrokerRegexpFormat, clusterPortSASLIAM))
	clusterBoostrapBrokersSASLIAMPublicRegexp = regexp.MustCompile(fmt.Sprintf(clusterBrokerRegexpFormat, clusterPortSASLIAMPublic))
	clusterBoostrapBrokersSASLScramRegexp     = regexp.MustCompile(fmt.Sprintf(clusterBrokerRegexpFormat, clusterPortSASLScram))
	clusterBoostrapBrokersTLSRegexp           = regexp.MustCompile(fmt.Sprintf(clusterBrokerRegexpFormat, clusterPortTLS))

	clusterZookeeperConnectStringRegexp = regexp.MustCompile(fmt.Sprintf(clusterBrokerRegexpFormat, clusterPortZookeeper))
)

func TestAccKafkaCluster_basic(t *testing.T) {
	var cluster kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kafka", regexp.MustCompile(`cluster/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_public_sasl_iam", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_public_sasl_scram", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_public_tls", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_tls", clusterBoostrapBrokersTLSRegexp),
					testAccCheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_tls"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.az_distribution", kafka.BrokerAZDistributionDefault),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.ebs_volume_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.client_subnets.#", "3"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az1", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az2", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az3", "id"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.public_access.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.public_access.0.type", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.instance_type", "kafka.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.security_groups.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.security_groups.*", "aws_security_group.example_sg", "id"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cluster_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.#", "1"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "encryption_info.0.encryption_at_rest_kms_key_arn", "kms", regexp.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.0.client_broker", "TLS"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.0.in_cluster", "true"),
					resource.TestCheckResourceAttr(resourceName, "enhanced_monitoring", kafka.EnhancedMonitoringDefault),
					resource.TestCheckResourceAttr(resourceName, "kafka_version", "2.7.1"),
					resource.TestCheckResourceAttr(resourceName, "number_of_broker_nodes", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "zookeeper_connect_string", clusterZookeeperConnectStringRegexp),
					testAccCheckResourceAttrIsSortedCSV(resourceName, "zookeeper_connect_string"),
					testAccCheckResourceAttrIsSortedCSV(resourceName, "zookeeper_connect_string_tls"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
		},
	})
}

func TestAccKafkaCluster_BrokerNodeGroupInfo_ebsVolumeSize(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_deprecatedBrokerNodeGroupInfoEBSVolumeSize(rName, 11),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.ebs_volume_size", "11"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
			{
				// BadRequestException: The minimum increase in storage size of the cluster should be atleast 100GB
				Config: testAccClusterConfig_deprecatedBrokerNodeGroupInfoEBSVolumeSize(rName, 112),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.ebs_volume_size", "112"),
				),
			},
		},
	})
}

func TestAccKafkaCluster_BrokerNodeGroupInfo_storageInfo(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"
	original_volume_size := 11
	updated_volume_size := 112

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_brokerNodeGroupInfoStorageInfoVolumeSizeSetAndProvThroughputNotEnabled(rName, original_volume_size, "kafka.m5.4xlarge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.ebs_volume_size", strconv.Itoa(original_volume_size)),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.0.volume_size", strconv.Itoa(original_volume_size)),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.0.provisioned_throughput.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.0.provisioned_throughput.0.enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
			{
				// update broker_node_group_info.0.storage_info.0.ebs_storage_info.0.volume_size
				Config: testAccClusterConfig_brokerNodeGroupInfoStorageInfoVolumeSizeSetAndProvThroughputEnabled(rName, updated_volume_size, "kafka.m5.4xlarge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.ebs_volume_size", strconv.Itoa(updated_volume_size)),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.0.volume_size", strconv.Itoa(updated_volume_size)),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.0.provisioned_throughput.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.0.provisioned_throughput.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.0.provisioned_throughput.0.volume_throughput", "250"),
				),
			},
		},
	})
}

func TestAccKafkaCluster_BrokerNodeGroupInfo_modifyEBSVolumeSizeToStorageInfo(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"
	original_volume_size := 11
	updated_volume_size := 112

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				// init with the deprecated ebs_volume_size
				Config: testAccClusterConfig_deprecatedBrokerNodeGroupInfoEBSVolumeSize(rName, original_volume_size),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.ebs_volume_size", strconv.Itoa(original_volume_size)),
				),
			},
			{
				// refactor deprecated ebs_volume_size to storage_info
				Config:   testAccClusterConfig_brokerNodeGroupInfoStorageInfoVolumeSizeOnly(rName, original_volume_size, "kafka.m5.large"),
				PlanOnly: true,
			},
			{
				// upgrade the instance type, update storage, and enable provisioned throughput
				Config: testAccClusterConfig_brokerNodeGroupInfoStorageInfoVolumeSizeSetAndProvThroughputEnabled(rName, updated_volume_size, "kafka.m5.4xlarge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.0.volume_size", strconv.Itoa(updated_volume_size)),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.0.provisioned_throughput.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.0.provisioned_throughput.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.storage_info.0.ebs_storage_info.0.provisioned_throughput.0.volume_throughput", "250"),
				),
			},
		},
	})
}

func TestAccKafkaCluster_BrokerNodeGroupInfo_instanceType(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_brokerNodeGroupInfoInstanceType(rName, "kafka.t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.instance_type", "kafka.t3.small"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bootstrap_brokers",     // API may mutate ordering and selection of brokers to return
					"bootstrap_brokers_tls", // API may mutate ordering and selection of brokers to return
					"current_version",
				},
			},
			{
				Config: testAccClusterConfig_brokerNodeGroupInfoInstanceType(rName, "kafka.m5.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.instance_type", "kafka.m5.large"),
				),
			},
		},
	})
}

func TestAccKafkaCluster_BrokerNodeGroupInfo_publicAccessSASLIAM(t *testing.T) {
	var cluster1 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_brokerNodeGroupInfoNoPublicAccessSASLIAM(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_public_sasl_iam", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_public_sasl_scram", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_public_tls", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_sasl_iam", clusterBoostrapBrokersSASLIAMRegexp),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_tls", ""),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.public_access.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.public_access.0.type", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.iam", "true"),
				),
			},
			{
				Config: testAccClusterConfig_brokerNodeGroupInfoPublicAccessSASLIAM(rName, "SERVICE_PROVIDED_EIPS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_public_sasl_iam", clusterBoostrapBrokersSASLIAMPublicRegexp),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_public_sasl_scram", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_public_tls", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_sasl_iam", clusterBoostrapBrokersSASLIAMRegexp),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_tls", ""),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.public_access.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.public_access.0.type", "SERVICE_PROVIDED_EIPS"),
				),
			},
			{
				Config: testAccClusterConfig_brokerNodeGroupInfoPublicAccessSASLIAM(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_public_sasl_iam", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_public_sasl_scram", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_public_tls", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_sasl_iam", clusterBoostrapBrokersSASLIAMRegexp),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_tls", ""),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.public_access.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.connectivity_info.0.public_access.0.type", "DISABLED"),
				),
			},
		},
	})
}

func TestAccKafkaCluster_ClientAuthenticationSASL_scram(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_clientAuthenticationSASLScram(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.iam", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.scram", "true"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.unauthenticated", "false"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", clusterBoostrapBrokersSASLScramRegexp),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_tls", ""),
					testAccCheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_sasl_scram"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
			{
				Config: testAccClusterConfig_clientAuthenticationSASLScram(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.iam", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.scram", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.unauthenticated", "true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_tls", clusterBoostrapBrokersTLSRegexp),
					testAccCheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_tls"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
		},
	})
}

func TestAccKafkaCluster_ClientAuthenticationSASL_iam(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_clientAuthenticationSASLIAM(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.iam", "true"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.scram", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.unauthenticated", "false"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_sasl_iam", clusterBoostrapBrokersSASLIAMRegexp),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_tls", ""),
					testAccCheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_sasl_iam"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
			{
				Config: testAccClusterConfig_clientAuthenticationSASLIAM(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.iam", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.scram", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.tls.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.unauthenticated", "true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_iam", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_tls", clusterBoostrapBrokersTLSRegexp),
					testAccCheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_tls"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
		},
	})
}

func TestAccKafkaCluster_ClientAuthenticationTLS_certificateAuthorityARNs(t *testing.T) {
	var cluster1 kafka.ClusterInfo
	var ca acmpca.CertificateAuthority
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"
	acmCAResourceName := "aws_acmpca_certificate_authority.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			// We need to create and activate the CA before creating the MSK cluster.
			{
				Config: testAccClusterConfig_rootCA(commonName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(acmCAResourceName, &ca),
					acctest.CheckACMPCACertificateAuthorityActivateRootCA(&ca),
				),
			},
			{
				Config: testAccClusterConfig_clientAuthenticationTLSCertificateAuthorityARNs(rName, commonName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.iam", "true"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.scram", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.tls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.tls.0.certificate_authority_arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "client_authentication.0.tls.0.certificate_authority_arns.*", acmCAResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.unauthenticated", "false"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.0.client_broker", "TLS"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.0.in_cluster", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
			{
				Config: testAccClusterConfig_clientAuthenticationTLSCertificateAuthorityARNs(rName, commonName),
				Check: resource.ComposeTestCheckFunc(
					// CA must be DISABLED for deletion.
					acctest.CheckACMPCACertificateAuthorityDisableCA(&ca),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKafkaCluster_ClientAuthenticationTLS_initiallyNoAuthentication(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	var ca acmpca.CertificateAuthority
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"
	acmCAResourceName := "aws_acmpca_certificate_authority.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			// We need to create and activate the CA before creating the MSK cluster.
			{
				Config: testAccClusterConfig_rootCA(commonName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(acmCAResourceName, &ca),
					acctest.CheckACMPCACertificateAuthorityActivateRootCA(&ca),
				),
			},
			{
				Config: testAccClusterConfig_rootCANoClientAuthentication(rName, commonName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.0.client_broker", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.0.in_cluster", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
			{
				Config: testAccClusterConfig_clientAuthenticationTLSCertificateAuthorityARNs(rName, commonName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.iam", "true"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.scram", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.tls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.tls.0.certificate_authority_arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "client_authentication.0.tls.0.certificate_authority_arns.*", acmCAResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.unauthenticated", "false"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.0.client_broker", "TLS"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.0.in_cluster", "true"),
				),
			},
			{
				Config: testAccClusterConfig_clientAuthenticationTLSCertificateAuthorityARNs(rName, commonName),
				Check: resource.ComposeTestCheckFunc(
					// CA must be DISABLED for deletion.
					acctest.CheckACMPCACertificateAuthorityDisableCA(&ca),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKafkaCluster_Info_revision(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	configurationResourceName := "aws_msk_configuration.test1"
	configurationResourceName2 := "aws_msk_configuration.test2"
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_configurationInfoRevision1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "configuration_info.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_info.0.arn", configurationResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_info.0.revision", configurationResourceName, "latest_revision"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
			{
				Config: testAccClusterConfig_configurationInfoRevision2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "configuration_info.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_info.0.arn", configurationResourceName2, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_info.0.revision", configurationResourceName2, "latest_revision"),
				),
			},
		},
	})
}

func TestAccKafkaCluster_EncryptionInfo_encryptionAtRestKMSKeyARN(t *testing.T) {
	var cluster kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encryptionInfoEncryptionAtRestKMSKeyARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_info.0.encryption_at_rest_kms_key_arn", "aws_kms_key.example_key", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
		},
	})
}

func TestAccKafkaCluster_EncryptionInfoEncryptionInTransit_clientBroker(t *testing.T) {
	var cluster1 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encryptionInfoEncryptionInTransitClientBroker(rName, "PLAINTEXT"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.0.client_broker", "PLAINTEXT"),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers", clusterBoostrapBrokersRegexp),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_tls", ""),
					testAccCheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
		},
	})
}

func TestAccKafkaCluster_EncryptionInfoEncryptionInTransit_inCluster(t *testing.T) {
	var cluster1 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encryptionInfoEncryptionInTransitIn(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.0.in_cluster", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
		},
	})
}

func TestAccKafkaCluster_enhancedMonitoring(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_enhancedMonitoring(rName, "PER_BROKER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "enhanced_monitoring", kafka.EnhancedMonitoringPerBroker),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
			{
				Config: testAccClusterConfig_enhancedMonitoring(rName, "PER_TOPIC_PER_BROKER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "enhanced_monitoring", kafka.EnhancedMonitoringPerTopicPerBroker),
				),
			},
		},
	})
}

func TestAccKafkaCluster_numberOfBrokerNodes(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_numberOfBrokerNodes(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_tls", clusterBoostrapBrokersTLSRegexp),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.client_subnets.#", "3"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az1", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az2", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az3", "id"),
					resource.TestCheckResourceAttr(resourceName, "number_of_broker_nodes", "3"),
					testAccCheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_tls"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
			{
				Config: testAccClusterConfig_numberOfBrokerNodes(rName, 6),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_tls", clusterBoostrapBrokersTLSRegexp),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.client_subnets.#", "3"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az1", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az2", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az3", "id"),
					resource.TestCheckResourceAttr(resourceName, "number_of_broker_nodes", "6"),
					testAccCheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_tls"),
				),
			},
		},
	})
}

func TestAccKafkaCluster_openMonitoring(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_openMonitoring(rName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "open_monitoring.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "open_monitoring.0.prometheus.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "open_monitoring.0.prometheus.0.jmx_exporter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "open_monitoring.0.prometheus.0.jmx_exporter.0.enabled_in_broker", "false"),
					resource.TestCheckResourceAttr(resourceName, "open_monitoring.0.prometheus.0.node_exporter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "open_monitoring.0.prometheus.0.node_exporter.0.enabled_in_broker", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
			{
				Config: testAccClusterConfig_openMonitoring(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "open_monitoring.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "open_monitoring.0.prometheus.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "open_monitoring.0.prometheus.0.jmx_exporter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "open_monitoring.0.prometheus.0.jmx_exporter.0.enabled_in_broker", "true"),
					resource.TestCheckResourceAttr(resourceName, "open_monitoring.0.prometheus.0.node_exporter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "open_monitoring.0.prometheus.0.node_exporter.0.enabled_in_broker", "false"),
				),
			},
		},
	})
}

func TestAccKafkaCluster_loggingInfo(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_loggingInfo(rName, false, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "logging_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_info.0.broker_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_info.0.broker_logs.0.cloudwatch_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_info.0.broker_logs.0.cloudwatch_logs.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "logging_info.0.broker_logs.0.firehose.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_info.0.broker_logs.0.firehose.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "logging_info.0.broker_logs.0.s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_info.0.broker_logs.0.s3.0.enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
			{
				Config: testAccClusterConfig_loggingInfo(rName, true, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "logging_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_info.0.broker_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_info.0.broker_logs.0.cloudwatch_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_info.0.broker_logs.0.cloudwatch_logs.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "logging_info.0.broker_logs.0.firehose.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_info.0.broker_logs.0.firehose.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "logging_info.0.broker_logs.0.s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_info.0.broker_logs.0.s3.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccKafkaCluster_kafkaVersionUpgrade(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_version(rName, "2.7.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "kafka_version", "2.7.1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
			{
				Config: testAccClusterConfig_version(rName, "2.8.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "kafka_version", "2.8.0"),
				),
			},
		},
	})
}

func TestAccKafkaCluster_kafkaVersionDowngrade(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_version(rName, "2.8.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "kafka_version", "2.8.0"),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers", clusterBoostrapBrokersRegexp),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_tls", clusterBoostrapBrokersTLSRegexp),
					testAccCheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers"),
					testAccCheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_tls"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
			{
				Config: testAccClusterConfig_version(rName, "2.7.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "kafka_version", "2.7.1"),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers", clusterBoostrapBrokersRegexp),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_tls", clusterBoostrapBrokersTLSRegexp),
					testAccCheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers"),
					testAccCheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_tls"),
				),
			},
		},
	})
}

func TestAccKafkaCluster_kafkaVersionUpgradeWithInfo(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	configurationResourceName1 := "aws_msk_configuration.config1"
	configurationResourceName2 := "aws_msk_configuration.config2"
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_versionConfigurationInfo(rName, "2.7.1", "config1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "kafka_version", "2.7.1"),
					resource.TestCheckResourceAttr(resourceName, "configuration_info.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_info.0.arn", configurationResourceName1, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_info.0.revision", configurationResourceName1, "latest_revision"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
			{
				Config: testAccClusterConfig_versionConfigurationInfo(rName, "2.8.0", "config2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "kafka_version", "2.8.0"),
					resource.TestCheckResourceAttr(resourceName, "configuration_info.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_info.0.arn", configurationResourceName2, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_info.0.revision", configurationResourceName2, "latest_revision"),
				),
			},
		},
	})
}

func TestAccKafkaCluster_tags(t *testing.T) {
	var cluster kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafka.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"current_version",
				},
			},
			{
				Config: testAccClusterConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccClusterConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckResourceAttrIsSortedCSV(resourceName, attributeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		is, err := acctest.PrimaryInstanceState(s, resourceName)
		if err != nil {
			return err
		}

		v, ok := is.Attributes[attributeName]
		if !ok {
			return fmt.Errorf("%s: No attribute %q found", resourceName, attributeName)
		}

		splitV := strings.Split(v, ",")
		if !sort.StringsAreSorted(splitV) {
			return fmt.Errorf("%s: Expected attribute %q to be sorted, got %q", resourceName, attributeName, v)
		}

		return nil
	}
}

func testAccCheckClusterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_msk_cluster" {
			continue
		}

		_, err := tfkafka.FindClusterByARN(context.TODO(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("MSK Cluster %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckClusterExists(n string, v *kafka.ClusterInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MSK Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConn

		output, err := tfkafka.FindClusterByARN(context.TODO(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClusterNotRecreated(i, j *kafka.ClusterInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ClusterArn) != aws.StringValue(j.ClusterArn) {
			return fmt.Errorf("MSK Cluster (%s) recreated", aws.StringValue(i.ClusterArn))
		}

		return nil
	}
}

func testAccCheckClusterRecreated(i, j *kafka.ClusterInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ClusterArn) == aws.StringValue(j.ClusterArn) {
			return fmt.Errorf("MSK Cluster (%s) was not recreated", aws.StringValue(i.ClusterArn))
		}

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConn

	input := &kafka.ListClustersInput{}

	_, err := conn.ListClusters(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccClusterBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "example_vpc" {
  cidr_block = "192.168.0.0/22"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "example_subnet_az1" {
  vpc_id            = aws_vpc.example_vpc.id
  cidr_block        = "192.168.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "example_subnet_az2" {
  vpc_id            = aws_vpc.example_vpc.id
  cidr_block        = "192.168.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "example_subnet_az3" {
  vpc_id            = aws_vpc.example_vpc.id
  cidr_block        = "192.168.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[2]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "example_sg" {
  vpc_id = aws_vpc.example_vpc.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccClusterBasePublicAccessConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "example_vpc" {
  cidr_block = "192.168.0.0/21"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "example_subnet_az1" {
  vpc_id            = aws_vpc.example_vpc.id
  cidr_block        = "192.168.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "example_subnet_az2" {
  vpc_id            = aws_vpc.example_vpc.id
  cidr_block        = "192.168.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "example_subnet_az3" {
  vpc_id            = aws_vpc.example_vpc.id
  cidr_block        = "192.168.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[2]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "example_sg" {
  vpc_id = aws_vpc.example_vpc.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "example_public_subnet_az1" {
  vpc_id                  = aws_vpc.example_vpc.id
  cidr_block              = "192.168.3.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "example_public_subnet_az2" {
  vpc_id                  = aws_vpc.example_vpc.id
  cidr_block              = "192.168.4.0/24"
  availability_zone       = data.aws_availability_zones.available.names[1]
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "example_public_subnet_az3" {
  vpc_id                  = aws_vpc.example_vpc.id
  cidr_block              = "192.168.5.0/24"
  availability_zone       = data.aws_availability_zones.available.names[2]
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "example_igw" {
  vpc_id = aws_vpc.example_vpc.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "example_route_tbl" {
  vpc_id = aws_vpc.example_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.example_igw.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "route_tbl_assoc_1" {
  subnet_id      = aws_subnet.example_public_subnet_az1.id
  route_table_id = aws_route_table.example_route_tbl.id
}

resource "aws_route_table_association" "route_tbl_assoc_2" {
  subnet_id      = aws_subnet.example_public_subnet_az2.id
  route_table_id = aws_route_table.example_route_tbl.id
}

resource "aws_route_table_association" "route_tbl_assoc_3" {
  subnet_id      = aws_subnet.example_public_subnet_az3.id
  route_table_id = aws_route_table.example_route_tbl.id
}
`, rName))
}

func testAccClusterConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }
}
`, rName))
}

func testAccClusterConfig_deprecatedBrokerNodeGroupInfoEBSVolumeSize(rName string, ebsVolumeSize int) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = %[2]d
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }
}
`, rName, ebsVolumeSize))
}

func testAccClusterConfig_brokerNodeGroupInfoStorageInfoVolumeSizeOnly(rName string, ebsVolumeSize int, instanceType string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    instance_type   = %[3]q
    security_groups = [aws_security_group.example_sg.id]
    storage_info {
      ebs_storage_info {
        volume_size = %[2]d
      }
    }
  }
}
`, rName, ebsVolumeSize, instanceType))
}

func testAccClusterConfig_brokerNodeGroupInfoStorageInfoVolumeSizeSetAndProvThroughputNotEnabled(rName string, ebsVolumeSize int, instanceType string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    instance_type   = %[3]q
    security_groups = [aws_security_group.example_sg.id]
    storage_info {
      ebs_storage_info {
        provisioned_throughput {
          enabled = false
        }
        volume_size = %[2]d
      }
    }
  }
}
`, rName, ebsVolumeSize, instanceType))
}

func testAccClusterConfig_brokerNodeGroupInfoStorageInfoVolumeSizeSetAndProvThroughputEnabled(rName string, ebsVolumeSize int, instanceType string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    instance_type   = %[3]q
    security_groups = [aws_security_group.example_sg.id]
    storage_info {
      ebs_storage_info {
        provisioned_throughput {
          enabled           = true
          volume_throughput = 250
        }
        volume_size = %[2]d
      }
    }
  }
}
`, rName, ebsVolumeSize, instanceType))
}

func testAccClusterConfig_brokerNodeGroupInfoInstanceType(rName string, t string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = %[2]q
    security_groups = [aws_security_group.example_sg.id]
  }
}
`, rName, t))
}

func testAccConfigurationAllowEveryoneNoACLFoundFalse(rName string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  kafka_versions = ["2.7.1"]
  name           = %[1]q

  server_properties = <<-PROPERTIES
    allow.everyone.if.no.acl.found = false
  PROPERTIES
}`, rName)
}

func testAccClusterConfig_brokerNodeGroupInfoNoPublicAccessSASLIAM(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBasePublicAccessConfig(rName),
		testAccConfigurationAllowEveryoneNoACLFoundFalse(rName),
		fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_public_subnet_az1.id, aws_subnet.example_public_subnet_az2.id, aws_subnet.example_public_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  configuration_info {
    arn      = aws_msk_configuration.test.arn
    revision = aws_msk_configuration.test.latest_revision
  }

  client_authentication {
    sasl {
      iam = true
    }
  }
}
`, rName))
}

func testAccClusterConfig_brokerNodeGroupInfoPublicAccessSASLIAM(rName string, pa string) string {
	return acctest.ConfigCompose(
		testAccClusterBasePublicAccessConfig(rName),
		testAccConfigurationAllowEveryoneNoACLFoundFalse(rName),
		fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_public_subnet_az1.id, aws_subnet.example_public_subnet_az2.id, aws_subnet.example_public_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]

    connectivity_info {
      public_access {
        type = %[2]q
      }
    }
  }

  configuration_info {
    arn      = aws_msk_configuration.test.arn
    revision = aws_msk_configuration.test.latest_revision
  }

  client_authentication {
    sasl {
      iam = true
    }
  }
}
`, rName, pa))
}

func testAccClusterConfig_rootCA(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}
`, commonName)
}

func testAccClusterConfig_clientAuthenticationTLSCertificateAuthorityARNs(rName, commonName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseConfig(rName),
		testAccClusterConfig_rootCA(commonName),
		fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  client_authentication {
    sasl {
      iam = true
    }

    tls {
      certificate_authority_arns = [aws_acmpca_certificate_authority.test.arn]
    }
  }

  encryption_info {
    encryption_in_transit {
      client_broker = "TLS"
      in_cluster    = true
    }
  }
}
`, rName, commonName))
}

func testAccClusterConfig_rootCANoClientAuthentication(rName, commonName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseConfig(rName),
		testAccClusterConfig_rootCA(commonName),
		fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  encryption_info {
    encryption_in_transit {
      client_broker = "PLAINTEXT"
      in_cluster    = true
    }
  }
}
`, rName, commonName))
}

func testAccClusterConfig_clientAuthenticationSASLScram(rName string, scramEnabled bool) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  client_authentication {
    sasl {
      scram = %[2]t
    }

    unauthenticated = %[3]t
  }
}
`, rName, scramEnabled, !scramEnabled))
}

func testAccClusterConfig_clientAuthenticationSASLIAM(rName string, saslEnabled bool) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  client_authentication {
    sasl {
      iam = %[2]t
    }

    unauthenticated = %[3]t
  }
}
`, rName, saslEnabled, !saslEnabled))
}

func testAccClusterConfig_configurationInfoRevision1(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_configuration" "test1" {
  kafka_versions = ["2.7.1"]
  name           = "%[1]s-1"

  server_properties = <<PROPERTIES
log.cleaner.delete.retention.ms = 86400000
PROPERTIES
}

resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  configuration_info {
    arn      = aws_msk_configuration.test1.arn
    revision = aws_msk_configuration.test1.latest_revision
  }
}
`, rName))
}

func testAccClusterConfig_configurationInfoRevision2(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_configuration" "test1" {
  kafka_versions = ["2.7.1"]
  name           = "%[1]s-1"

  server_properties = <<PROPERTIES
log.cleaner.delete.retention.ms = 86400000
PROPERTIES
}

resource "aws_msk_configuration" "test2" {
  kafka_versions = ["2.7.1"]
  name           = "%[1]s-2"

  server_properties = <<PROPERTIES
log.cleaner.delete.retention.ms = 86400001
PROPERTIES
}

resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  configuration_info {
    arn      = aws_msk_configuration.test2.arn
    revision = aws_msk_configuration.test2.latest_revision
  }
}
`, rName))
}

func testAccClusterConfig_encryptionInfoEncryptionAtRestKMSKeyARN(rName string) string { // nosemgrep:msk-in-func-name
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_kms_key" "example_key" {
  description = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  encryption_info {
    encryption_at_rest_kms_key_arn = aws_kms_key.example_key.arn
  }
}
`, rName))

}

func testAccClusterConfig_encryptionInfoEncryptionInTransitClientBroker(rName, clientBroker string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  encryption_info {
    encryption_in_transit {
      client_broker = %[2]q
    }
  }
}
`, rName, clientBroker))
}

func testAccClusterConfig_encryptionInfoEncryptionInTransitIn(rName string, inCluster bool) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  encryption_info {
    encryption_in_transit {
      in_cluster = %[2]t
    }
  }
}
`, rName, inCluster))
}

func testAccClusterConfig_enhancedMonitoring(rName, enhancedMonitoring string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  enhanced_monitoring    = %[2]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }
}
`, rName, enhancedMonitoring))

}

func testAccClusterConfig_numberOfBrokerNodes(rName string, brokerCount int) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = %[2]d

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }
}
`, rName, brokerCount))

}

func testAccClusterConfig_openMonitoring(rName string, jmxExporterEnabled bool, nodeExporterEnabled bool) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  open_monitoring {
    prometheus {
      jmx_exporter {
        enabled_in_broker = %[2]t
      }

      node_exporter {
        enabled_in_broker = %[3]t
      }
    }
  }
}
`, rName, jmxExporterEnabled, nodeExporterEnabled))
}

func testAccClusterConfig_loggingInfo(rName string, cloudwatchLogsEnabled bool, firehoseEnabled bool, s3Enabled bool) string {
	cloudwatchLogsLogGroup := "\"\""
	firehoseDeliveryStream := "\"\""
	s3Bucket := "\"\""

	if cloudwatchLogsEnabled {
		cloudwatchLogsLogGroup = "aws_cloudwatch_log_group.test.name"
	}
	if firehoseEnabled {
		firehoseDeliveryStream = "aws_kinesis_firehose_delivery_stream.test.name"
	}
	if s3Enabled {
		s3Bucket = "aws_s3_bucket.bucket.id"
	}

	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}

resource "aws_iam_role" "firehose_role" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[1]q
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose_role.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  tags = {
    LogDeliveryEnabled = "placeholder"
  }

  lifecycle {
    ignore_changes = [
      # Ignore changes to LogDeliveryEnabled tag as API adds this tag when broker log delivery is enabled
      tags["LogDeliveryEnabled"],
    ]
  }
}

resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  logging_info {
    broker_logs {
      cloudwatch_logs {
        enabled   = %[2]t
        log_group = %[3]s
      }

      firehose {
        enabled         = %[4]t
        delivery_stream = %[5]s
      }

      s3 {
        enabled = %[6]t
        bucket  = %[7]s
        prefix  = ""
      }
    }
  }
}
`, rName, cloudwatchLogsEnabled, cloudwatchLogsLogGroup, firehoseEnabled, firehoseDeliveryStream, s3Enabled, s3Bucket))
}

func testAccClusterConfig_version(rName string, kafkaVersion string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = %[2]q
  number_of_broker_nodes = 3

  encryption_info {
    encryption_in_transit {
      client_broker = "TLS_PLAINTEXT"
    }
  }

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }
}
`, rName, kafkaVersion))
}

func testAccClusterConfig_versionConfigurationInfo(rName string, kafkaVersion string, configResourceName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_configuration" "config1" {
  kafka_versions    = ["2.7.1"]
  name              = "%[1]s-1"
  server_properties = <<PROPERTIES
log.cleaner.delete.retention.ms = 86400000
PROPERTIES
}

resource "aws_msk_configuration" "config2" {
  kafka_versions    = ["2.8.0"]
  name              = "%[1]s-2"
  server_properties = <<PROPERTIES
log.cleaner.delete.retention.ms = 86400001
PROPERTIES
}

resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = %[2]q
  number_of_broker_nodes = 3

  encryption_info {
    encryption_in_transit {
      client_broker = "TLS_PLAINTEXT"
    }
  }

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  configuration_info {
    arn      = aws_msk_configuration.%[3]s.arn
    revision = aws_msk_configuration.%[3]s.latest_revision
  }
}
`, rName, kafkaVersion, configResourceName))
}

func testAccClusterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
