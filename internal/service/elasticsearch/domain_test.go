// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticsearch "github.com/hashicorp/terraform-provider-aws/internal/service/elasticsearch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestEBSVolumeTypePermitsIopsInput(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		volumeType string
		want       bool
	}{
		{"empty", "", false},
		{"gp2", elasticsearch.VolumeTypeGp2, false},
		{"gp3", elasticsearch.VolumeTypeGp3, true},
		{"io1", elasticsearch.VolumeTypeIo1, true},
		{"standard", elasticsearch.VolumeTypeStandard, false},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := tfelasticsearch.EBSVolumeTypePermitsIopsInput(testCase.volumeType); got != testCase.want {
				t.Errorf("EBSVolumeTypePermitsIopsInput() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestEBSVolumeTypePermitsThroughputInput(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		volumeType string
		want       bool
	}{
		{"empty", "", false},
		{"gp2", elasticsearch.VolumeTypeGp2, false},
		{"gp3", elasticsearch.VolumeTypeGp3, true},
		{"io1", elasticsearch.VolumeTypeIo1, false},
		{"standard", elasticsearch.VolumeTypeStandard, false},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := tfelasticsearch.EBSVolumeTypePermitsThroughputInput(testCase.volumeType); got != testCase.want {
				t.Errorf("EBSVolumeTypePermitsThroughputInput() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestAccElasticsearchDomain_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_version", "1.5"),
					resource.TestMatchResourceAttr(resourceName, "kibana_endpoint", regexache.MustCompile(`.*es\..*/_plugin/kibana/`)),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElasticsearchDomain_requireHTTPS(t *testing.T) {
	ctx := acctest.Context(t)
	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_endpointOptions(rName, true, "Policy-Min-TLS-1-0-2019-07"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, "aws_elasticsearch_domain.test", &domain),
					testAccCheckDomainEndpointOptions(true, "Policy-Min-TLS-1-0-2019-07", &domain),
				),
			},
			{
				ResourceName:      "aws_elasticsearch_domain.test",
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_endpointOptions(rName, true, "Policy-Min-TLS-1-2-2019-07"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, "aws_elasticsearch_domain.test", &domain),
					testAccCheckDomainEndpointOptions(true, "Policy-Min-TLS-1-2-2019-07", &domain),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_customEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"
	customEndpoint := fmt.Sprintf("%s.example.com", rName)
	certResourceName := "aws_acm_certificate.test"
	certKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, certKey, customEndpoint)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_customEndpoint(rName, true, "Policy-Min-TLS-1-0-2019-07", true, customEndpoint, certKey, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "domain_endpoint_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "domain_endpoint_options.0.custom_endpoint_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "domain_endpoint_options.0.custom_endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_endpoint_options.0.custom_endpoint_certificate_arn", certResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_customEndpoint(rName, true, "Policy-Min-TLS-1-0-2019-07", true, customEndpoint, certKey, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckDomainEndpointOptions(true, "Policy-Min-TLS-1-0-2019-07", &domain),
					testAccCheckCustomEndpoint(resourceName, true, customEndpoint, &domain),
				),
			},
			{
				Config: testAccDomainConfig_customEndpoint(rName, true, "Policy-Min-TLS-1-0-2019-07", false, customEndpoint, certKey, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckDomainEndpointOptions(true, "Policy-Min-TLS-1-0-2019-07", &domain),
					testAccCheckCustomEndpoint(resourceName, false, customEndpoint, &domain),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_Cluster_zoneAwareness(t *testing.T) {
	ctx := acctest.Context(t)
	var domain1, domain2, domain3, domain4 elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_clusterZoneAwarenessAvailabilityZoneCount(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.0.availability_zone_count", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_clusterZoneAwarenessAvailabilityZoneCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain2),
					testAccCheckDomainNotRecreated(&domain1, &domain2), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.0.availability_zone_count", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccDomainConfig_clusterZoneAwarenessEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain3),
					testAccCheckDomainNotRecreated(&domain2, &domain3), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.#", acctest.Ct0),
				),
			},
			{
				Config: testAccDomainConfig_clusterZoneAwarenessAvailabilityZoneCount(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain4),
					testAccCheckDomainNotRecreated(&domain3, &domain4), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.0.availability_zone_count", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_warm(t *testing.T) {
	ctx := acctest.Context(t)
	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_warm(rName, "ultrawarm1.medium.elasticsearch", false, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", ""),
				),
			},
			{
				Config: testAccDomainConfig_warm(rName, "ultrawarm1.medium.elasticsearch", true, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", "6"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", "ultrawarm1.medium.elasticsearch"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_warm(rName, "ultrawarm1.medium.elasticsearch", true, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", "7"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", "ultrawarm1.medium.elasticsearch"),
				),
			},
			{
				Config: testAccDomainConfig_warm(rName, "ultrawarm1.large.elasticsearch", true, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", "7"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", "ultrawarm1.large.elasticsearch"),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_withColdStorageOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_coldStorageOptions(rName, false, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cluster_config.0.cold_storage_options.*", map[string]string{
						names.AttrEnabled: acctest.CtFalse,
					})),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_coldStorageOptions(rName, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cluster_config.0.cold_storage_options.*", map[string]string{
						names.AttrEnabled: acctest.CtTrue,
					})),
			},
		},
	})
}

func TestAccElasticsearchDomain_withDedicatedMaster(t *testing.T) {
	ctx := acctest.Context(t)
	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_dedicatedClusterMaster(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_dedicatedClusterMaster(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
				),
			},
			{
				Config: testAccDomainConfig_dedicatedClusterMaster(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_duplicate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn(ctx)
			_, err := conn.DeleteElasticsearchDomainWithContext(ctx, &elasticsearch.DeleteElasticsearchDomainInput{
				DomainName: aws.String(rName),
			})
			return err
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Create duplicate
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn(ctx)
					_, err := conn.CreateElasticsearchDomainWithContext(ctx, &elasticsearch.CreateElasticsearchDomainInput{
						DomainName: aws.String(rName),
						EBSOptions: &elasticsearch.EBSOptions{
							EBSEnabled: aws.Bool(true),
							VolumeSize: aws.Int64(10),
						},
					})
					if err != nil {
						t.Fatal(err)
					}

					err = tfelasticsearch.WaitForDomainCreation(ctx, conn, rName, 60*time.Minute)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(
						resourceName, "elasticsearch_version", "1.5"),
				),
				ExpectError: regexache.MustCompile(`Elasticsearch Domain .+ already exists`),
			},
		},
	})
}

func TestAccElasticsearchDomain_v23(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_v23(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(
						resourceName, "elasticsearch_version", "2.3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElasticsearchDomain_complex(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_complex(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElasticsearchDomain_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElasticsearchDomain_VPC_update(t *testing.T) {
	ctx := acctest.Context(t)
	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_vpcUpdate1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckNumberOfSecurityGroups(1, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_vpcUpdate2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckNumberOfSecurityGroups(2, &domain),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_internetToVPCEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_internetToVPCEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_AutoTuneOptions(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	autoTuneStartAtTime := testAccGetValidStartAtTime(t, "24h")
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_autoTuneOptions(rName, autoTuneStartAtTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(
						resourceName, "elasticsearch_version", "6.7"),
					resource.TestMatchResourceAttr(resourceName, "kibana_endpoint", regexache.MustCompile(`.*es\..*/_plugin/kibana/`)),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.desired_state", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.0.start_at", autoTuneStartAtTime),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.0.duration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.0.duration.0.value", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.0.duration.0.unit", "HOURS"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.0.cron_expression_for_recurrence", "cron(0 0 ? * 1 *)"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.rollback_on_disable", "NO_ROLLBACK"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElasticsearchDomain_AdvancedSecurityOptions_userDB(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_advancedSecurityOptionsUserDB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckAdvancedSecurityOptions(true, true, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_advancedSecurityOptionsIAM(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckAdvancedSecurityOptions(true, false, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_advancedSecurityOptionsDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckAdvancedSecurityOptions(false, false, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_logPublishingOptions(rName, elasticsearch.LogTypeIndexSlowLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": elasticsearch.LogTypeIndexSlowLogs,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElasticsearchDomain_LogPublishingOptions_searchSlowLogs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_logPublishingOptions(rName, elasticsearch.LogTypeSearchSlowLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": elasticsearch.LogTypeSearchSlowLogs,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElasticsearchDomain_LogPublishingOptions_esApplicationLogs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_logPublishingOptions(rName, elasticsearch.LogTypeEsApplicationLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": elasticsearch.LogTypeEsApplicationLogs,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElasticsearchDomain_LogPublishingOptions_auditLogs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_logPublishingOptions(rName, elasticsearch.LogTypeAuditLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": elasticsearch.LogTypeAuditLogs,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
				// MasterUserOptions are not returned from DescribeElasticsearchDomainConfig
				ImportStateVerifyIgnore: []string{"advanced_security_options.0.master_user_options"},
			},
		},
	})
}

func TestAccElasticsearchDomain_cognitoOptionsCreateAndRemove(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckCognitoIdentityProvider(ctx, t)
			testAccPreCheckIAMServiceLinkedRole(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_cognitoOptions(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckCognitoOptions(true, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_cognitoOptions(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckCognitoOptions(false, &domain),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_cognitoOptionsUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckCognitoIdentityProvider(ctx, t)
			testAccPreCheckIAMServiceLinkedRole(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_cognitoOptions(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckCognitoOptions(false, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_cognitoOptions(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckCognitoOptions(true, &domain),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_policy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElasticsearchDomain_policyIgnoreEquivalent(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_policyOrder(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
				),
			},
			{
				Config:   testAccDomainConfig_policyNewOrder(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccElasticsearchDomain_Encryption_atRestDefaultKey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "6.0", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckDomainEncrypted(true, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElasticsearchDomain_Encryption_atRestSpecifyKey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_encryptAtRestKey(rName, "6.0", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckDomainEncrypted(true, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElasticsearchDomain_Encryption_atRestEnable(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "6.7", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					testAccCheckDomainEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "6.7", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain2),
					testAccCheckDomainEncrypted(true, &domain2),
					testAccCheckDomainNotRecreated(&domain1, &domain2), // note: this check does not work and always passes
				),
			},
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "6.7", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					testAccCheckDomainEncrypted(false, &domain1),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_Encryption_atRestEnableLegacy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "5.6", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					testAccCheckDomainEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "5.6", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain2),
					testAccCheckDomainEncrypted(true, &domain2),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_Encryption_nodeToNode(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "6.0", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckNodeToNodeEncrypted(true, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElasticsearchDomain_Encryption_nodeToNodeEnable(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "6.7", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "6.7", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain2),
					testAccCheckNodeToNodeEncrypted(true, &domain2),
					testAccCheckDomainNotRecreated(&domain1, &domain2), // note: this check does not work and always passes
				),
			},
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "6.7", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_Encryption_nodeToNodeEnableLegacy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "6.0", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "6.0", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain2),
					testAccCheckNodeToNodeEncrypted(true, &domain2),
				),
			},
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "6.0", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDomainConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccElasticsearchDomain_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var input elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_clusterUpdate(rName, 2, 22),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &input),
					testAccCheckNumberOfInstances(2, &input),
					testAccCheckSnapshotHour(22, &input),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_clusterUpdate(rName, 4, 23),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &input),
					testAccCheckNumberOfInstances(4, &input),
					testAccCheckSnapshotHour(23, &input),
				),
			},
		}})
}

func TestAccElasticsearchDomain_VolumeType_update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)

	var input elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_clusterUpdateEBSVolume(rName, 24, 250, 3500),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &input),
					testAccCheckEBSVolumeEnabled(true, &input),
					testAccCheckEBSVolumeSize(24, &input),
					testAccCheckEBSVolumeThroughput(250, &input),
					testAccCheckEBSVolumeIops(3500, &input),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_clusterUpdateInstanceStore(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &input),
					testAccCheckEBSVolumeEnabled(false, &input),
				),
			},
			{
				Config: testAccDomainConfig_clusterUpdateEBSVolume(rName, 12, 125, 3000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &input),
					testAccCheckEBSVolumeEnabled(true, &input),
					testAccCheckEBSVolumeSize(12, &input),
					testAccCheckEBSVolumeThroughput(125, &input),
					testAccCheckEBSVolumeIops(3000, &input),
				),
			},
		}})
}

// Verifies that EBS volume_type can be changed from gp3 to a type which does not
// support the throughput and iops input values (ex. gp2)
//
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/32613
func TestAccElasticsearchDomain_VolumeType_gp3ToGP2(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var input elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_clusterEBSVolumeGP3DefaultIopsThroughput(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &input),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.0.ebs_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.0.volume_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.0.volume_type", "gp3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_clusterEBSVolumeGP2(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &input),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.0.ebs_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.0.volume_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.0.volume_type", "gp2"),
				),
			},
		}})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13867
func TestAccElasticsearchDomain_VolumeType_missing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)

	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_disabledEBSNullVolumeType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.instance_type", "i3.xlarge.elasticsearch"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.instance_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.0.ebs_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.0.volume_size", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ebs_options.0.volume_type", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElasticsearchDomain_Update_version(t *testing.T) {
	ctx := acctest.Context(t)
	var domain1, domain2, domain3 elasticsearch.ElasticsearchDomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_clusterUpdateVersion(rName, "5.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_version", "5.5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_clusterUpdateVersion(rName, "5.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain2),
					testAccCheckDomainNotRecreated(&domain1, &domain2), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_version", "5.6"),
				),
			},
			{
				Config: testAccDomainConfig_clusterUpdateVersion(rName, "6.3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain3),
					testAccCheckDomainNotRecreated(&domain2, &domain3), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_version", "6.3"),
				),
			},
		}})
}

func TestAccElasticsearchDomain_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := testAccRandomDomainName()
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticsearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelasticsearch.ResourceDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRandomDomainName() string {
	return fmt.Sprintf("%s-%s", acctest.ResourcePrefix, sdkacctest.RandString(28-(len(acctest.ResourcePrefix)+1)))
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

func testAccCheckEBSVolumeThroughput(ebsVolumeThroughput int, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.EBSOptions
		if *conf.Throughput != int64(ebsVolumeThroughput) {
			return fmt.Errorf("EBS throughput differ. Given: %d, Expected: %d", *conf.Throughput, ebsVolumeThroughput)
		}
		return nil
	}
}

func testAccCheckEBSVolumeIops(ebsVolumeIops int, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.EBSOptions
		if *conf.Iops != int64(ebsVolumeIops) {
			return fmt.Errorf("EBS IOPS differ. Given: %d, Expected: %d", *conf.Iops, ebsVolumeIops)
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

func testAccCheckDomainExists(ctx context.Context, n string, domain *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ES Domain ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn(ctx)
		resp, err := tfelasticsearch.FindDomainByName(ctx, conn, rs.Primary.Attributes[names.AttrDomainName])
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
			conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn(ctx)

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

func testAccCheckDomainDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elasticsearch_domain" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn(ctx)
			_, err := tfelasticsearch.FindDomainByName(ctx, conn, rs.Primary.Attributes[names.AttrDomainName])

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
}

func testAccGetValidStartAtTime(t *testing.T, timeUntilStart string) string {
	n := time.Now().UTC()
	d, err := time.ParseDuration(timeUntilStart)
	if err != nil {
		t.Fatalf("err parsing timeUntilStart: %s", err)
	}
	return n.Add(d).Format(time.RFC3339)
}

func testAccPreCheckIAMServiceLinkedRole(ctx context.Context, t *testing.T) {
	acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/es")
}

func testAccDomainConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName)
}

func testAccDomainConfig_autoTuneOptions(rName, autoTuneStartAtTime string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name           = %[1]q
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

func testAccDomainConfig_disabledEBSNullVolumeType(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name           = %[1]q
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

func testAccDomainConfig_endpointOptions(rName string, enforceHttps bool, tlsSecurityPolicy string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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

resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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

func testAccDomainConfig_clusterZoneAwarenessAvailabilityZoneCount(rName string, availabilityZoneCount int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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

func testAccDomainConfig_clusterZoneAwarenessEnabled(rName string, zoneAwarenessEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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

func testAccDomainConfig_warm(rName, warmType string, enabled bool, warmCnt int) string {
	warmConfig := ""
	if enabled {
		warmConfig = fmt.Sprintf(`
    warm_count = %[1]d
    warm_type = %[2]q
`, warmCnt, warmType)
	}

	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name           = %[1]q
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

func testAccDomainConfig_dedicatedClusterMaster(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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

func testAccDomainConfig_coldStorageOptions(rName string, dMasterEnabled bool, warmEnabled bool, csEnabled bool) string {
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
  domain_name = %[1]q

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

func testAccDomainConfig_clusterUpdate(rName string, instanceInt, snapshotInt int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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

func testAccDomainConfig_clusterUpdateEBSVolume(rName string, volumeSize int, volumeThroughput int, volumeIops int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

  elasticsearch_version = "6.0"

  advanced_options = {
    "indices.fielddata.cache.size" = 80
  }

  ebs_options {
    ebs_enabled = true
    volume_size = %d
    throughput  = %d
    volume_type = "gp3"
    iops        = %d
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t3.small.elasticsearch"
  }
}
`, rName, volumeSize, volumeThroughput, volumeIops)
}

func testAccDomainConfig_clusterEBSVolumeGP3DefaultIopsThroughput(rName string, volumeSize int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name           = %[1]q
  elasticsearch_version = "7.10"

  ebs_options {
    ebs_enabled = true
    volume_size = %[2]d
    volume_type = "gp3"
  }

  cluster_config {
    instance_type = "t3.small.elasticsearch"
  }
}
`, rName, volumeSize)
}

func testAccDomainConfig_clusterEBSVolumeGP2(rName string, volumeSize int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name           = %[1]q
  elasticsearch_version = "7.10"

  ebs_options {
    ebs_enabled = true
    volume_size = %[2]d
    volume_type = "gp2"
  }

  cluster_config {
    instance_type = "t3.small.elasticsearch"
  }
}
`, rName, volumeSize)
}

func testAccDomainConfig_clusterUpdateVersion(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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

func testAccDomainConfig_clusterUpdateInstanceStore(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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

func testAccDomainConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q
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

resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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

resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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

resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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

func testAccDomainConfig_encryptAtRestKey(rName, version string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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

func testAccDomainConfig_nodeToNodeEncryption(rName, version string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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
  domain_name = %[1]q

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

func testAccDomainConfig_v23(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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
  domain_name = %[1]q

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

resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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

resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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

resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

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

func testAccDomainConfig_advancedSecurityOptionsUserDB(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name           = %[1]q
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

func testAccDomainConfig_advancedSecurityOptionsIAM(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_elasticsearch_domain" "test" {
  domain_name           = %[1]q
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

func testAccDomainConfig_advancedSecurityOptionsDisabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name           = %[1]q
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

func testAccDomainConfig_logPublishingOptions(rName, logType string) string {
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
  domain_name           = %[1]q
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
  domain_name = %[1]q

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
