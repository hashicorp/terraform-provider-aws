package eks_test

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfeks "github.com/hashicorp/terraform-provider-aws/internal/service/eks"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEKSCluster_basic(t *testing.T) {
	var cluster eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "eks", regexp.MustCompile(fmt.Sprintf("cluster/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_authority.0.data"),
					resource.TestMatchResourceAttr(resourceName, "endpoint", regexp.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "identity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity.0.oidc.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "identity.0.oidc.0.issuer", regexp.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "kubernetes_network_config.0.service_ipv4_cidr"),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.0.ip_family", "ipv4"),
					resource.TestMatchResourceAttr(resourceName, "platform_version", regexp.MustCompile(`^eks\.\d+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", eks.ClusterStatusActive),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "version", regexp.MustCompile(`^\d+\.\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_private_access", "false"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_public_access", "true"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "vpc_config.0.vpc_id", regexp.MustCompile(`^vpc-.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEKSCluster_disappears(t *testing.T) {
	var cluster eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					acctest.CheckResourceDisappears(acctest.Provider, tfeks.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEKSCluster_Encryption_create(t *testing.T) {
	var cluster eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.provider.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_config.0.provider.0.key_arn", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.resources.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEKSCluster_Encryption_update(t *testing.T) {
	var cluster1, cluster2 eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.#", "0"),
				),
			},
			{
				Config: testAccClusterConfig_encryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.provider.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_config.0.provider.0.key_arn", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.resources.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/19968.
func TestAccEKSCluster_Encryption_versionUpdate(t *testing.T) {
	var cluster1, cluster2 eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encryptionVersion(rName, "1.19"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.provider.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_config.0.provider.0.key_arn", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "version", "1.19"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_encryptionVersion(rName, "1.20"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.provider.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_config.0.provider.0.key_arn", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "encryption_config.0.resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "version", "1.20"),
				),
			},
		},
	})
}

func TestAccEKSCluster_version(t *testing.T) {
	var cluster1, cluster2 eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_version(rName, "1.19"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "version", "1.19"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_version(rName, "1.20"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "version", "1.20"),
				),
			},
		},
	})
}

func TestAccEKSCluster_logging(t *testing.T) {
	var cluster1, cluster2 eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_logging(rName, []string{"api"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "enabled_cluster_log_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cluster_log_types.*", "api"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_logging(rName, []string{"api", "audit"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "enabled_cluster_log_types.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cluster_log_types.*", "api"),
					resource.TestCheckTypeSetElemAttr(resourceName, "enabled_cluster_log_types.*", "audit"),
				),
			},
			// Disable all log types.
			{
				Config: testAccClusterConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "enabled_cluster_log_types.#", "0"),
				),
			},
		},
	})
}

func TestAccEKSCluster_tags(t *testing.T) {
	var cluster1, cluster2, cluster3 eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccClusterConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEKSCluster_VPC_securityGroupIDs(t *testing.T) {
	var cluster eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_vpcSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEKSCluster_VPC_endpointPrivateAccess(t *testing.T) {
	var cluster1, cluster2, cluster3 eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_vpcEndpointPrivateAccess(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_private_access", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_vpcEndpointPrivateAccess(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_private_access", "false"),
				),
			},
			{
				Config: testAccClusterConfig_vpcEndpointPrivateAccess(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster3),
					testAccCheckClusterNotRecreated(&cluster2, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_private_access", "true"),
				),
			},
		},
	})
}

func TestAccEKSCluster_VPC_endpointPublicAccess(t *testing.T) {
	var cluster1, cluster2, cluster3 eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_vpcEndpointPublicAccess(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_public_access", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_vpcEndpointPublicAccess(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_public_access", "true"),
				),
			},
			{
				Config: testAccClusterConfig_vpcEndpointPublicAccess(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster3),
					testAccCheckClusterNotRecreated(&cluster2, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.endpoint_public_access", "false"),
				),
			},
		},
	})
}

func TestAccEKSCluster_VPC_publicAccessCIDRs(t *testing.T) {
	var cluster eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_vpcPublicAccessCIDRs(rName, `["1.2.3.4/32", "5.6.7.8/32"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.public_access_cidrs.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_vpcPublicAccessCIDRs(rName, `["4.3.2.1/32", "8.7.6.5/32"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.public_access_cidrs.#", "2"),
				),
			},
		},
	})
}

func TestAccEKSCluster_Network_serviceIPv4CIDR(t *testing.T) {
	var cluster1, cluster2 eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_networkServiceIPv4CIDR(rName, `"10.0.0.0/11"`),
				ExpectError: regexp.MustCompile(`expected .* to contain a network Value with between`),
			},
			{
				Config:      testAccClusterConfig_networkServiceIPv4CIDR(rName, `"10.0.0.0/25"`),
				ExpectError: regexp.MustCompile(`expected .* to contain a network Value with between`),
			},
			{
				Config:      testAccClusterConfig_networkServiceIPv4CIDR(rName, `"9.0.0.0/16"`),
				ExpectError: regexp.MustCompile(`must be within`),
			},
			{
				Config:      testAccClusterConfig_networkServiceIPv4CIDR(rName, `"172.14.0.0/24"`),
				ExpectError: regexp.MustCompile(`must be within`),
			},
			{
				Config:      testAccClusterConfig_networkServiceIPv4CIDR(rName, `"192.167.0.0/24"`),
				ExpectError: regexp.MustCompile(`must be within`),
			},
			{
				Config: testAccClusterConfig_networkServiceIPv4CIDR(rName, `"192.168.0.0/24"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.0.service_ipv4_cidr", "192.168.0.0/24"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccClusterConfig_networkServiceIPv4CIDR(rName, `"192.168.0.0/24"`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccClusterConfig_networkServiceIPv4CIDR(rName, `"192.168.1.0/24"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.0.service_ipv4_cidr", "192.168.1.0/24"),
				),
			},
		},
	})
}

func TestAccEKSCluster_Network_ipFamily(t *testing.T) {
	var cluster1, cluster2 eks.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_networkIPFamily(rName, `"v6"`),
				ExpectError: regexp.MustCompile(`expected .* to be one of \[ipv4 ipv6]`),
			},
			{
				Config:      testAccClusterConfig_networkIPFamily(rName, `"IPv4"`),
				ExpectError: regexp.MustCompile(`expected .* to be one of \[ipv4 ipv6]`),
			},
			{
				Config: testAccClusterConfig_networkIPFamily(rName, `"ipv6"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.0.ip_family", "ipv6"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccClusterConfig_networkIPFamily(rName, `"ipv6"`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccClusterConfig_networkIPFamily(rName, `"ipv4"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kubernetes_network_config.0.ip_family", "ipv4"),
				),
			},
		},
	})
}

func testAccCheckClusterExists(resourceName string, cluster *eks.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EKS Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSConn

		output, err := tfeks.FindClusterByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*cluster = *output

		return nil
	}
}

func testAccCheckClusterDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eks_cluster" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EKSConn

		_, err := tfeks.FindClusterByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EKS Cluster %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckClusterRecreated(i, j *eks.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreatedAt).Equal(aws.TimeValue(j.CreatedAt)) {
			return errors.New("EKS Cluster was not recreated")
		}

		return nil
	}
}

func testAccCheckClusterNotRecreated(i, j *eks.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreatedAt).Equal(aws.TimeValue(j.CreatedAt)) {
			return errors.New("EKS Cluster was recreated")
		}

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EKSConn

	input := &eks.ListClustersInput{}

	_, err := conn.ListClusters(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccClusterConfig_Base(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.test.name
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)
  assign_ipv6_address_on_creation = true

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}
`, rName)
}

func testAccClusterConfig_required(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName))
}

func testAccClusterConfig_version(rName, version string) string {
	return acctest.ConfigCompose(testAccClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  version  = %[2]q

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, version))
}

func testAccClusterConfig_logging(rName string, logTypes []string) string {
	return acctest.ConfigCompose(testAccClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name                      = %[1]q
  role_arn                  = aws_iam_role.test.arn
  enabled_cluster_log_types = ["%[2]v"]

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, strings.Join(logTypes, "\", \"")))
}

func testAccClusterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccClusterConfig_encryption(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  encryption_config {
    resources = ["secrets"]

    provider {
      key_arn = aws_kms_key.test.arn
    }
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName))
}

func testAccClusterConfig_encryptionVersion(rName, version string) string {
	return acctest.ConfigCompose(testAccClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  version  = %[2]q

  encryption_config {
    resources = ["secrets"]

    provider {
      key_arn = aws_kms_key.test.arn
    }
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, version))
}

func testAccClusterConfig_vpcSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName))
}

func testAccClusterConfig_vpcEndpointPrivateAccess(rName string, endpointPrivateAccess bool) string {
	return acctest.ConfigCompose(testAccClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    endpoint_private_access = %[2]t
    endpoint_public_access  = true
    subnet_ids              = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, endpointPrivateAccess))
}

func testAccClusterConfig_vpcEndpointPublicAccess(rName string, endpointPublicAccess bool) string {
	return acctest.ConfigCompose(testAccClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    endpoint_private_access = true
    endpoint_public_access  = %[2]t
    subnet_ids              = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, endpointPublicAccess))
}

func testAccClusterConfig_vpcPublicAccessCIDRs(rName string, publicAccessCidr string) string {
	return acctest.ConfigCompose(testAccClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    endpoint_private_access = true
    endpoint_public_access  = true
    public_access_cidrs     = %[2]s
    subnet_ids              = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, publicAccessCidr))
}

func testAccClusterConfig_networkServiceIPv4CIDR(rName string, serviceIpv4Cidr string) string {
	return acctest.ConfigCompose(testAccClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  kubernetes_network_config {
    service_ipv4_cidr = %[2]s
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, serviceIpv4Cidr))
}

func testAccClusterConfig_networkIPFamily(rName string, ipFamily string) string {
	return acctest.ConfigCompose(testAccClusterConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  kubernetes_network_config {
    ip_family = %[2]s
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName, ipFamily))
}
