package elasticsearch_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticsearch "github.com/hashicorp/terraform-provider-aws/internal/service/elasticsearch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccElasticsearchDomain_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_version", "1.5"),
					resource.TestMatchResourceAttr(resourceName, "kibana_endpoint", regexp.MustCompile(`.*es\..*/_plugin/kibana/`)),
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

func TestAccElasticsearchDomain_requireHTTPS(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_DomainEndpointOptions(rName, true, "Policy-Min-TLS-1-0-2019-07"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists("aws_elasticsearch_domain.test", &domain),
					testAccCheckDomainEndpointOptions(true, "Policy-Min-TLS-1-0-2019-07", &domain),
				),
			},
			{
				ResourceName:      "aws_elasticsearch_domain.test",
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_DomainEndpointOptions(rName, true, "Policy-Min-TLS-1-2-2019-07"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists("aws_elasticsearch_domain.test", &domain),
					testAccCheckDomainEndpointOptions(true, "Policy-Min-TLS-1-2-2019-07", &domain),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_customEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"
	customEndpoint := fmt.Sprintf("%s.example.com", rName[:28])
	certResourceName := "aws_acm_certificate.test"
	certKey := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(certKey, customEndpoint)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_CustomEndpoint(rName, true, "Policy-Min-TLS-1-0-2019-07", true, customEndpoint, certKey, certificate),
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
				Config: testAccDomainConfig_CustomEndpoint(rName, true, "Policy-Min-TLS-1-0-2019-07", true, customEndpoint, certKey, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckDomainEndpointOptions(true, "Policy-Min-TLS-1-0-2019-07", &domain),
					testAccCheckCustomEndpoint(resourceName, true, customEndpoint, &domain),
				),
			},
			{
				Config: testAccDomainConfig_CustomEndpoint(rName, true, "Policy-Min-TLS-1-0-2019-07", false, customEndpoint, certKey, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckDomainEndpointOptions(true, "Policy-Min-TLS-1-0-2019-07", &domain),
					testAccCheckCustomEndpoint(resourceName, false, customEndpoint, &domain),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_Cluster_zoneAwareness(t *testing.T) {
	var domain1, domain2, domain3, domain4 elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_ClusterConfig_ZoneAwarenessConfig_AvailabilityZoneCount(rName, 3),
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
				Config: testAccDomainConfig_ClusterConfig_ZoneAwarenessConfig_AvailabilityZoneCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain2),
					testAccCheckDomainNotRecreated(&domain1, &domain2), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.0.availability_zone_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_enabled", "true"),
				),
			},
			{
				Config: testAccDomainConfig_ClusterConfig_ZoneAwarenessEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain3),
					testAccCheckDomainNotRecreated(&domain2, &domain3), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.#", "0"),
				),
			},
			{
				Config: testAccDomainConfig_ClusterConfig_ZoneAwarenessConfig_AvailabilityZoneCount(rName, 3),
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

func TestAccElasticsearchDomain_warm(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigWarm(rName, "ultrawarm1.medium.elasticsearch", false, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", ""),
				),
			},
			{
				Config: testAccDomainConfigWarm(rName, "ultrawarm1.medium.elasticsearch", true, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", "6"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", "ultrawarm1.medium.elasticsearch"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfigWarm(rName, "ultrawarm1.medium.elasticsearch", true, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", "7"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", "ultrawarm1.medium.elasticsearch"),
				),
			},
			{
				Config: testAccDomainConfigWarm(rName, "ultrawarm1.large.elasticsearch", true, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", "7"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", "ultrawarm1.large.elasticsearch"),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_withColdStorageOptions(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_WithColdStorageOptions(rName, false, false, false),
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
				Config: testAccDomainConfig_WithColdStorageOptions(rName, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cluster_config.0.cold_storage_options.*", map[string]string{
						"enabled": "true",
					})),
			},
		},
	})
}

func TestAccElasticsearchDomain_withDedicatedMaster(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_WithDedicatedClusterMaster(rName, false),
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
				Config: testAccDomainConfig_WithDedicatedClusterMaster(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
				),
			},
			{
				Config: testAccDomainConfig_WithDedicatedClusterMaster(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_duplicate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn
			_, err := conn.DeleteElasticsearchDomain(&elasticsearch.DeleteElasticsearchDomainInput{
				DomainName: aws.String(rName[:28]),
			})
			return err
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Create duplicate
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn
					_, err := conn.CreateElasticsearchDomain(&elasticsearch.CreateElasticsearchDomainInput{
						DomainName: aws.String(rName[:28]),
						EBSOptions: &elasticsearch.EBSOptions{
							EBSEnabled: aws.Bool(true),
							VolumeSize: aws.Int64(10),
						},
					})
					if err != nil {
						t.Fatal(err)
					}

					err = tfelasticsearch.WaitForDomainCreation(conn, rName[:28], 60*time.Minute)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccDomainConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(
						resourceName, "elasticsearch_version", "1.5"),
				),
				ExpectError: regexp.MustCompile(`Elasticsearch Domain .+ already exists`),
			},
		},
	})
}

func TestAccElasticsearchDomain_v23(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigV23(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(
						resourceName, "elasticsearch_version", "2.3"),
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

func TestAccElasticsearchDomain_complex(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
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

func TestAccElasticsearchDomain_vpc(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
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

func TestAccElasticsearchDomain_VPC_update(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_vpc_update1(rName),
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
				Config: testAccDomainConfig_vpc_update2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckNumberOfSecurityGroups(2, &domain),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_internetToVPCEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig(rName),
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
				Config: testAccDomainConfig_internetToVpcEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_AutoTuneOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	autoTuneStartAtTime := testAccGetValidStartAtTime(t, "24h")
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigWithAutoTuneOptions(rName, autoTuneStartAtTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(
						resourceName, "elasticsearch_version", "6.7"),
					resource.TestMatchResourceAttr(resourceName, "kibana_endpoint", regexp.MustCompile(`.*es\..*/_plugin/kibana/`)),
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

func TestAccElasticsearchDomain_AdvancedSecurityOptions_userDB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_AdvancedSecurityOptionsUserDb(rName),
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
				// MasterUserOptions are not returned from DescribeElasticsearchDomainConfig
				ImportStateVerifyIgnore: []string{
					"advanced_security_options.0.internal_user_database_enabled",
					"advanced_security_options.0.master_user_options",
				},
			},
		},
	})
}

func TestAccElasticsearchDomain_AdvancedSecurityOptions_iam(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_AdvancedSecurityOptionsIAM(rName),
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
				// MasterUserOptions are not returned from DescribeElasticsearchDomainConfig
				ImportStateVerifyIgnore: []string{
					"advanced_security_options.0.internal_user_database_enabled",
					"advanced_security_options.0.master_user_options",
				},
			},
		},
	})
}

func TestAccElasticsearchDomain_AdvancedSecurityOptions_disabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_AdvancedSecurityOptionsDisabled(rName),
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
				// MasterUserOptions are not returned from DescribeElasticsearchDomainConfig
				ImportStateVerifyIgnore: []string{
					"advanced_security_options.0.internal_user_database_enabled",
					"advanced_security_options.0.master_user_options",
				},
			},
		},
	})
}

func TestAccElasticsearchDomain_LogPublishingOptions_indexSlowLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_LogPublishingOptions(rName, elasticsearch.LogTypeIndexSlowLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": elasticsearch.LogTypeIndexSlowLogs,
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

func TestAccElasticsearchDomain_LogPublishingOptions_searchSlowLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_LogPublishingOptions(rName, elasticsearch.LogTypeSearchSlowLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": elasticsearch.LogTypeSearchSlowLogs,
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

func TestAccElasticsearchDomain_LogPublishingOptions_esApplicationLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_LogPublishingOptions(rName, elasticsearch.LogTypeEsApplicationLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": elasticsearch.LogTypeEsApplicationLogs,
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

func TestAccElasticsearchDomain_LogPublishingOptions_auditLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_LogPublishingOptions(rName, elasticsearch.LogTypeAuditLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": elasticsearch.LogTypeAuditLogs,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
				// MasterUserOptions are not returned from DescribeElasticsearchDomainConfig
				ImportStateVerifyIgnore: []string{"advanced_security_options.0.master_user_options"},
			},
		},
	})
}

func TestAccElasticsearchDomain_cognitoOptionsCreateAndRemove(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckCognitoIdentityProvider(t)
			testAccPreCheckIamServiceLinkedRoleEs(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_CognitoOptions(rName, true),
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
				Config: testAccDomainConfig_CognitoOptions(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckCognitoOptions(false, &domain),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_cognitoOptionsUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckCognitoIdentityProvider(t)
			testAccPreCheckIamServiceLinkedRoleEs(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_CognitoOptions(rName, false),
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
				Config: testAccDomainConfig_CognitoOptions(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					testAccCheckCognitoOptions(true, &domain),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_policy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigWithPolicy(rName),
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

func TestAccElasticsearchDomain_policyIgnoreEquivalent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainPolicyOrderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
				),
			},
			{
				Config:   testAccDomainPolicyNewOrderConfig(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccElasticsearchDomain_Encryption_atRestDefaultKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigWithEncryptAtRestDefaultKey(rName, "6.0", true),
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

func TestAccElasticsearchDomain_Encryption_atRestSpecifyKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigWithEncryptAtRestWithKey(rName, "6.0", true),
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

func TestAccElasticsearchDomain_Encryption_atRestEnable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigWithEncryptAtRestDefaultKey(rName, "6.7", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					testAccCheckDomainEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfigWithEncryptAtRestDefaultKey(rName, "6.7", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain2),
					testAccCheckDomainEncrypted(true, &domain2),
					testAccCheckDomainNotRecreated(&domain1, &domain2), // note: this check does not work and always passes
				),
			},
			{
				Config: testAccDomainConfigWithEncryptAtRestDefaultKey(rName, "6.7", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					testAccCheckDomainEncrypted(false, &domain1),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_Encryption_atRestEnableLegacy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigWithEncryptAtRestDefaultKey(rName, "5.6", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					testAccCheckDomainEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfigWithEncryptAtRestDefaultKey(rName, "5.6", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain2),
					testAccCheckDomainEncrypted(true, &domain2),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_Encryption_nodeToNode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigwithNodeToNodeEncryption(rName, "6.0", true),
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

func TestAccElasticsearchDomain_Encryption_nodeToNodeEnable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigwithNodeToNodeEncryption(rName, "6.7", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfigwithNodeToNodeEncryption(rName, "6.7", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain2),
					testAccCheckNodeToNodeEncrypted(true, &domain2),
					testAccCheckDomainNotRecreated(&domain1, &domain2), // note: this check does not work and always passes
				),
			},
			{
				Config: testAccDomainConfigwithNodeToNodeEncryption(rName, "6.7", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_Encryption_nodeToNodeEnableLegacy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigwithNodeToNodeEncryption(rName, "6.0", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfigwithNodeToNodeEncryption(rName, "6.0", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain2),
					testAccCheckNodeToNodeEncrypted(true, &domain2),
				),
			},
			{
				Config: testAccDomainConfigwithNodeToNodeEncryption(rName, "6.0", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigTags1(rName, "key1", "value1"),
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
				Config: testAccDomainConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDomainConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var input elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_ClusterUpdate(rName, 2, 22),
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
				Config: testAccDomainConfig_ClusterUpdate(rName, 4, 23),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &input),
					testAccCheckNumberOfInstances(4, &input),
					testAccCheckSnapshotHour(23, &input),
				),
			},
		}})
}

func TestAccElasticsearchDomain_UpdateVolume_type(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var input elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_ClusterUpdateEBSVolume(rName, 24),
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
				Config: testAccDomainConfig_ClusterUpdateInstanceStore(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &input),
					testAccCheckEBSVolumeEnabled(false, &input),
				),
			},
			{
				Config: testAccDomainConfig_ClusterUpdateEBSVolume(rName, 12),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &input),
					testAccCheckEBSVolumeEnabled(true, &input),
					testAccCheckEBSVolumeSize(12, &input),
				),
			},
		}})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13867
func TestAccElasticsearchDomain_WithVolumeType_missing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfigWithDisabledEBSNullVolumeType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.instance_type", "i3.xlarge.elasticsearch"),
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

func TestAccElasticsearchDomain_Update_version(t *testing.T) {
	var domain1, domain2, domain3 elasticsearch.ElasticsearchDomainStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_ClusterUpdateVersion(rName, "5.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain1),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_version", "5.5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName[:28],
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_ClusterUpdateVersion(rName, "5.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain2),
					testAccCheckDomainNotRecreated(&domain1, &domain2), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_version", "5.6"),
				),
			},
			{
				Config: testAccDomainConfig_ClusterUpdateVersion(rName, "6.3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(resourceName, &domain3),
					testAccCheckDomainNotRecreated(&domain2, &domain3), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_version", "6.3"),
				),
			},
		}})
}

func TestAccElasticsearchDomain_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticsearch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceDisappears(acctest.Provider, tfelasticsearch.ResourceDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainEndpointOptions(enforceHTTPS bool, tls string, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
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

func testAccCheckCustomEndpoint(n string, customEndpointEnabled bool, customEndpoint string, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
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

func testAccCheckNumberOfSecurityGroups(numberOfSecurityGroups int, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		count := len(status.VPCOptions.SecurityGroupIds)
		if count != numberOfSecurityGroups {
			return fmt.Errorf("Number of security groups differ. Given: %d, Expected: %d", count, numberOfSecurityGroups)
		}
		return nil
	}
}

func testAccCheckEBSVolumeSize(ebsVolumeSize int, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.EBSOptions
		if *conf.VolumeSize != int64(ebsVolumeSize) {
			return fmt.Errorf("EBS volume size differ. Given: %d, Expected: %d", *conf.VolumeSize, ebsVolumeSize)
		}
		return nil
	}
}

func testAccCheckEBSVolumeEnabled(ebsEnabled bool, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.EBSOptions
		if *conf.EBSEnabled != ebsEnabled {
			return fmt.Errorf("EBS volume enabled. Given: %t, Expected: %t", *conf.EBSEnabled, ebsEnabled)
		}
		return nil
	}
}

func testAccCheckSnapshotHour(snapshotHour int, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.SnapshotOptions
		if *conf.AutomatedSnapshotStartHour != int64(snapshotHour) {
			return fmt.Errorf("Snapshots start hour differ. Given: %d, Expected: %d", *conf.AutomatedSnapshotStartHour, snapshotHour)
		}
		return nil
	}
}

func testAccCheckNumberOfInstances(numberOfInstances int, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.ElasticsearchClusterConfig
		if *conf.InstanceCount != int64(numberOfInstances) {
			return fmt.Errorf("Number of instances differ. Given: %d, Expected: %d", *conf.InstanceCount, numberOfInstances)
		}
		return nil
	}
}

func testAccCheckDomainEncrypted(encrypted bool, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.EncryptionAtRestOptions
		if *conf.Enabled != encrypted {
			return fmt.Errorf("Encrypt at rest not set properly. Given: %t, Expected: %t", *conf.Enabled, encrypted)
		}
		return nil
	}
}

func testAccCheckNodeToNodeEncrypted(encrypted bool, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		options := status.NodeToNodeEncryptionOptions
		if aws.BoolValue(options.Enabled) != encrypted {
			return fmt.Errorf("Node-to-Node Encryption not set properly. Given: %t, Expected: %t", aws.BoolValue(options.Enabled), encrypted)
		}
		return nil
	}
}

func testAccCheckAdvancedSecurityOptions(enabled bool, userDbEnabled bool, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
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

func testAccCheckCognitoOptions(enabled bool, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.CognitoOptions
		if *conf.Enabled != enabled {
			return fmt.Errorf("CognitoOptions not set properly. Given: %t, Expected: %t", *conf.Enabled, enabled)
		}
		return nil
	}
}

func testAccCheckDomainExists(n string, domain *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ES Domain ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn
		resp, err := tfelasticsearch.FindDomainByName(conn, rs.Primary.Attributes["domain_name"])
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
func testAccCheckDomainNotRecreated(domain1, domain2 *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		/*
			conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn

			ic, err := conn.DescribeElasticsearchDomainConfig(&elasticsearch.DescribeElasticsearchDomainConfigInput{
				DomainName: domain1.DomainName,
			})
			if err != nil {
				return fmt.Errorf("while checking if domain (%s) was not recreated, describing domain config: %w", aws.StringValue(domain1.DomainName), err)
			}

			jc, err := conn.DescribeElasticsearchDomainConfig(&elasticsearch.DescribeElasticsearchDomainConfigInput{
				DomainName: domain2.DomainName,
			})
			if err != nil {
				return fmt.Errorf("while checking if domain (%s) was not recreated, describing domain config: %w", aws.StringValue(domain2.DomainName), err)
			}

			if aws.StringValue(domain1.Endpoint) != aws.StringValue(domain2.Endpoint) || !aws.TimeValue(ic.DomainConfig.ElasticsearchClusterConfig.Status.CreationDate).Equal(aws.TimeValue(jc.DomainConfig.ElasticsearchClusterConfig.Status.CreationDate)) {
				return fmt.Errorf("domain (%s) was recreated, before endpoint (%s, create time: %s), after endpoint (%s, create time: %s)",
					aws.StringValue(domain1.DomainName),
					aws.StringValue(domain1.Endpoint),
					aws.TimeValue(ic.DomainConfig.ElasticsearchClusterConfig.Status.CreationDate),
					aws.StringValue(domain2.Endpoint),
					aws.TimeValue(jc.DomainConfig.ElasticsearchClusterConfig.Status.CreationDate),
				)
			}
		*/

		return nil
	}
}

func testAccCheckDomainDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticsearch_domain" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn
		_, err := tfelasticsearch.FindDomainByName(conn, rs.Primary.Attributes["domain_name"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Elasticsearch domain %s still exists", rs.Primary.ID)

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

func testAccPreCheckIamServiceLinkedRoleEs(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
	dnsSuffix := acctest.Provider.Meta().(*conns.AWSClient).DNSSuffix

	input := &iam.ListRolesInput{
		PathPrefix: aws.String("/aws-service-role/es."),
	}

	var role *iam.Role
	err := conn.ListRolesPages(input, func(page *iam.ListRolesOutput, lastPage bool) bool {
		for _, r := range page.Roles {
			if strings.HasPrefix(aws.StringValue(r.Path), "/aws-service-role/es.") {
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
		t.Fatalf("missing IAM Service Linked Role (es.%s), please create it in the AWS account and retry", dnsSuffix)
	}
}

func testAccDomainConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName)
}

func testAccDomainConfigWithAutoTuneOptions(rName, autoTuneStartAtTime string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name           = substr(%[1]q, 0, 28)
  elasticsearch_version = "6.7"

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

func testAccDomainConfigWithDisabledEBSNullVolumeType(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name           = substr(%[1]q, 0, 28)
  elasticsearch_version = "6.0"

  cluster_config {
    instance_type  = "i3.xlarge.elasticsearch"
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

func testAccDomainConfig_DomainEndpointOptions(rName string, enforceHttps bool, tlsSecurityPolicy string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
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

func testAccDomainConfig_CustomEndpoint(rName string, enforceHttps bool, tlsSecurityPolicy string, customEndpointEnabled bool, customEndpoint string, certKey string, certBody string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  private_key      = "%[6]s"
  certificate_body = "%[7]s"
}

resource "aws_elasticsearch_domain" "test" {
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

func testAccDomainConfig_ClusterConfig_ZoneAwarenessConfig_AvailabilityZoneCount(rName string, availabilityZoneCount int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  cluster_config {
    instance_type          = "t2.small.elasticsearch"
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

func testAccDomainConfig_ClusterConfig_ZoneAwarenessEnabled(rName string, zoneAwarenessEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  cluster_config {
    instance_type          = "t2.small.elasticsearch"
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

func testAccDomainConfigWarm(rName, warmType string, enabled bool, warmCnt int) string {
	warmConfig := ""
	if enabled {
		warmConfig = fmt.Sprintf(`
    warm_count = %[1]d
    warm_type = %[2]q
`, warmCnt, warmType)
	}

	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name           = substr(%[1]q, 0, 28)
  elasticsearch_version = "6.8"

  cluster_config {
    zone_awareness_enabled   = true
    instance_type            = "c5.large.elasticsearch"
    instance_count           = "3"
    dedicated_master_enabled = true
    dedicated_master_count   = "3"
    dedicated_master_type    = "c5.large.elasticsearch"
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

func testAccDomainConfig_WithDedicatedClusterMaster(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  cluster_config {
    instance_type            = "t2.small.elasticsearch"
    instance_count           = "1"
    dedicated_master_enabled = %t
    dedicated_master_count   = "3"
    dedicated_master_type    = "t2.small.elasticsearch"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName, enabled)
}

func testAccDomainConfig_WithColdStorageOptions(rName string, dMasterEnabled bool, warmEnabled bool, csEnabled bool) string {
	warmConfig := ""
	if warmEnabled {
		warmConfig = `
	warm_count = "2"
	warm_type = "ultrawarm1.medium.elasticsearch"
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
resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  elasticsearch_version = "7.9"

  cluster_config {
    instance_type            = "m3.medium.elasticsearch"
    instance_count           = "1"
    dedicated_master_enabled = %t
    dedicated_master_count   = "3"
    dedicated_master_type    = "m3.medium.elasticsearch"
    warm_enabled             = %[3]t
    %[4]s
    %[5]s
  }
  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
  timeouts {
    update = "180m"
  }
}
`, rName, dMasterEnabled, warmEnabled, warmConfig, coldConfig)
}

func testAccDomainConfig_ClusterUpdate(rName string, instanceInt, snapshotInt int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
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
    instance_type          = "t2.small.elasticsearch"
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

func testAccDomainConfig_ClusterUpdateEBSVolume(rName string, volumeSize int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  elasticsearch_version = "6.0"

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
    instance_type          = "t2.small.elasticsearch"
  }
}
`, rName, volumeSize)
}

func testAccDomainConfig_ClusterUpdateVersion(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  elasticsearch_version = "%v"

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 1
    zone_awareness_enabled = false
    instance_type          = "t2.small.elasticsearch"
  }
}
`, rName, version)
}

func testAccDomainConfig_ClusterUpdateInstanceStore(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  elasticsearch_version = "6.0"

  advanced_options = {
    "indices.fielddata.cache.size" = 80
  }

  ebs_options {
    ebs_enabled = false
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "i3.large.elasticsearch"
  }
}
`, rName)
}

func testAccDomainConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
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

func testAccDomainConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
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

func testAccDomainConfigWithPolicy(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_elasticsearch_domain" "test" {
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

func testAccDomainPolicyOrderConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_elasticsearch_domain" "test" {
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

func testAccDomainPolicyNewOrderConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_elasticsearch_domain" "test" {
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

func testAccDomainConfigWithEncryptAtRestDefaultKey(rName, version string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  elasticsearch_version = %[2]q

  # Encrypt at rest requires m4/c4/r4/i2 instances. See http://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/aes-supported-instance-types.html
  cluster_config {
    instance_type = "m4.large.elasticsearch"
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

func testAccDomainConfigWithEncryptAtRestWithKey(rName, version string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  elasticsearch_version = %[2]q

  # Encrypt at rest requires m4/c4/r4/i2 instances. See http://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/aes-supported-instance-types.html
  cluster_config {
    instance_type = "m4.large.elasticsearch"
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

func testAccDomainConfigwithNodeToNodeEncryption(rName, version string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  elasticsearch_version = %[2]q

  cluster_config {
    instance_type = "m4.large.elasticsearch"
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
resource "aws_elasticsearch_domain" "test" {
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
    instance_type          = "t2.small.elasticsearch"
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

func testAccDomainConfigV23(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  elasticsearch_version = "2.3"
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

resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.elasticsearch"
  }

  vpc_options {
    security_group_ids = [aws_security_group.test.id, aws_security_group.test2.id]
    subnet_ids         = [aws_subnet.test.id, aws_subnet.test2.id]
  }
}
`, rName))
}

func testAccDomainConfig_vpc_update1(rName string) string {
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

resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.elasticsearch"
  }

  vpc_options {
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.az1_first.id, aws_subnet.az2_first.id]
  }
}
`, rName))
}

func testAccDomainConfig_vpc_update2(rName string) string {
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

resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.elasticsearch"
  }

  vpc_options {
    security_group_ids = [aws_security_group.test.id, aws_security_group.test2.id]
    subnet_ids         = [aws_subnet.az1_second.id, aws_subnet.az2_second.id]
  }
}
`, rName))
}

func testAccDomainConfig_internetToVpcEndpoint(rName string) string {
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

resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.elasticsearch"
  }

  vpc_options {
    security_group_ids = [aws_security_group.test.id, aws_security_group.test2.id]
    subnet_ids         = [aws_subnet.test.id, aws_subnet.test2.id]
  }
}
`, rName))
}

func testAccDomainConfig_AdvancedSecurityOptionsUserDb(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name           = substr(%[1]q, 0, 28)
  elasticsearch_version = "7.1"

  cluster_config {
    instance_type = "r5.large.elasticsearch"
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

func testAccDomainConfig_AdvancedSecurityOptionsIAM(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_elasticsearch_domain" "test" {
  domain_name           = substr(%[1]q, 0, 28)
  elasticsearch_version = "7.1"

  cluster_config {
    instance_type = "r5.large.elasticsearch"
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

func testAccDomainConfig_AdvancedSecurityOptionsDisabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name           = substr(%[1]q, 0, 28)
  elasticsearch_version = "7.1"

  cluster_config {
    instance_type = "r5.large.elasticsearch"
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

func testAccDomain_LogPublishingOptions_BaseConfig(rName string) string {
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
        Service = "es.${data.aws_partition.current.dns_suffix}"
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

func testAccDomainConfig_LogPublishingOptions(rName, logType string) string {
	var auditLogsConfig string
	if logType == elasticsearch.LogTypeAuditLogs {
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
	return acctest.ConfigCompose(testAccDomain_LogPublishingOptions_BaseConfig(rName), fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name           = substr(%[1]q, 0, 28)
  elasticsearch_version = "7.1" # needed for ESApplication/Audit Log Types

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

func testAccDomainConfig_CognitoOptions(rName string, includeCognitoOptions bool) string {
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
    sid     = ""
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      type        = "Service"
      identifiers = ["es.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonESCognitoAccess"
}

resource "aws_elasticsearch_domain" "test" {
  domain_name = substr(%[1]q, 0, 28)

  elasticsearch_version = "6.0"

  %[2]s

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  depends_on = [
    aws_iam_role.test,
    aws_iam_role_policy_attachment.test,
  ]
}
`, rName, cognitoOptions)
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
