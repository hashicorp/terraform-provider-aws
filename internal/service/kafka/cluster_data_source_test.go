package kafka_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/kafka"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKafkaClusterDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_msk_cluster.test"
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers", resourceName, "bootstrap_brokers"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_sasl_scram", resourceName, "bootstrap_brokers_sasl_scram"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_tls", resourceName, "bootstrap_brokers_tls"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_name", resourceName, "cluster_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kafka_version", resourceName, "kafka_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "number_of_broker_nodes", resourceName, "number_of_broker_nodes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "zookeeper_connect_string", resourceName, "zookeeper_connect_string"),
				),
			},
		},
	})
}

func TestAccKafkaClusterDataSource_publicAccessSASLIam(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_msk_cluster.test"
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfigPublicAccessSASLIam(rName, "DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_sasl_iam", resourceName, "bootstrap_brokers_sasl_iam"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_name", resourceName, "cluster_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kafka_version", resourceName, "kafka_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "number_of_broker_nodes", resourceName, "number_of_broker_nodes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "zookeeper_connect_string", resourceName, "zookeeper_connect_string"),
				),
			},
			{
				Config: testAccClusterDataSourceConfigPublicAccessSASLIam(rName, "SERVICE_PROVIDED_EIPS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_sasl_iam", resourceName, "bootstrap_brokers_sasl_iam"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_public_sasl_iam", resourceName, "bootstrap_brokers_public_sasl_iam"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_name", resourceName, "cluster_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kafka_version", resourceName, "kafka_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "number_of_broker_nodes", resourceName, "number_of_broker_nodes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "zookeeper_connect_string", resourceName, "zookeeper_connect_string"),
				),
			},
		},
	})
}

func TestAccKafkaClusterDataSource_publicAccessSASLScram(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_msk_cluster.test"
	resourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfigPublicAccessSASLScram(rName, "DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_sasl_iam", resourceName, "bootstrap_brokers_sasl_iam"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_name", resourceName, "cluster_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kafka_version", resourceName, "kafka_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "number_of_broker_nodes", resourceName, "number_of_broker_nodes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "zookeeper_connect_string", resourceName, "zookeeper_connect_string"),
				),
			},
			{
				Config: testAccClusterDataSourceConfigPublicAccessSASLScram(rName, "SERVICE_PROVIDED_EIPS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_sasl_scram", resourceName, "bootstrap_brokers_sasl_scram"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_public_sasl_scram", resourceName, "bootstrap_brokers_public_sasl_scram"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_name", resourceName, "cluster_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kafka_version", resourceName, "kafka_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "number_of_broker_nodes", resourceName, "number_of_broker_nodes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "zookeeper_connect_string", resourceName, "zookeeper_connect_string"),
				),
			},
		},
	})
}

func TestAccKafkaClusterDataSource_publicAccessTLS(t *testing.T) {
	var ca acmpca.CertificateAuthority
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_msk_cluster.test"
	resourceName := "aws_msk_cluster.test"
	acmCAResourceName := "aws_acmpca_certificate_authority.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafka.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigRootCA(commonName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(acmCAResourceName, &ca),
					acctest.CheckACMPCACertificateAuthorityActivateCA(&ca),
				),
			},
			{
				Config: testAccClusterDataSourceConfigPublicAccessTLS(rName, commonName, "DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers", resourceName, "bootstrap_brokers"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_tls", resourceName, "bootstrap_brokers_tls"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_name", resourceName, "cluster_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kafka_version", resourceName, "kafka_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "number_of_broker_nodes", resourceName, "number_of_broker_nodes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "zookeeper_connect_string", resourceName, "zookeeper_connect_string"),
				),
			},
			{
				Config: testAccClusterDataSourceConfigPublicAccessTLS(rName, commonName, "SERVICE_PROVIDED_EIPS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_tls", resourceName, "bootstrap_brokers_tls"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bootstrap_brokers_public_tls", resourceName, "bootstrap_brokers_public_tls"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_name", resourceName, "cluster_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kafka_version", resourceName, "kafka_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "number_of_broker_nodes", resourceName, "number_of_broker_nodes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "zookeeper_connect_string", resourceName, "zookeeper_connect_string"),
				),
			},
			{
				Config: testAccClusterConfigClientAuthenticationTlsCertificateAuthorityArns(rName, commonName),
				Check: resource.ComposeTestCheckFunc(
					// CA must be DISABLED for deletion.
					acctest.CheckACMPCACertificateAuthorityDisableCA(&ca),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccClusterDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(testAccClusterBaseConfig(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.2.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.example_sg.id]
  }

  tags = {
    key1 = "value1"
  }
}

data "aws_msk_cluster" "test" {
  cluster_name = aws_msk_cluster.test.cluster_name
}
`, rName))
}

func testAccClusterDataSourceConfigPublicAccessSASLIam(rName string, pa string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseConfig(rName),
		testAccClusterConfigAllowEveryoneNoAclFoundFalse(rName),
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
    public_access {
      type = %[2]q
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

data "aws_msk_cluster" "test" {
  cluster_name = aws_msk_cluster.test.cluster_name
}
`, rName, pa))
}

func testAccClusterDataSourceConfigPublicAccessSASLScram(rName string, pa string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseConfig(rName),
		testAccClusterConfigAllowEveryoneNoAclFoundFalse(rName),
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
    public_access {
      type = %[2]q
    }
  }

  configuration_info {
    arn      = aws_msk_configuration.test.arn
    revision = aws_msk_configuration.test.latest_revision
  }

  client_authentication {
    sasl {
      scram = true
    }
  }
}

data "aws_msk_cluster" "test" {
  cluster_name = aws_msk_cluster.test.cluster_name
}
`, rName, pa))
}

func testAccClusterDataSourceConfigPublicAccessTLS(rName string, commonName string, pa string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseConfig(rName),
		testAccClusterConfigRootCA(commonName),
		testAccClusterConfigAllowEveryoneNoAclFoundFalse(rName),
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
    public_access {
      type = %[2]q
    }
  }

  configuration_info {
    arn      = aws_msk_configuration.test.arn
    revision = aws_msk_configuration.test.latest_revision
  }

  client_authentication {
    tls {
      certificate_authority_arns = [aws_acmpca_certificate_authority.test.arn]
    }
  }
}

data "aws_msk_cluster" "test" {
  cluster_name = aws_msk_cluster.test.cluster_name
}
`, rName, pa))
}
