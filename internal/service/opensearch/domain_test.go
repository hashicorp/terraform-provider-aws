// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfopensearch "github.com/hashicorp/terraform-provider-aws/internal/service/opensearch"
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
		{"gp2", opensearchservice.VolumeTypeGp2, false},
		{"gp3", opensearchservice.VolumeTypeGp3, true},
		{"io1", opensearchservice.VolumeTypeIo1, true},
		{"standard", opensearchservice.VolumeTypeStandard, false},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := tfopensearch.EBSVolumeTypePermitsIopsInput(testCase.volumeType); got != testCase.want {
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
		{"gp2", opensearchservice.VolumeTypeGp2, false},
		{"gp3", opensearchservice.VolumeTypeGp3, true},
		{"io1", opensearchservice.VolumeTypeIo1, false},
		{"standard", opensearchservice.VolumeTypeStandard, false},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := tfopensearch.EBSVolumeTypePermitsThroughputInput(testCase.volumeType); got != testCase.want {
				t.Errorf("EBSVolumeTypePermitsThroughputInput() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestParseEngineVersion(t *testing.T) {
	t.Parallel()

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
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestMatchResourceAttr(resourceName, "dashboard_endpoint", regexache.MustCompile(`.*(opensearch|es)\..*/_dashboards`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestMatchResourceAttr(resourceName, "kibana_endpoint", regexache.MustCompile(`.*(opensearch|es)\..*/_plugin/kibana/`)),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vpc_options.#", acctest.Ct0),
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

func TestAccOpenSearchDomain_requireHTTPS(t *testing.T) {
	ctx := acctest.Context(t)
	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_endpointOptions(rName, true, "Policy-Min-TLS-1-0-2019-07"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, "aws_opensearch_domain.test", &domain),
					testAccCheckDomainEndpointOptions(true, "Policy-Min-TLS-1-0-2019-07", &domain),
				),
			},
			{
				ResourceName:      "aws_opensearch_domain.test",
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_endpointOptions(rName, true, "Policy-Min-TLS-1-2-2019-07"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, "aws_opensearch_domain.test", &domain),
					testAccCheckDomainEndpointOptions(true, "Policy-Min-TLS-1-2-2019-07", &domain),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_customEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"
	customEndpoint := fmt.Sprintf("%s.example.com", rName)
	certResourceName := "aws_acm_certificate.test"
	certKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, certKey, customEndpoint)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
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

func TestAccOpenSearchDomain_Cluster_zoneAwareness(t *testing.T) {
	ctx := acctest.Context(t)
	var domain1, domain2, domain3, domain4 opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_clusterZoneAwarenessAZCount(rName, 3),
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
				Config: testAccDomainConfig_clusterZoneAwarenessAZCount(rName, 2),
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
				Config: testAccDomainConfig_clusterZoneAwarenessAZCount(rName, 3),
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

func TestAccOpenSearchDomain_Cluster_coldStorage(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_clusterColdStorageOptions(rName, true, false),
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
				Config: testAccDomainConfig_clusterColdStorageOptions(rName, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cluster_config.0.cold_storage_options.*", map[string]string{
						names.AttrEnabled: acctest.CtTrue,
					})),
			},
		},
	})
}

func TestAccOpenSearchDomain_Cluster_warm(t *testing.T) {
	ctx := acctest.Context(t)
	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_clusterWarm(rName, "ultrawarm1.medium.search", false, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", ""),
				),
			},
			{
				Config: testAccDomainConfig_clusterWarm(rName, "ultrawarm1.medium.search", true, 6),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", "6"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", "ultrawarm1.medium.search"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_clusterWarm(rName, "ultrawarm1.medium.search", true, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", "7"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", "ultrawarm1.medium.search"),
				),
			},
			{
				Config: testAccDomainConfig_clusterWarm(rName, "ultrawarm1.large.search", true, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_count", "7"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.warm_type", "ultrawarm1.large.search"),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_Cluster_dedicatedMaster(t *testing.T) {
	ctx := acctest.Context(t)
	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
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

func TestAccOpenSearchDomain_Cluster_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var input opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
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

func TestAccOpenSearchDomain_Cluster_multiAzWithStandbyEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_multiAzWithStandbyEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.multi_az_with_standby_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_multiAzWithStandbyEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.multi_az_with_standby_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_duplicate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchConn(ctx)
			_, err := conn.DeleteDomainWithContext(ctx, &opensearchservice.DeleteDomainInput{
				DomainName: aws.String(rName),
			})
			return err
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Create duplicate
					conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchConn(ctx)
					_, err := conn.CreateDomainWithContext(ctx, &opensearchservice.CreateDomainInput{
						DomainName: aws.String(rName),
						EBSOptions: &opensearchservice.EBSOptions{
							EBSEnabled: aws.Bool(true),
							VolumeSize: aws.Int64(10),
						},
					})
					if err != nil {
						t.Fatal(err)
					}

					err = tfopensearch.WaitForDomainCreation(ctx, conn, rName, 60*time.Minute)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
				),
				ExpectError: regexache.MustCompile(`OpenSearch Domain ".+" already exists`),
			},
		},
	})
}

func TestAccOpenSearchDomain_v23(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_v23(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrEngineVersion, "Elasticsearch_2.3"),
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

func TestAccOpenSearchDomain_complex(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
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

func TestAccOpenSearchDomain_VPC_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
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

func TestAccOpenSearchDomain_VPC_update(t *testing.T) {
	ctx := acctest.Context(t)
	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
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

func TestAccOpenSearchDomain_VPC_internetToVPCEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
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

func TestAccOpenSearchDomain_ipAddressType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_ipAddressType(rName, "dualstack"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestMatchResourceAttr(resourceName, "dashboard_endpoint", regexache.MustCompile(`.*(opensearch|es)\..*/_dashboards`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "dualstack"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_ipAddressType(rName, "ipv4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestMatchResourceAttr(resourceName, "dashboard_endpoint", regexache.MustCompile(`.*(opensearch|es)\..*/_dashboards`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddressType, "ipv4"),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_autoTuneOptions(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	autoTuneStartAtTime := testAccGetValidStartAtTime(t, "24h")
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_autoTuneOptionsMaintenanceSchedule(rName, autoTuneStartAtTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "Elasticsearch_6.7"),
					resource.TestMatchResourceAttr(resourceName, "kibana_endpoint", regexache.MustCompile(`.*(opensearch|es)\..*/_plugin/kibana/`)),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.desired_state", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.0.start_at", autoTuneStartAtTime),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.0.duration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.0.duration.0.value", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.0.duration.0.unit", "HOURS"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.0.cron_expression_for_recurrence", "cron(0 0 ? * 1 *)"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.rollback_on_disable", "NO_ROLLBACK"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.use_off_peak_window", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_autoTuneOptionsUseOffPeakWindow(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "Elasticsearch_6.7"),
					resource.TestMatchResourceAttr(resourceName, "kibana_endpoint", regexache.MustCompile(`.*(opensearch|es)\..*/_plugin/kibana/`)),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.desired_state", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.maintenance_schedule.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.rollback_on_disable", "NO_ROLLBACK"),
					resource.TestCheckResourceAttr(resourceName, "auto_tune_options.0.use_off_peak_window", acctest.CtTrue),
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

func TestAccOpenSearchDomain_AdvancedSecurityOptions_userDB(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_advancedSecurityOptionsUserDB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckAdvancedSecurityOptions(true, true, false, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
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

func TestAccOpenSearchDomain_AdvancedSecurityOptions_anonymousAuth(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_advancedSecurityOptionsAnonymousAuth(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckAdvancedSecurityOptions(false, true, true, &domain),
				),
			},
			{
				Config: testAccDomainConfig_advancedSecurityOptionsAnonymousAuth(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckAdvancedSecurityOptions(true, true, true, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_advancedSecurityOptionsIAM(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckAdvancedSecurityOptions(true, false, false, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_advancedSecurityOptionsDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					testAccCheckAdvancedSecurityOptions(false, false, false, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
				// MasterUserOptions are not returned from DescribeDomainConfig
				ImportStateVerifyIgnore: []string{
					"advanced_security_options",
				},
			},
		},
	})
}

func TestAccOpenSearchDomain_LogPublishingOptions_indexSlowLogs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_logPublishingOptions(rName, opensearchservice.LogTypeIndexSlowLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": opensearchservice.LogTypeIndexSlowLogs,
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

func TestAccOpenSearchDomain_LogPublishingOptions_searchSlowLogs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_logPublishingOptions(rName, opensearchservice.LogTypeSearchSlowLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": opensearchservice.LogTypeSearchSlowLogs,
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

func TestAccOpenSearchDomain_LogPublishingOptions_applicationLogs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_logPublishingOptions(rName, opensearchservice.LogTypeEsApplicationLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": opensearchservice.LogTypeEsApplicationLogs,
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

func TestAccOpenSearchDomain_LogPublishingOptions_auditLogs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_logPublishingOptions(rName, opensearchservice.LogTypeAuditLogs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "log_publishing_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "log_publishing_options.*", map[string]string{
						"log_type": opensearchservice.LogTypeAuditLogs,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
				// MasterUserOptions are not returned from DescribeDomainConfig
				ImportStateVerifyIgnore: []string{"advanced_security_options.0.master_user_options"},
			},
		},
	})
}

func TestAccOpenSearchDomain_CognitoOptions_createAndRemove(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckCognitoIdentityProvider(ctx, t)
			testAccPreCheckIAMServiceLinkedRole(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
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

func TestAccOpenSearchDomain_CognitoOptions_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckCognitoIdentityProvider(ctx, t)
			testAccPreCheckIAMServiceLinkedRole(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
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

func TestAccOpenSearchDomain_Policy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
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

func TestAccOpenSearchDomain_Policy_addPrincipal(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_policyDocument(rName, 1),
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
				Config: testAccDomainConfig_policyDocument(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateId:           rName,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"access_policies"}, // Principals is a set, and `structure.NormalizeJsonString` doesn't guarantee order
			},
		},
	})
}

func TestAccOpenSearchDomain_Policy_ignoreEquivalent(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
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

func TestAccOpenSearchDomain_Encryption_atRestDefaultKey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "Elasticsearch_6.0", true),
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

func TestAccOpenSearchDomain_Encryption_atRestSpecifyKey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_encryptAtRestKey(rName, "Elasticsearch_6.0", true),
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

func TestAccOpenSearchDomain_Encryption_atRestEnable(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "OpenSearch_2.5", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					testAccCheckDomainEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "OpenSearch_2.5", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain2),
					testAccCheckDomainEncrypted(true, &domain2),
					testAccCheckDomainNotRecreated(&domain1, &domain2), // note: this check does not work and always passes
				),
			},
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "OpenSearch_2.5", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					testAccCheckDomainEncrypted(false, &domain1),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_Encryption_atRestEnableLegacy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "Elasticsearch_5.6", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					testAccCheckDomainEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfig_encryptAtRestDefaultKey(rName, "Elasticsearch_5.6", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain2),
					testAccCheckDomainEncrypted(true, &domain2),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_Encryption_nodeToNode(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "Elasticsearch_6.0", true),
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

func TestAccOpenSearchDomain_Encryption_nodeToNodeEnable(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "OpenSearch_2.5", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "OpenSearch_2.5", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain2),
					testAccCheckNodeToNodeEncrypted(true, &domain2),
					testAccCheckDomainNotRecreated(&domain1, &domain2), // note: this check does not work and always passes
				),
			},
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "OpenSearch_2.5", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_Encryption_nodeToNodeEnableLegacy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain1, domain2 opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "Elasticsearch_6.0", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "Elasticsearch_6.0", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain2),
					testAccCheckNodeToNodeEncrypted(true, &domain2),
				),
			},
			{
				Config: testAccDomainConfig_nodeToNodeEncryption(rName, "Elasticsearch_6.0", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					testAccCheckNodeToNodeEncrypted(false, &domain1),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_offPeakWindowOptions(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_offPeakWindowOptions(rName, 9, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.0.off_peak_window.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.0.off_peak_window.0.window_start_time.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.0.off_peak_window.0.window_start_time.0.hours", "9"),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.0.off_peak_window.0.window_start_time.0.minutes", "30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_offPeakWindowOptions(rName, 10, 15),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.0.off_peak_window.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.0.off_peak_window.0.window_start_time.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.0.off_peak_window.0.window_start_time.0.hours", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.0.off_peak_window.0.window_start_time.0.minutes", "15"),
				),
			},
			{
				Config: testAccDomainConfig_offPeakWindowOptions(rName, 0, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.0.off_peak_window.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.0.off_peak_window.0.window_start_time.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.0.off_peak_window.0.window_start_time.0.hours", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "off_peak_window_options.0.off_peak_window.0.window_start_time.0.minutes", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccOpenSearchDomain_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
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

func TestAccOpenSearchDomain_VolumeType_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var input opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
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
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/27467
func TestAccOpenSearchDomain_VolumeType_gp3ToGP2(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var input opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
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
func TestAccOpenSearchDomain_VolumeType_missing(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	resourceName := "aws_opensearch_domain.test"
	rName := testAccRandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_disabledEBSNullVolume(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.instance_type", "i3.xlarge.search"),
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

func TestAccOpenSearchDomain_versionUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var domain1, domain2, domain3 opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_clusterUpdateVersion(rName, "Elasticsearch_5.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "Elasticsearch_5.5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_clusterUpdateVersion(rName, "Elasticsearch_5.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain2),
					testAccCheckDomainNotRecreated(&domain1, &domain2), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "Elasticsearch_5.6"),
				),
			},
			{
				Config: testAccDomainConfig_clusterUpdateVersion(rName, "Elasticsearch_6.3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain3),
					testAccCheckDomainNotRecreated(&domain2, &domain3), // note: this check does not work and always passes
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "Elasticsearch_6.3"),
				),
			},
		}})
}

func TestAccOpenSearchDomain_softwareUpdateOptions(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var domain opensearchservice.DomainStatus
	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_softwareUpdateOptions(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "software_update_options.0.auto_software_update_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccDomainConfig_softwareUpdateOptions(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "software_update_options.0.auto_software_update_enabled", acctest.CtTrue),
				),
			},
		},
	})
}
func TestAccOpenSearchDomain_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := testAccRandomDomainName()
	resourceName := "aws_opensearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIAMServiceLinkedRole(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfopensearch.ResourceDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRandomDomainName() string {
	return fmt.Sprintf("%s-%s", acctest.ResourcePrefix, sdkacctest.RandString(28-(len(acctest.ResourcePrefix)+1)))
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

func testAccCheckEBSVolumeThroughput(ebsVolumeThroughput int, status *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.EBSOptions
		if *conf.Throughput != int64(ebsVolumeThroughput) {
			return fmt.Errorf("EBS throughput differ. Given: %d, Expected: %d", *conf.Throughput, ebsVolumeThroughput)
		}
		return nil
	}
}

func testAccCheckEBSVolumeIops(ebsVolumeIops int, status *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.EBSOptions
		if *conf.Iops != int64(ebsVolumeIops) {
			return fmt.Errorf("EBS IOPS differ. Given: %d, Expected: %d", *conf.Iops, ebsVolumeIops)
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

func testAccCheckAdvancedSecurityOptions(enabled bool, userDbEnabled bool, anonymousAuthEnabled bool, status *opensearchservice.DomainStatus) resource.TestCheckFunc {
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

		if aws.BoolValue(conf.Enabled) {
			if aws.BoolValue(conf.AnonymousAuthEnabled) != anonymousAuthEnabled {
				return fmt.Errorf(
					"AdvancedSecurityOptions.AnonymousAuthEnabled not set properly. Given: %t, Expected: %t",
					aws.BoolValue(conf.AnonymousAuthEnabled),
					anonymousAuthEnabled,
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

func testAccCheckDomainExists(ctx context.Context, n string, domain *opensearchservice.DomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No OpenSearch Domain ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchConn(ctx)
		resp, err := tfopensearch.FindDomainByName(ctx, conn, rs.Primary.Attributes[names.AttrDomainName])
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
			conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchConn(ctx)

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

func testAccCheckDomainDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearch_domain" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchConn(ctx)
			_, err := tfopensearch.FindDomainByName(ctx, conn, rs.Primary.Attributes[names.AttrDomainName])

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
	acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/opensearchservice")
}

func testAccDomainConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = %[1]q

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName)
}

func testAccDomainConfig_ipAddressType(rName, ipAddressType string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name     = %[1]q
  ip_address_type = %[2]q

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName, ipAddressType)
}

func testAccDomainConfig_autoTuneOptionsMaintenanceSchedule(rName, autoTuneStartAtTime string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = %[1]q
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

func testAccDomainConfig_autoTuneOptionsUseOffPeakWindow(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = %[1]q
  engine_version = "Elasticsearch_6.7"

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  auto_tune_options {
    desired_state       = "ENABLED"
    rollback_on_disable = "NO_ROLLBACK"
    use_off_peak_window = true
  }
}
`, rName)
}

func testAccDomainConfig_disabledEBSNullVolume(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = %[1]q
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

resource "aws_opensearch_domain" "test" {
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

func testAccDomainConfig_clusterZoneAwarenessAZCount(rName string, availabilityZoneCount int) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = %[1]q
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
  domain_name    = %[1]q
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
  domain_name    = %[1]q
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
  domain_name    = %[1]q
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
  domain_name = %[1]q

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

func testAccDomainConfig_multiAzWithStandbyEnabled(rName string, enableStandby bool) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = %[1]q
  engine_version = "OpenSearch_2.7"

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-0-2019-07"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 20
  }

  auto_tune_options {
    desired_state       = "ENABLED"
    rollback_on_disable = "NO_ROLLBACK"
  }

  cluster_config {
    zone_awareness_enabled   = true
    instance_count           = 3
    instance_type            = "m6g.large.search"
    dedicated_master_enabled = true
    dedicated_master_count   = 3
    dedicated_master_type    = "m6g.large.search"

    zone_awareness_config {
      availability_zone_count = 3
    }

    multi_az_with_standby_enabled = %[2]t
  }
}
`, rName, enableStandby)
}

func testAccDomainConfig_clusterUpdate(rName string, instanceInt, snapshotInt int) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
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

func testAccDomainConfig_clusterUpdateEBSVolume(rName string, volumeSize int, volumeThroughput int, volumeIops int) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = %[1]q

  engine_version = "Elasticsearch_6.0"

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
    instance_type          = "t3.small.search"
  }
}
`, rName, volumeSize, volumeThroughput, volumeIops)
}

func testAccDomainConfig_clusterEBSVolumeGP3DefaultIopsThroughput(rName string, volumeSize int) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = %[1]q

  ebs_options {
    ebs_enabled = true
    volume_size = %[2]d
    volume_type = "gp3"
  }

  cluster_config {
    instance_type = "t3.small.search"
  }
}
`, rName, volumeSize)
}

func testAccDomainConfig_clusterEBSVolumeGP2(rName string, volumeSize int) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = %[1]q

  ebs_options {
    ebs_enabled = true
    volume_size = %[2]d
    volume_type = "gp2"
  }

  cluster_config {
    instance_type = "t3.small.search"
  }
}
`, rName, volumeSize)
}

func testAccDomainConfig_clusterUpdateVersion(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = %[1]q

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
  domain_name = %[1]q

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
resource "aws_opensearch_domain" "test" {
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

resource "aws_opensearch_domain" "test" {
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

resource "aws_opensearch_domain" "test" {
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

resource "aws_opensearch_domain" "test" {
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

func testAccDomainConfig_policyDocument(rName string, roleCount int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_opensearch_domain" "test" {
  domain_name = %[1]q

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  access_policies = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["es:*"]
    resources = ["arn:${data.aws_partition.current.partition}:es:*"]
    principals {
      type        = "AWS"
      identifiers = aws_iam_role.test[*].arn
    }
  }
}

resource "aws_iam_role" "test" {
  count              = %[2]d
  name               = "%[1]s-${count.index}"
  assume_role_policy = data.aws_iam_policy_document.role.json
}

data "aws_iam_policy_document" "role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.${data.aws_partition.current.dns_suffix}"]
    }
  }
}
`, rName, roleCount)
}

func testAccDomainConfig_encryptAtRestDefaultKey(rName, version string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = %[1]q

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
  domain_name = %[1]q

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
  domain_name = %[1]q

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
  domain_name = %[1]q

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
  domain_name = %[1]q

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
  domain_name = %[1]q

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
  domain_name = %[1]q

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
  domain_name = %[1]q

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
  domain_name    = %[1]q
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

func testAccDomainConfig_advancedSecurityOptionsAnonymousAuth(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = %[1]q
  engine_version = "Elasticsearch_7.1"

  cluster_config {
    instance_type = "r5.large.search"
  }

  advanced_security_options {
    enabled                        = %[2]t
    anonymous_auth_enabled         = true
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
`, rName, enabled)
}

func testAccDomainConfig_advancedSecurityOptionsIAM(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_opensearch_domain" "test" {
  domain_name    = %[1]q
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
  domain_name    = %[1]q
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
  domain_name    = %[1]q
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
  domain_name = %[1]q

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

func testAccDomainConfig_offPeakWindowOptions(rName string, h, m int) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name    = %[1]q
  engine_version = "Elasticsearch_6.7"

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  off_peak_window_options {
    off_peak_window {
      window_start_time {
        hours   = %[2]d
        minutes = %[3]d
      }
    }
  }
}
`, rName, h, m)
}

func testAccDomainConfig_softwareUpdateOptions(rName string, option bool) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = %[1]q

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  software_update_options {
    auto_software_update_enabled = %[2]t
  }
}
`, rName, option)
}
