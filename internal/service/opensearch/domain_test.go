package opensearch_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestParseEngineVersion(t *testing.T) {
	testCases := []struct {
		TestName           string
		InputEngineVersion string
		ExpectError        bool
		ExpectedEngineType string
		ExpectedSemver     string
	}{
		{
			TestName:    "empty engine version",
			ExpectError: true,
		},
		{
			TestName:           "no separator",
			InputEngineVersion: "OpenSearch2.0",
			ExpectError:        true,
		},
		{
			TestName:           "too many separators",
			InputEngineVersion: "Open_Search_2.0",
			ExpectError:        true,
		},
		{
			TestName:           "valid",
			InputEngineVersion: "Elasticsearch_7.2",
			ExpectedEngineType: "Elasticsearch",
			ExpectedSemver:     "7.2",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			engineType, semver, err := tfopensearch.ParseEngineVersion(testCase.InputEngineVersion)

			if err == nil && testCase.ExpectError {
				t.Fatal("expected error, got no error")
			}

			if err != nil && !testCase.ExpectError {
				t.Fatalf("got unexpected error: %s", err)
			}

			if engineType != testCase.ExpectedEngineType {
				t.Errorf("engine type got %s, expected %s", engineType, testCase.ExpectedEngineType)
			}

			if semver != testCase.ExpectedSemver {
				t.Errorf("semver got %s, expected %s", semver, testCase.ExpectedSemver)
			}
		})
	}
}

func TestAccOpenSearchDomain_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "OpenSearch_1.1"),
					resource.TestMatchResourceAttr(resourceName, "kibana_endpoint", regexp.MustCompile(`.*(opensearch|es)\..*/_plugin/kibana/`)),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchDomain_requireHTTPS(t *testing.T) {
	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_endpointOptions(rName, true, "Policy-Min-TLS-1-0-2019-07"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists("aws_opensearch_domain.test", &domain),
					testAccCheckDomainEndpointOptions(true, "Policy-Min-TLS-1-0-2019-07", &domain),
				),
			},
			{
				ResourceName:      "aws_opensearch_domain.test",
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_endpointOptions(rName, true, "Policy-Min-TLS-1-2-2019-07"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists("aws_opensearch_domain.test", &domain),
					testAccCheckDomainEndpointOptions(true, "Policy-Min-TLS-1-2-2019-07", &domain),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_customEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"
	customEndpoint := fmt.Sprintf("%s.example.com", rName[:28])
	certResourceName := "aws_acm_certificate.test"
	certKey := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(certKey, customEndpoint)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_customEndpoint(rName, true, "Policy-Min-TLS-1-0-2019-07", true, customEndpoint, certKey, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "domain_endpoint_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "domain_endpoint_options.0.custom_endpoint_enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_endpoint_options.0.custom_endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_endpoint_options.0.custom_endpoint_certificate_arn", certResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_customEndpoint(rName, true, "Policy-Min-TLS-1-0-2019-07", true, customEndpoint, certKey, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckDomainEndpointOptions(true, "Policy-Min-TLS-1-0-2019-07", &domain),
					testAccCheckCustomEndpoint(resourceName, true, customEndpoint, &domain),
				),
			},
			{
				Config: testAccDomainConfig_customEndpoint(rName, true, "Policy-Min-TLS-1-0-2019-07", false, customEndpoint, certKey, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckDomainEndpointOptions(true, "Policy-Min-TLS-1-0-2019-07", &domain),
					testAccCheckCustomEndpoint(resourceName, false, customEndpoint, &domain),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_Cluster_zoneAwareness(t *testing.T) {
	var domain1, domain2, domain3, domain4 opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_clusterZoneAwarenessAZCount(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.0.availability_zone_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_clusterZoneAwarenessAZCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain2),
					testAccCheckDomainNotRecreated(&domain1, &domain2), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.0.availability_zone_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_enabled", "true"),
				),
			},
			{
				Config: testAccDomainConfig_clusterZoneAwarenessEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain3),
					testAccCheckDomainNotRecreated(&domain2, &domain3), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.#", "0"),
				),
			},
			{
				Config: testAccDomainConfig_clusterZoneAwarenessAZCount(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain4),
					testAccCheckDomainNotRecreated(&domain3, &domain4), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.0.availability_zone_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_enabled", "true"),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_Cluster_coldStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_clusterColdStorageOptions(rName, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cluster_config.0.cold_storage_options.*", map[string]string{
						"enabled": "false",
					})),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_clusterColdStorageOptions(rName, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cluster_config.0.cold_storage_options.*", map[string]string{
						"enabled": "true",
					})),
			},
		},
	})
}

func TestAccOpenSearchDomain_Cluster_warm(t *testing.T) {
	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_clusterWarm(rName, "ultrawarm1.medium.search", false, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", ""),
				),
			},
			{
				Config: testAccDomainConfig_clusterWarm(rName, "ultrawarm1.medium.search", true, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", "6"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", "ultrawarm1.medium.search"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_clusterWarm(rName, "ultrawarm1.medium.search", true, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", "7"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", "ultrawarm1.medium.search"),
				),
			},
			{
				Config: testAccDomainConfig_clusterWarm(rName, "ultrawarm1.large.search", true, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", "7"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", "ultrawarm1.large.search"),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_Cluster_dedicatedMaster(t *testing.T) {
	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_dedicatedClusterMaster(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_dedicatedClusterMaster(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
				),
			},
			{
				Config: testAccDomainConfig_dedicatedClusterMaster(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_Cluster_update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var input opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_clusterUpdate(rName, 2, 22),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &input),
					testAccCheckNumberOfInstances(2, &input),
					testAccCheckSnapshotHour(22, &input),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_clusterUpdate(rName, 4, 23),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &input),
					testAccCheckNumberOfInstances(4, &input),
					testAccCheckSnapshotHour(23, &input),
				),
			},
		}})
}

func TestAccOpenSearchDomain_duplicate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchConn
			_, err := conn.DeleteDomain(&opensearchservice.DeleteDomainInput{
				DomainName: aws.String(rName[:28]),
			})
			return err
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Create duplicate
					conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchConn
					_, err := conn.CreateDomain(&opensearchservice.CreateDomainInput{
						DomainName: aws.String(rName[:28]),
						EBSOptions: &opensearchservice.EBSOptions{
							EBSEnabled: aws.Bool(true),
							VolumeSize: aws.Int64(10),
						},
					})
					if err != nil {
						t.Fatal(err)
					}

					err = tfopensearch.WaitForDomainCreation(conn, rName[:28], 60*time.Minute)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "OpenSearch_1.1")),
				ExpectError: regexp.MustCompile(`domain .+ already exists`),
			},
		},
	})
}

func TestAccOpenSearchDomain_v23(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_v23(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(
						resourceName, "engine_version", "Elasticsearch_2.3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchDomain_complex(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_complex(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchDomain_VPC_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchDomain_VPC_update(t *testing.T) {
	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_vpcUpdate1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckNumberOfSecurityGroups(1, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_vpcUpdate2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckNumberOfSecurityGroups(2, &domain),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_VPC_internetToVPCEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_internetToVPCEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_autoTuneOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	autoTuneStartAtTime := testAccGetValidStartAtTime(t, "24h")
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_autoTuneOptions(rName, autoTuneStartAtTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "Elasticsearch_6.7"),
					resource.TestMatchResourceAttr(resourceName, "kibana_endpoint", regexp.MustCompile(`.*(opensearch|es)\..*/_plugin/kibana/`)),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.desired_state", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.0.start_at", autoTuneStartAtTime),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.0.duration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.0.duration.0.value", "2"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.0.duration.0.unit", "HOURS"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.0.cron_expression_for_recurrence", "cron(0 0 ? * 1 *)"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.rollback_on_disable", "NO_ROLLBACK"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchDomain_AdvancedSecurityOptions_userDB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_advancedSecurityOptionsUserDB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckAdvancedSecurityOptions(true, true, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
				// MasterUserOptions are not returned from DescribeDomainConfig
				ImportStateVerifyIgnore: []string{
					"advanced_security_options.0.internal_user_database_enabled",
					"advanced_security_options.0.master_user_options",
				},
			},
		},
	})
}

func TestAccOpenSearchDomain_AdvancedSecurityOptions_iam(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_advancedSecurityOptionsIAM(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckAdvancedSecurityOptions(true, false, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
				// MasterUserOptions are not returned from DescribeDomainConfig
				ImportStateVerifyIgnore: []string{
					"advanced_security_options.0.internal_user_database_enabled",
					"advanced_security_options.0.master_user_options",
				},
			},
		},
	})
}

func TestAccOpenSearchDomain_AdvancedSecurityOptions_disabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_advancedSecurityOptionsDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckAdvancedSecurityOptions(false, false, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
				// MasterUserOptions are not returned from DescribeDomainConfig
				ImportStateVerifyIgnore: []string{
					"advanced_security_options.0.internal_user_database_enabled",
					"advanced_security_options.0.master_user_options",
				},
			},
		},
	})
}

func TestAccOpenSearchDomain_LogPublishingOptions_indexSlowLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_logPublishingOptions(rName, opensearchservice.LogTypeIndexSlowLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": opensearchservice.LogTypeIndexSlowLogs,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchDomain_LogPublishingOptions_searchSlowLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_logPublishingOptions(rName, opensearchservice.LogTypeSearchSlowLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": opensearchservice.LogTypeSearchSlowLogs,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchDomain_LogPublishingOptions_applicationLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_logPublishingOptions(rName, opensearchservice.LogTypeEsApplicationLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": opensearchservice.LogTypeEsApplicationLogs,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchDomain_LogPublishingOptions_auditLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_logPublishingOptions(rName, opensearchservice.LogTypeAuditLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": opensearchservice.LogTypeAuditLogs,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
				// MasterUserOptions are not returned from DescribeDomainConfig
				ImportStateVerifyIgnore: []string{"advanced_security_options.0.master_user_options"},
			},
		},
	})
}

func TestAccOpenSearchDomain_CognitoOptions_createAndRemove(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckCognitoIdentityProvider(t)
			testAccPreCheckIAMServiceLinkedRole(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_cognitoOptions(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckCognitoOptions(true, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_cognitoOptions(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckCognitoOptions(false, &domain),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_CognitoOptions_update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckCognitoIdentityProvider(t)
			testAccPreCheckIAMServiceLinkedRole(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_cognitoOptions(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckCognitoOptions(false, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_cognitoOptions(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckCognitoOptions(true, &domain),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_Policy_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchDomain_Policy_ignoreEquivalent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_policyOrder(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
				),
			},
			{
				Config:   testAccDomainConfig_policyNewOrder(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccOpenSearchDomain_Encryption_atRestDefaultKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "Elasticsearch_6.0", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckDomainEncrypted(true, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchDomain_Encryption_atRestSpecifyKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_encryptAtRestKey(rName, "Elasticsearch_6.0", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckDomainEncrypted(true, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchDomain_Encryption_atRestEnable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "OpenSearch_1.1", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					testAccCheckDomainEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "OpenSearch_1.1", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain2),
					testAccCheckDomainEncrypted(true, &domain2),
					testAccCheckDomainNotRecreated(&domain1, &domain2), // note: this check does not work and always passes
				),
			},
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "OpenSearch_1.1", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					testAccCheckDomainEncrypted(false, &domain1),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_Encryption_atRestEnableLegacy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "Elasticsearch_5.6", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					testAccCheckDomainEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "Elasticsearch_5.6", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain2),
					testAccCheckDomainEncrypted(true, &domain2),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_Encryption_nodeToNode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "Elasticsearch_6.0", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckNodeToNodeEncrypted(true, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchDomain_Encryption_nodeToNodeEnable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "OpenSearch_1.1", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "OpenSearch_1.1", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain2),
					testAccCheckNodeToNodeEncrypted(true, &domain2),
					testAccCheckDomainNotRecreated(&domain1, &domain2), // note: this check does not work and always passes
				),
			},
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "OpenSearch_1.1", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_Encryption_nodeToNodeEnableLegacy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "Elasticsearch_6.0", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "Elasticsearch_6.0", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain2),
					testAccCheckNodeToNodeEncrypted(true, &domain2),
				),
			},
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "Elasticsearch_6.0", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDomainConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_VolumeType_update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var input opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_clusterUpdateEBSVolume(rName, 24),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &input),
					testAccCheckEBSVolumeEnabled(true, &input),
					testAccCheckEBSVolumeSize(24, &input),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_clusterUpdateInstanceStore(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &input),
					testAccCheckEBSVolumeEnabled(false, &input),
				),
			},
			{
				Config: testAccDomainConfig_clusterUpdateEBSVolume(rName, 12),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &input),
					testAccCheckEBSVolumeEnabled(true, &input),
					testAccCheckEBSVolumeSize(12, &input),
				),
			},
		}})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13867
func TestAccOpenSearchDomain_VolumeType_missing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_disabledEBSNullVolume(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.instance_type", "i3.xlarge.search"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.instance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.0.ebs_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.0.volume_size", "0"),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.0.volume_type", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchDomain_versionUpdate(t *testing.T) {
	var domain1, domain2, domain3 opensearchservice.DomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_clusterUpdateVersion(rName, "Elasticsearch_5.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "Elasticsearch_5.5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_clusterUpdateVersion(rName, "Elasticsearch_5.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain2),
					testAccCheckDomainNotRecreated(&domain1, &domain2), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, "engine_version", "Elasticsearch_5.6"),
				),
			},
			{
				Config: testAccDomainConfig_clusterUpdateVersion(rName, "Elasticsearch_6.3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain3),
					testAccCheckDomainNotRecreated(&domain2, &domain3), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, "engine_version", "Elasticsearch_6.3"),
				),
			},
		}})
}

func TestAccOpenSearchDomain_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIAMServiceLinkedRole(t) },
		ErrorCheck:        acctest.ErrorCheck(t, opensearchservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceDisappears(acctest.Provider, tfopensearch.ResourceDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainEndpointOptions(enforceHTTPS bool, tls string, status *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		options := status.DomainEndpointOptions
		if *options.EnforceHTTPS != enforceHTTPS {
			return fmt.Errorf("EnforceHTTPS differ. Given: %t, Expected: %t", *options.EnforceHTTPS, enforceHTTPS)
		}
		if *options.TLSSecurityPolicy != tls {
			return fmt.Errorf("TLSSecurityPolicy differ. Given: %s, Expected: %s", *options.TLSSecurityPolicy, tls)
		}
		return nil
	}
}

func testAccCheckCustomEndpoint(n string, customEndpointEnabled bool, customEndpoint string, status *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		options := status.DomainEndpointOptions
		if *options.CustomEndpointEnabled != customEndpointEnabled {
			return fmt.Errorf("CustomEndpointEnabled differ. Given: %t, Expected: %t", *options.CustomEndpointEnabled, customEndpointEnabled)
		}
		if *options.CustomEndpointEnabled {
			if *options.CustomEndpoint != customEndpoint {
				return fmt.Errorf("CustomEndpoint differ. Given: %s, Expected: %s", *options.CustomEndpoint, customEndpoint)
			}
			customEndpointCertificateArn := rs.Primary.Attributes["domain_endpoint_options.0.custom_endpoint_certificate_arn"]
			if *options.CustomEndpointCertificateArn != customEndpointCertificateArn {
				return fmt.Errorf("CustomEndpointCertificateArn differ. Given: %s, Expected: %s", *options.CustomEndpointCertificateArn, customEndpointCertificateArn)
			}
		}
		return nil
	}
}

func testAccCheckNumberOfSecurityGroups(numberOfSecurityGroups int, status *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		count := len(status.VPCOptions.SecurityGroupIds)
		if count != numberOfSecurityGroups {
			return fmt.Errorf("Number of security groups differ. Given: %d, Expected: %d", count, numberOfSecurityGroups)
		}
		return nil
	}
}

func testAccCheckEBSVolumeSize(ebsVolumeSize int, status *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.EBSOptions
		if *conf.VolumeSize != int64(ebsVolumeSize) {
			return fmt.Errorf("EBS volume size differ. Given: %d, Expected: %d", *conf.VolumeSize, ebsVolumeSize)
		}
		return nil
	}
}

func testAccCheckEBSVolumeEnabled(ebsEnabled bool, status *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.EBSOptions
		if *conf.EBSEnabled != ebsEnabled {
			return fmt.Errorf("EBS volume enabled. Given: %t, Expected: %t", *conf.EBSEnabled, ebsEnabled)
		}
		return nil
	}
}

func testAccCheckSnapshotHour(snapshotHour int, status *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.SnapshotOptions
		if *conf.AutomatedSnapshotStartHour != int64(snapshotHour) {
			return fmt.Errorf("Snapshots start hour differ. Given: %d, Expected: %d", *conf.AutomatedSnapshotStartHour, snapshotHour)
		}
		return nil
	}
}

func testAccCheckNumberOfInstances(numberOfInstances int, status *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.ClusterConfig
		if *conf.InstanceCount != int64(numberOfInstances) {
			return fmt.Errorf("Number of instances differ. Given: %d, Expected: %d", *conf.InstanceCount, numberOfInstances)
		}
		return nil
	}
}

func testAccCheckDomainEncrypted(encrypted bool, status *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.EncryptionAtRestOptions
		if aws.BoolValue(conf.Enabled) != encrypted {
			return fmt.Errorf("Encrypt at rest not set properly. Given: %t, Expected: %t", *conf.Enabled, encrypted)
		}
		return nil
	}
}

func testAccCheckNodeToNodeEncrypted(encrypted bool, status *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		options := status.NodeToNodeEncryptionOptions
		if aws.BoolValue(options.Enabled) != encrypted {
			return fmt.Errorf("Node-to-Node Encryption not set properly. Given: %t, Expected: %t", aws.BoolValue(options.Enabled), encrypted)
		}
		return nil
	}
}

func testAccCheckAdvancedSecurityOptions(enabled bool, userDbEnabled bool, status *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.AdvancedSecurityOptions

		if aws.BoolValue(conf.Enabled) != enabled {
			return fmt.Errorf(
				"AdvancedSecurityOptions.Enabled not set properly. Given: %t, Expected: %t",
				aws.BoolValue(conf.Enabled),
				enabled,
			)
		}

		if aws.BoolValue(conf.Enabled) {
			if aws.BoolValue(conf.InternalUserDatabaseEnabled) != userDbEnabled {
				return fmt.Errorf(
					"AdvancedSecurityOptions.InternalUserDatabaseEnabled not set properly. Given: %t, Expected: %t",
					aws.BoolValue(conf.InternalUserDatabaseEnabled),
					userDbEnabled,
				)
			}
		}

		return nil
	}
}

func testAccCheckCognitoOptions(enabled bool, status *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.CognitoOptions
		if *conf.Enabled != enabled {
			return fmt.Errorf("CognitoOptions not set properly. Given: %t, Expected: %t", *conf.Enabled, enabled)
		}
		return nil
	}
}

func testAccCheckDomainExists(n string, domain *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No OpenSearch Domain ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchConn
		resp, err := tfopensearch.FindDomainByName(conn, rs.Primary.Attributes["domain_name"])
		if err != nil {
			return fmt.Errorf("Error describing domain: %s", err.Error())
		}

		*domain = *resp

		return nil
	}
}

// testAccCheckDomainNotRecreated does not work. Inexplicably, a deleted
// domain's create time (& endpoint) carry over to a newly created domain with
// the same name, if it's created within any reasonable time after deletion.
// Also, domain ID is not unique and is simply the domain name so won't work
// for this check either.
func testAccCheckDomainNotRecreated(domain1, domain2 *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		/*
			conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchConn

			ic, err := conn.DescribeDomainConfig(&opensearchservice.DescribeDomainConfigInput{
				DomainName: domain1.DomainName,
			})
			if err != nil {
				return fmt.Errorf("while checking if domain (%s) was not recreated, describing domain config: %w", aws.StringValue(domain1.DomainName), err)
			}

			jc, err := conn.DescribeDomainConfig(&opensearchservice.DescribeDomainConfigInput{
				DomainName: domain2.DomainName,
			})
			if err != nil {
				return fmt.Errorf("while checking if domain (%s) was not recreated, describing domain config: %w", aws.StringValue(domain2.DomainName), err)
			}

			if aws.StringValue(domain1.Endpoint) != aws.StringValue(domain2.Endpoint) || !aws.TimeValue(ic.DomainConfig.ClusterConfig.Status.CreationDate).Equal(aws.TimeValue(jc.DomainConfig.ClusterConfig.Status.CreationDate)) {
				return fmt.Errorf("domain (%s) was recreated, before endpoint (%s, create time: %s), after endpoint (%s, create time: %s)",
					aws.StringValue(domain1.DomainName),
					aws.StringValue(domain1.Endpoint),
					aws.TimeValue(ic.DomainConfig.ClusterConfig.Status.CreationDate),
					aws.StringValue(domain2.Endpoint),
					aws.TimeValue(jc.DomainConfig.ClusterConfig.Status.CreationDate),
				)
			}
		*/

		return nil
	}
}

func testAccCheckDomainDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_opensearch_domain" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchConn
		_, err := tfopensearch.FindDomainByName(conn, rs.Primary.Attributes["domain_name"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("OpenSearch domain %s still exists", rs.Primary.ID)

	}
	return nil
}

func testAccGetValidStartAtTime(t *testing.T, timeUntilStart string) string {
	n := time.Now().UTC()
	d, err := time.ParseDuration(timeUntilStart)
	if err != nil {
		t.Fatalf("err parsing timeUntilStart: %s", err)
	}
	return n.Add(d).Format(time.RFC3339)
}

func testAccPreCheckIAMServiceLinkedRole(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
	dnsSuffix := acctest.Provider.Meta().(*conns.AWSClient).DNSSuffix

	input := &iam.ListRolesInput{
		PathPrefix: aws.String("/aws-service-role/opensearchservice."),
	}

	var role *iam.Role
	err := conn.ListRolesPages(input, func(page *iam.ListRolesOutput, lastPage bool) bool {
		for _, r := range page.Roles {
			if strings.HasPrefix(aws.StringValue(r.Path), "/aws-service-role/opensearchservice.") {
				role = r
			}
		}

		return !lastPage
	})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if role == nil {
		t.Fatalf("missing IAM Service Linked Role (opensearchservice.%s), please create it in the AWS account and retry", dnsSuffix)
	}
}

func testAccPreCheckCognitoIdentityProvider(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

	input := &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int64(1),
	}

	_, err := conn.ListUserPools(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckELBDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ELBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elb" {
			continue
		}

		describe, err := conn.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
			LoadBalancerNames: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			if len(describe.LoadBalancerDescriptions) != 0 &&
				*describe.LoadBalancerDescriptions[0].LoadBalancerName == rs.Primary.ID {
				return fmt.Errorf("ELB still exists")
			}
		}

		// Verify the error
		providerErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}

		if providerErr.Code() != elb.ErrCodeAccessPointNotFoundException {
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}

	return nil
}

func testAccDomainConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName)
}

func testAccDomainConfig_autoTuneOptions(rName, autoTuneStartAtTime string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = substr(%[1]q, 0, 28)
  engine_version = "Elasticsearch_6.7"

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  auto_tune_options {
    desired_state = "ENABLED"

    maintenance_schedule {
      start_at = %[2]q
      duration {
        value = "2"
        unit  = "HOURS"
      }
      cron_expression_for_recurrence = "cron(0 0 ? * 1 *)"
    }

    rollback_on_disable = "NO_ROLLBACK"

  }
}
`, rName, autoTuneStartAtTime)
}

func testAccDomainConfig_disabledEBSNullVolume(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = substr(%[1]q, 0, 28)
  engine_version = "Elasticsearch_6.0"

  cluster_config {
    instance_type  = "i3.xlarge.search"
    instance_count = 1
  }

  ebs_options {
    ebs_enabled = false
    volume_size = 0
    volume_type = null
  }
}
`, rName)
}

func testAccDomainConfig_endpointOptions(rName string, enforceHttps bool, tlsSecurityPolicy string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  domain_endpoint_options {
    enforce_https       = %[2]t
    tls_security_policy = %[3]q
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName, enforceHttps, tlsSecurityPolicy)
}

func testAccDomainConfig_customEndpoint(rName string, enforceHttps bool, tlsSecurityPolicy string, customEndpointEnabled bool, customEndpoint string, certKey string, certBody string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  private_key      = "%[6]s"
  certificate_body = "%[7]s"
}

resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  domain_endpoint_options {
    enforce_https                   = %[2]t
    tls_security_policy             = %[3]q
    custom_endpoint_enabled         = %[4]t
    custom_endpoint                 = "%[5]s"
    custom_endpoint_certificate_arn = aws_acm_certificate.test.arn
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName, enforceHttps, tlsSecurityPolicy, customEndpointEnabled, customEndpoint, acctest.TLSPEMEscapeNewlines(certKey), acctest.TLSPEMEscapeNewlines(certBody))
}

func testAccDomainConfig_clusterZoneAwarenessAZCount(rName string, availabilityZoneCount int) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = substr(%[1]q, 0, 28)
  engine_version = "Elasticsearch_1.5"

  cluster_config {
    instance_type          = "t2.small.search"
    instance_count         = 6
    zone_awareness_enabled = true

    zone_awareness_config {
      availability_zone_count = %[2]d
    }
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName, availabilityZoneCount)
}

func testAccDomainConfig_clusterColdStorageOptions(rName string, warmEnabled bool, csEnabled bool) string {
	warmConfig := ""
	if warmEnabled {
		warmConfig = `
	warm_count = "2"
	warm_type = "ultrawarm1.medium.search"
`
	}

	coldConfig := ""
	if csEnabled {
		coldConfig = `
	cold_storage_options {
	  enabled = true
	}
`
	}

	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = substr(%[1]q, 0, 28)
  engine_version = "Elasticsearch_7.9"

  cluster_config {
    zone_awareness_enabled   = true
    instance_type            = "c5.large.search"
    instance_count           = "3"
    dedicated_master_enabled = true
    dedicated_master_count   = "3"
    dedicated_master_type    = "c5.large.search"
    warm_enabled             = %[2]t

    %[3]s
    %[4]s

    zone_awareness_config {
      availability_zone_count = 3
    }
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName, warmEnabled, warmConfig, coldConfig)
}

func testAccDomainConfig_clusterZoneAwarenessEnabled(rName string, zoneAwarenessEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = substr(%[1]q, 0, 28)
  engine_version = "Elasticsearch_1.5"

  cluster_config {
    instance_type          = "t2.small.search"
    instance_count         = 6
    zone_awareness_enabled = %[2]t
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName, zoneAwarenessEnabled)
}

func testAccDomainConfig_clusterWarm(rName, warmType string, enabled bool, warmCnt int) string {
	warmConfig := ""
	if enabled {
		warmConfig = fmt.Sprintf(`
    warm_count = %[1]d
    warm_type = %[2]q
`, warmCnt, warmType)
	}

	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = substr(%[1]q, 0, 28)
  engine_version = "Elasticsearch_6.8"

  cluster_config {
    zone_awareness_enabled   = true
    instance_type            = "c5.large.search"
    instance_count           = "3"
    dedicated_master_enabled = true
    dedicated_master_count   = "3"
    dedicated_master_type    = "c5.large.search"
    warm_enabled             = %[2]t

    %[3]s

    zone_awareness_config {
      availability_zone_count = 3
    }
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName, enabled, warmConfig)
}

func testAccDomainConfig_dedicatedClusterMaster(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  cluster_config {
    instance_type            = "t2.small.search"
    instance_count           = "1"
    dedicated_master_enabled = %t
    dedicated_master_count   = "3"
    dedicated_master_type    = "t2.small.search"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName, enabled)
}

func testAccDomainConfig_clusterUpdate(rName string, instanceInt, snapshotInt int) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  advanced_options = {
    "indices.fielddata.cache.size" = 80
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = %d
    zone_awareness_enabled = true
    instance_type          = "t2.small.search"
  }

  snapshot_options {
    automated_snapshot_start_hour = %d
  }

  timeouts {
    update = "180m"
  }
}
`, rName, instanceInt, snapshotInt)
}

func testAccDomainConfig_clusterUpdateEBSVolume(rName string, volumeSize int) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  engine_version = "Elasticsearch_6.0"

  advanced_options = {
    "indices.fielddata.cache.size" = 80
  }

  ebs_options {
    ebs_enabled = true
    volume_size = %d
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.search"
  }
}
`, rName, volumeSize)
}

func testAccDomainConfig_clusterUpdateVersion(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  engine_version = %[2]q

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 1
    zone_awareness_enabled = false
    instance_type          = "t2.small.search"
  }
}
`, rName, version)
}

func testAccDomainConfig_clusterUpdateInstanceStore(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  engine_version = "Elasticsearch_6.0"

  advanced_options = {
    "indices.fielddata.cache.size" = 80
  }

  ebs_options {
    ebs_enabled = false
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "i3.large.search"
  }
}
`, rName)
}

func testAccDomainConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDomainConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)
  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDomainConfig_policy(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  access_policies = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = aws_iam_role.test.arn
      }
      Action   = "es:*"
      Resource = "arn:${data.aws_partition.current.partition}:es:*"
    }]
  })
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.${data.aws_partition.current.dns_suffix}"]
    }
  }
}
`, rName)
}

func testAccDomainConfig_policyOrder(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  access_policies = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = [
          aws_iam_role.test.arn,
          aws_iam_role.test2.arn,
        ]
      }
      Action   = "es:*"
      Resource = "arn:${data.aws_partition.current.partition}:es:*"
    }]
  })
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role" "test2" {
  name               = "%[1]s-2"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.${data.aws_partition.current.dns_suffix}"]
    }
  }
}
`, rName)
}

func testAccDomainConfig_policyNewOrder(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  access_policies = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = [
          aws_iam_role.test2.arn,
          aws_iam_role.test.arn,
        ]
      }
      Action   = "es:*"
      Resource = "arn:${data.aws_partition.current.partition}:es:*"
    }]
  })
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role" "test2" {
  name               = "%[1]s-2"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.${data.aws_partition.current.dns_suffix}"]
    }
  }
}
`, rName)
}

func testAccDomainConfig_encryptAtRestDefaultKey(rName, version string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  engine_version = %[2]q

  cluster_config {
    instance_type = "m4.large.search"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  encrypt_at_rest {
    enabled = %[3]t
  }
}
`, rName, version, enabled)
}

func testAccDomainConfig_encryptAtRestKey(rName, version string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  engine_version = %[2]q

  # Encrypt at rest requires m4/c4/r4/i2 instances. See http://docs.aws.amazon.com/opensearch-service/latest/developerguide/aes-supported-instance-types.html
  cluster_config {
    instance_type = "m4.large.search"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  encrypt_at_rest {
    enabled    = %[3]t
    kms_key_id = aws_kms_key.test.key_id
  }
}
`, rName, version, enabled)
}

func testAccDomainConfig_nodeToNodeEncryption(rName, version string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  engine_version = %[2]q

  cluster_config {
    instance_type = "m4.large.search"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  node_to_node_encryption {
    enabled = %[3]t
  }
}
`, rName, version, enabled)
}

func testAccDomainConfig_complex(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  advanced_options = {
    "indices.fielddata.cache.size" = 80
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.search"
  }

  snapshot_options {
    automated_snapshot_start_hour = 23
  }

  tags = {
    bar = "complex"
  }
}
`, rName)
}

func testAccDomainConfig_v23(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  engine_version = "Elasticsearch_2.3"
}
`, rName)
}

func testAccDomainConfig_vpc(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/22"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "192.168.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
}

resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.search"
  }

  vpc_options {
    security_group_ids = [aws_security_group.test.id, aws_security_group.test2.id]
    subnet_ids         = [aws_subnet.test.id, aws_subnet.test2.id]
  }
}
`, rName))
}

func testAccDomainConfig_vpcUpdate1(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/22"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "az1_first" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "az2_first" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "192.168.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "az1_second" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.2.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "az2_second" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "192.168.3.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
}

resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.search"
  }

  vpc_options {
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.az1_first.id, aws_subnet.az2_first.id]
  }
}
`, rName))
}

func testAccDomainConfig_vpcUpdate2(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/22"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "az1_first" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "az2_first" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "192.168.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "az1_second" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.2.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "az2_second" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "192.168.3.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
}

resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.search"
  }

  vpc_options {
    security_group_ids = [aws_security_group.test.id, aws_security_group.test2.id]
    subnet_ids         = [aws_subnet.az1_second.id, aws_subnet.az2_second.id]
  }
}
`, rName))
}

func testAccDomainConfig_internetToVPCEndpoint(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/22"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "192.168.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
}

resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.search"
  }

  vpc_options {
    security_group_ids = [aws_security_group.test.id, aws_security_group.test2.id]
    subnet_ids         = [aws_subnet.test.id, aws_subnet.test2.id]
  }
}
`, rName))
}

func testAccDomainConfig_advancedSecurityOptionsUserDB(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = substr(%[1]q, 0, 28)
  engine_version = "Elasticsearch_7.1"

  cluster_config {
    instance_type = "r5.large.search"
  }

  advanced_security_options {
    enabled                        = true
    internal_user_database_enabled = true
    master_user_options {
      master_user_name     = "testmasteruser"
      master_user_password = "Barbarbarbar1!"
    }
  }

  encrypt_at_rest {
    enabled = true
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }

  node_to_node_encryption {
    enabled = true
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName)
}

func testAccDomainConfig_advancedSecurityOptionsIAM(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_opensearch_domain" "test" {
  domain_name    = substr(%[1]q, 0, 28)
  engine_version = "Elasticsearch_7.1"

  cluster_config {
    instance_type = "r5.large.search"
  }

  advanced_security_options {
    enabled                        = true
    internal_user_database_enabled = false
    master_user_options {
      master_user_arn = aws_iam_user.test.arn
    }
  }

  encrypt_at_rest {
    enabled = true
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }

  node_to_node_encryption {
    enabled = true
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName)
}

func testAccDomainConfig_advancedSecurityOptionsDisabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = substr(%[1]q, 0, 28)
  engine_version = "Elasticsearch_7.1"

  cluster_config {
    instance_type = "r5.large.search"
  }

  advanced_security_options {
    enabled                        = false
    internal_user_database_enabled = true
    master_user_options {
      master_user_name     = "testmasteruser"
      master_user_password = "Barbarbarbar1!"
    }
  }

  encrypt_at_rest {
    enabled = true
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }

  node_to_node_encryption {
    enabled = true
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName)
}

func testAccDomain_logPublishingOptionsBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name = %[1]q

  policy_document = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = [
          "es.${data.aws_partition.current.dns_suffix}",
        ]
      }
      Action = [
        "logs:PutLogEvents",
        "logs:PutLogEventsBatch",
        "logs:CreateLogStream",
      ]
      Resource = "arn:${data.aws_partition.current.partition}:logs:*"
    }]
  })
}
`, rName)
}

func testAccDomainConfig_logPublishingOptions(rName, logType string) string {
	var auditLogsConfig string
	if logType == opensearchservice.LogTypeAuditLogs {
		auditLogsConfig = `
	  	advanced_security_options {
			enabled                        = true
			internal_user_database_enabled = true
			master_user_options {
			  master_user_name     = "testmasteruser"
			  master_user_password = "Barbarbarbar1!"
			}
	  	}
	
		domain_endpoint_options {
	  		enforce_https       = true
	  		tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
		}
	
		encrypt_at_rest {
			enabled = true
		}
	
		node_to_node_encryption {
			enabled = true
		}`
	}
	return acctest.ConfigCompose(testAccDomain_logPublishingOptionsBase(rName), fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = substr(%[1]q, 0, 28)
  engine_version = "Elasticsearch_7.1" # needed for ESApplication/Audit Log Types

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

    %[2]s

  log_publishing_options {
    log_type                 = %[3]q
    cloudwatch_log_group_arn = aws_cloudwatch_log_group.test.arn
  }
}
`, rName, auditLogsConfig, logType))
}

func testAccDomainConfig_cognitoOptions(rName string, includeCognitoOptions bool) string {
	var cognitoOptions string
	if includeCognitoOptions {
		cognitoOptions = `
		cognito_options {
			enabled          = true
			user_pool_id     = aws_cognito_user_pool.test.id
			identity_pool_id = aws_cognito_identity_pool.test.id
			role_arn         = aws_iam_role.test.arn
		}`
	} else {
		cognitoOptions = ""
	}

	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_domain" "test" {
  domain       = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_identity_pool" "test" {
  identity_pool_name               = %[1]q
  allow_unauthenticated_identities = false

  lifecycle {
    ignore_changes = [cognito_identity_providers]
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/service-role/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      type = "Service"
      identifiers = [
        "es.${data.aws_partition.current.dns_suffix}",
      ]
    }
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonOpenSearchServiceCognitoAccess"
}

resource "aws_opensearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  engine_version = "OpenSearch_1.1"

  %[2]s

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  depends_on = [
    aws_cognito_user_pool_domain.test,
    aws_iam_role_policy_attachment.test,
  ]
}
`, rName, cognitoOptions)
}
