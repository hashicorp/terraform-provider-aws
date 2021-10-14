package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/kafka"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/kafka/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
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
	input := &kafka.ListClustersInput{}
	sweepResources := make([]*testSweepResource, 0)

	err = conn.ListClustersPages(input, func(page *kafka.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.ClusterInfoList {
			r := resourceAwsMskCluster()
			d := r.Data(nil)
			d.SetId(aws.StringValue(cluster.ClusterArn))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping MSK Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing MSK Clusters (%s): %w", region, err)
	}

	err = testSweepResourceOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MSK Clusters (%s): %w", region, err)
	}

	return nil
}

const (
	mskClusterPortPlaintext = 9092
	mskClusterPortSaslScram = 9096
	mskClusterPortSaslIam   = 9098
	mskClusterPortTls       = 9094

	mskClusterPortZookeeper = 2181
)

const (
	mskClusterBrokerRegexpFormat = `^(([-\w]+\.){1,}[\w]+:%[1]d,){2,}([-\w]+\.){1,}[\w]+:%[1]d+$`
)

var (
	mskClusterBoostrapBrokersRegexp          = regexp.MustCompile(fmt.Sprintf(mskClusterBrokerRegexpFormat, mskClusterPortPlaintext))
	mskClusterBoostrapBrokersSaslScramRegexp = regexp.MustCompile(fmt.Sprintf(mskClusterBrokerRegexpFormat, mskClusterPortSaslScram))
	mskClusterBoostrapBrokersSaslIamRegexp   = regexp.MustCompile(fmt.Sprintf(mskClusterBrokerRegexpFormat, mskClusterPortSaslIam))
	mskClusterBoostrapBrokersTlsRegexp       = regexp.MustCompile(fmt.Sprintf(mskClusterBrokerRegexpFormat, mskClusterPortTls))

	mskClusterZookeeperConnectStringRegexp = regexp.MustCompile(fmt.Sprintf(mskClusterBrokerRegexpFormat, mskClusterPortZookeeper))
)

func TestAccAWSMskCluster_basic(t *testing.T) {
	var cluster kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kafka", regexp.MustCompile(`cluster/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.az_distribution", kafka.BrokerAZDistributionDefault),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.ebs_volume_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.client_subnets.#", "3"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az1", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az2", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az3", "id"),
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
					resource.TestMatchResourceAttr(resourceName, "zookeeper_connect_string", mskClusterZookeeperConnectStringRegexp),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_tls", mskClusterBoostrapBrokersTlsRegexp),
					acctest.CheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_tls"),
					acctest.CheckResourceAttrIsSortedCSV(resourceName, "zookeeper_connect_string"),
					acctest.CheckResourceAttrIsSortedCSV(resourceName, "zookeeper_connect_string_tls"),
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

func TestAccAWSMskCluster_BrokerNodeGroupInfo_EbsVolumeSize(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigBrokerNodeGroupInfoEbsVolumeSize(rName, 11),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster1),
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
				Config: testAccMskClusterConfigBrokerNodeGroupInfoEbsVolumeSize(rName, 112),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster2),
					testAccCheckMskClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.ebs_volume_size", "112"),
				),
			},
		},
	})
}

func TestAccAWSMskCluster_BrokerNodeGroupInfo_InstanceType(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigBrokerNodeGroupInfoInstanceType(rName, "kafka.t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster1),
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
				Config: testAccMskClusterConfigBrokerNodeGroupInfoInstanceType(rName, "kafka.m5.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster2),
					testAccCheckMskClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.instance_type", "kafka.m5.large"),
				),
			},
		},
	})
}

func TestAccAWSMskCluster_ClientAuthentication_Sasl_Scram(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigClientAuthenticationSaslScram(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.scram", "true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", mskClusterBoostrapBrokersSaslScramRegexp),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_tls", ""),
					acctest.CheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_sasl_scram"),
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
				Config: testAccMskClusterConfigClientAuthenticationSaslScram(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster2),
					testAccCheckMskClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.scram", "false"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_tls", mskClusterBoostrapBrokersTlsRegexp),
					acctest.CheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_tls"),
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

func TestAccAWSMskCluster_ClientAuthentication_Sasl_Iam(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigClientAuthenticationSaslIam(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.iam", "true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_sasl_iam", mskClusterBoostrapBrokersSaslIamRegexp),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_tls", ""),
					acctest.CheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_sasl_iam"),
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
				Config: testAccMskClusterConfigClientAuthenticationSaslIam(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster2),
					testAccCheckMskClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.iam", "false"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_iam", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_tls", mskClusterBoostrapBrokersTlsRegexp),
					acctest.CheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_tls"),
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

func TestAccAWSMskCluster_ClientAuthentication_Tls_CertificateAuthorityArns(t *testing.T) {
	var cluster1 kafka.ClusterInfo
	var ca acmpca.CertificateAuthority
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"
	acmCAResourceName := "aws_acmpca_certificate_authority.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			// We need to create and activate the CA before creating the MSK cluster.
			{
				Config: testAccMskClusterConfigRootCA(commonName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(acmCAResourceName, &ca),
					acctest.CheckACMPCACertificateAuthorityActivateCA(&ca),
				),
			},
			{
				Config: testAccMskClusterConfigClientAuthenticationTlsCertificateAuthorityArns(rName, commonName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.tls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.tls.0.certificate_authority_arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "client_authentication.0.tls.0.certificate_authority_arns.*", acmCAResourceName, "arn"),
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
				Config: testAccMskClusterConfigClientAuthenticationTlsCertificateAuthorityArns(rName, commonName),
				Check: resource.ComposeTestCheckFunc(
					// CA must be DISABLED for deletion.
					acctest.CheckACMPCACertificateAuthorityDisableCA(&ca),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSMskCluster_ConfigurationInfo_Revision(t *testing.T) {

	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	configurationResourceName := "aws_msk_configuration.test"
	configurationResourceName2 := "aws_msk_configuration.test2"
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigConfigurationInfoRevision1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster1),
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
				Config: testAccMskClusterConfigConfigurationInfoRevision2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster2),
					testAccCheckMskClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "configuration_info.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_info.0.arn", configurationResourceName2, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_info.0.revision", configurationResourceName2, "latest_revision"),
				),
			},
		},
	})
}

func TestAccAWSMskCluster_EncryptionInfo_EncryptionAtRestKmsKeyArn(t *testing.T) {
	var cluster kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigEncryptionInfoEncryptionAtRestKmsKeyArn(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster),
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

func TestAccAWSMskCluster_EncryptionInfo_EncryptionInTransit_ClientBroker(t *testing.T) {
	var cluster1 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigEncryptionInfoEncryptionInTransitClientBroker(rName, "PLAINTEXT"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_info.0.encryption_in_transit.0.client_broker", "PLAINTEXT"),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers", mskClusterBoostrapBrokersRegexp),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_tls", ""),
					acctest.CheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers"),
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

func TestAccAWSMskCluster_EncryptionInfo_EncryptionInTransit_InCluster(t *testing.T) {
	var cluster1 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigEncryptionInfoEncryptionInTransitInCluster(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster1),
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

func TestAccAWSMskCluster_EnhancedMonitoring(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigEnhancedMonitoring(rName, "PER_BROKER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster1),
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
				Config: testAccMskClusterConfigEnhancedMonitoring(rName, "PER_TOPIC_PER_BROKER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster2),
					testAccCheckMskClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "enhanced_monitoring", kafka.EnhancedMonitoringPerTopicPerBroker),
				),
			},
		},
	})
}

func TestAccAWSMskCluster_NumberOfBrokerNodes(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigNumberOfBrokerNodes(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_tls", mskClusterBoostrapBrokersTlsRegexp),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.client_subnets.#", "3"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az1", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az2", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az3", "id"),
					resource.TestCheckResourceAttr(resourceName, "number_of_broker_nodes", "3"),
					acctest.CheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_tls"),
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
				Config: testAccMskClusterConfigNumberOfBrokerNodes(rName, 6),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster2),
					testAccCheckMskClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_tls", mskClusterBoostrapBrokersTlsRegexp),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "broker_node_group_info.0.client_subnets.#", "3"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az1", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az2", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "broker_node_group_info.0.client_subnets.*", "aws_subnet.example_subnet_az3", "id"),
					resource.TestCheckResourceAttr(resourceName, "number_of_broker_nodes", "6"),
					acctest.CheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_tls"),
				),
			},
		},
	})
}

func TestAccAWSMskCluster_OpenMonitoring(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigOpenMonitoring(rName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster1),
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
				Config: testAccMskClusterConfigOpenMonitoring(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster2),
					testAccCheckMskClusterNotRecreated(&cluster1, &cluster2),
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

func TestAccAWSMskCluster_LoggingInfo(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigLoggingInfo(rName, false, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster1),
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
				Config: testAccMskClusterConfigLoggingInfo(rName, true, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster2),
					testAccCheckMskClusterNotRecreated(&cluster1, &cluster2),
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

func TestAccAWSMskCluster_KafkaVersionUpgrade(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigKafkaVersion(rName, "2.7.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster1),
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
				Config: testAccMskClusterConfigKafkaVersion(rName, "2.8.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster2),
					testAccCheckMskClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "kafka_version", "2.8.0"),
				),
			},
		},
	})
}

func TestAccAWSMskCluster_KafkaVersionDowngrade(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigKafkaVersion(rName, "2.8.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "kafka_version", "2.8.0"),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers", mskClusterBoostrapBrokersRegexp),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_tls", mskClusterBoostrapBrokersTlsRegexp),
					acctest.CheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers"),
					acctest.CheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_tls"),
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
				Config: testAccMskClusterConfigKafkaVersion(rName, "2.7.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster2),
					testAccCheckMskClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "kafka_version", "2.7.1"),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers", mskClusterBoostrapBrokersRegexp),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_brokers_sasl_scram", ""),
					resource.TestMatchResourceAttr(resourceName, "bootstrap_brokers_tls", mskClusterBoostrapBrokersTlsRegexp),
					acctest.CheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers"),
					acctest.CheckResourceAttrIsSortedCSV(resourceName, "bootstrap_brokers_tls"),
				),
			},
		},
	})
}

func TestAccAWSMskCluster_KafkaVersionUpgradeWithConfigurationInfo(t *testing.T) {
	var cluster1, cluster2 kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	configurationResourceName1 := "aws_msk_configuration.config1"
	configurationResourceName2 := "aws_msk_configuration.config2"
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigKafkaVersionWithConfigurationInfo(rName, "2.7.1", "config1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster1),
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
				Config: testAccMskClusterConfigKafkaVersionWithConfigurationInfo(rName, "2.8.0", "config2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster2),
					testAccCheckMskClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "kafka_version", "2.8.0"),
					resource.TestCheckResourceAttr(resourceName, "configuration_info.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_info.0.arn", configurationResourceName2, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_info.0.revision", configurationResourceName2, "latest_revision"),
				),
			},
		},
	})
}

func TestAccAWSMskCluster_Tags(t *testing.T) {
	var cluster kafka.ClusterInfo
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskClusterConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster),
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
				Config: testAccMskClusterConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccMskClusterConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMskClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckMskClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).kafkaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_msk_cluster" {
			continue
		}

		_, err := finder.ClusterByARN(conn, rs.Primary.ID)

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

func testAccCheckMskClusterExists(n string, v *kafka.ClusterInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MSK Cluster ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kafkaconn

		output, err := finder.ClusterByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckMskClusterNotRecreated(i, j *kafka.ClusterInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ClusterArn) != aws.StringValue(j.ClusterArn) {
			return fmt.Errorf("MSK Cluster (%s) recreated", aws.StringValue(i.ClusterArn))
		}

		return nil
	}
}

func testAccCheckMskClusterRecreated(i, j *kafka.ClusterInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ClusterArn) == aws.StringValue(j.ClusterArn) {
			return fmt.Errorf("MSK Cluster (%s) was not recreated", aws.StringValue(i.ClusterArn))
		}

		return nil
	}
}

func testAccPreCheckAWSMsk(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).kafkaconn

	input := &kafka.ListClustersInput{}

	_, err := conn.ListClusters(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccMskClusterBaseConfig(rName string) string {
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
}
`, rName))
}

func testAccMskClusterConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
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

func testAccMskClusterConfigBrokerNodeGroupInfoEbsVolumeSize(rName string, ebsVolumeSize int) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
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

func testAccMskClusterConfigBrokerNodeGroupInfoInstanceType(rName string, t string) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
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

func testAccMskClusterConfigRootCA(commonName string) string {
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

func testAccMskClusterConfigClientAuthenticationTlsCertificateAuthorityArns(rName, commonName string) string {
	return acctest.ConfigCompose(
		testAccMskClusterBaseConfig(rName),
		testAccMskClusterConfigRootCA(commonName),
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
    tls {
      certificate_authority_arns = [aws_acmpca_certificate_authority.test.arn]
    }
  }

  encryption_info {
    encryption_in_transit {
      client_broker = "TLS"
    }
  }
}
`, rName, commonName))
}

func testAccMskClusterConfigClientAuthenticationSaslScram(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
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
  }
}
`, rName, enabled))
}

func testAccMskClusterConfigClientAuthenticationSaslIam(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
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
  }
}
`, rName, enabled))
}

func testAccMskClusterConfigConfigurationInfoRevision1(rName string) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
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
    arn      = aws_msk_configuration.test.arn
    revision = aws_msk_configuration.test.latest_revision
  }
}
`, rName))
}

func testAccMskClusterConfigConfigurationInfoRevision2(rName string) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
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

func testAccMskClusterConfigEncryptionInfoEncryptionAtRestKmsKeyArn(rName string) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
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

func testAccMskClusterConfigEncryptionInfoEncryptionInTransitClientBroker(rName, clientBroker string) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
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

func testAccMskClusterConfigEncryptionInfoEncryptionInTransitInCluster(rName string, inCluster bool) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
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

func testAccMskClusterConfigEnhancedMonitoring(rName, enhancedMonitoring string) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
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

func testAccMskClusterConfigNumberOfBrokerNodes(rName string, brokerCount int) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
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

func testAccMskClusterConfigOpenMonitoring(rName string, jmxExporterEnabled bool, nodeExporterEnabled bool) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
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

func testAccMskClusterConfigLoggingInfo(rName string, cloudwatchLogsEnabled bool, firehoseEnabled bool, s3Enabled bool) string {
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

	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "bucket" {
  acl           = "private"
  bucket        = %[1]q
  force_destroy = true
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

func testAccMskClusterConfigKafkaVersion(rName string, kafkaVersion string) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
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

func testAccMskClusterConfigKafkaVersionWithConfigurationInfo(rName string, kafkaVersion string, configResourceName string) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
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

func testAccMskClusterConfigTags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
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

func testAccMskClusterConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccMskClusterBaseConfig(rName), fmt.Sprintf(`
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
