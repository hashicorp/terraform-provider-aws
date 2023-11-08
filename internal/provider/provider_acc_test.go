// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider_test

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccProvider_DefaultTags_emptyBlock(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_defaultTagsEmptyConfigurationBlock(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderDefaultTags_Tags(ctx, t, &provider, map[string]string{}),
				),
			},
		},
	})
}

func TestAccProvider_DefaultTagsTags_none(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: acctest.ConfigDefaultTags_Tags0(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderDefaultTags_Tags(ctx, t, &provider, map[string]string{}),
				),
			},
		},
	})
}

func TestAccProvider_DefaultTagsTags_one(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: acctest.ConfigDefaultTags_Tags1("test", "value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderDefaultTags_Tags(ctx, t, &provider, map[string]string{"test": "value"}),
				),
			},
		},
	})
}

func TestAccProvider_DefaultTagsTags_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: acctest.ConfigDefaultTags_Tags2("test1", "value1", "test2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderDefaultTags_Tags(ctx, t, &provider, map[string]string{
						"test1": "value1",
						"test2": "value2",
					}),
				),
			},
		},
	})
}

func TestAccProvider_DefaultAndIgnoreTags_emptyBlocks(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_defaultAndIgnoreTagsEmptyConfigurationBlock(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderDefaultTags_Tags(ctx, t, &provider, map[string]string{}),
					testAccCheckIgnoreTagsKeys(ctx, t, &provider, []string{}),
					testAccCheckIgnoreTagsKeyPrefixes(ctx, t, &provider, []string{}),
				),
			},
		},
	})
}

func TestAccProvider_endpoints(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider
	var endpoints strings.Builder

	// Initialize each endpoint configuration with matching name and value
	for _, serviceKey := range names.ProviderPackages() {
		endpoints.WriteString(fmt.Sprintf("%s = \"http://%s\"\n", serviceKey, serviceKey))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_endpoints(endpoints.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpoints(ctx, &provider),
				),
			},
		},
	})
}

// TODO: revert this test, it's being used as a simple placeholder
func TestAccProvider_fipsEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_fipsEndpoint(fmt.Sprintf("https://s3-fips.%s.%s", acctest.Region(), acctest.PartitionDNSSuffix()), rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "bucket", rName),
				),
			},
		},
	})
}

type unusualEndpoint struct {
	fieldName string
	thing     string
	url       string
}

func TestAccProvider_unusualEndpoints(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider
	unusual1 := unusualEndpoint{"es", "elasticsearch", "http://notarealendpoint"}
	unusual2 := unusualEndpoint{"databasemigration", "dms", "http://alsonotarealendpoint"}
	unusual3 := unusualEndpoint{"lexmodelbuildingservice", "lexmodels", "http://kingofspain"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_unusualEndpoints(unusual1, unusual2, unusual3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUnusualEndpoints(ctx, &provider, unusual1),
					testAccCheckUnusualEndpoints(ctx, &provider, unusual2),
					testAccCheckUnusualEndpoints(ctx, &provider, unusual3),
				),
			},
		},
	})
}

func TestAccProvider_IgnoreTags_emptyBlock(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_ignoreTagsEmptyConfigurationBlock(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnoreTagsKeys(ctx, t, &provider, []string{}),
					testAccCheckIgnoreTagsKeyPrefixes(ctx, t, &provider, []string{}),
				),
			},
		},
	})
}

func TestAccProvider_IgnoreTagsKeyPrefixes_none(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_ignoreTagsKeyPrefixes0(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnoreTagsKeyPrefixes(ctx, t, &provider, []string{}),
				),
			},
		},
	})
}

func TestAccProvider_IgnoreTagsKeyPrefixes_one(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_ignoreTagsKeyPrefixes3("test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnoreTagsKeyPrefixes(ctx, t, &provider, []string{"test"}),
				),
			},
		},
	})
}

func TestAccProvider_IgnoreTagsKeyPrefixes_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_ignoreTagsKeyPrefixes2("test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnoreTagsKeyPrefixes(ctx, t, &provider, []string{"test1", "test2"}),
				),
			},
		},
	})
}

func TestAccProvider_IgnoreTagsKeys_none(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_ignoreTagsKeys0(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnoreTagsKeys(ctx, t, &provider, []string{}),
				),
			},
		},
	})
}

func TestAccProvider_IgnoreTagsKeys_one(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_ignoreTagsKeys1("test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnoreTagsKeys(ctx, t, &provider, []string{"test"}),
				),
			},
		},
	})
}

func TestAccProvider_IgnoreTagsKeys_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_ignoreTagsKeys2("test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnoreTagsKeys(ctx, t, &provider, []string{"test1", "test2"}),
				),
			},
		},
	})
}

func TestAccProvider_Region_c2s(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_region(endpoints.UsIsoEast1RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSuffix(ctx, t, &provider, "c2s.ic.gov"),
					testAccCheckPartition(ctx, t, &provider, endpoints.AwsIsoPartitionID),
					testAccCheckReverseDNSPrefix(ctx, t, &provider, "gov.ic.c2s"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccProvider_Region_china(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_region(endpoints.CnNorthwest1RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSuffix(ctx, t, &provider, "amazonaws.com.cn"),
					testAccCheckPartition(ctx, t, &provider, endpoints.AwsCnPartitionID),
					testAccCheckReverseDNSPrefix(ctx, t, &provider, "cn.com.amazonaws"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccProvider_Region_commercial(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_region(endpoints.UsWest2RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSuffix(ctx, t, &provider, "amazonaws.com"),
					testAccCheckPartition(ctx, t, &provider, endpoints.AwsPartitionID),
					testAccCheckReverseDNSPrefix(ctx, t, &provider, "com.amazonaws"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccProvider_Region_govCloud(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_region(endpoints.UsGovWest1RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSuffix(ctx, t, &provider, "amazonaws.com"),
					testAccCheckPartition(ctx, t, &provider, endpoints.AwsUsGovPartitionID),
					testAccCheckReverseDNSPrefix(ctx, t, &provider, "com.amazonaws"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccProvider_Region_sc2s(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_region(endpoints.UsIsobEast1RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSuffix(ctx, t, &provider, "sc2s.sgov.gov"),
					testAccCheckPartition(ctx, t, &provider, endpoints.AwsIsoBPartitionID),
					testAccCheckReverseDNSPrefix(ctx, t, &provider, "gov.sgov.sc2s"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccProvider_Region_stsRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_stsRegion(endpoints.UsEast1RegionID, endpoints.UsWest2RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegion(ctx, t, &provider, endpoints.UsEast1RegionID),
					testAccCheckSTSRegion(ctx, t, &provider, endpoints.UsWest2RegionID),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccProvider_AssumeRole_empty(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_assumeRoleEmpty,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckCallerIdentityAccountID("data.aws_caller_identity.current"),
				),
			},
		},
	})
}

func testAccProtoV5ProviderFactoriesInternal(ctx context.Context, t *testing.T, v **schema.Provider) map[string]func() (tfprotov5.ProviderServer, error) {
	providerServerFactory, p, err := provider.ProtoV5ProviderServerFactory(ctx)

	if err != nil {
		t.Fatal(err)
	}

	providerServer := providerServerFactory()
	*v = p

	return map[string]func() (tfprotov5.ProviderServer, error){
		acctest.ProviderName: func() (tfprotov5.ProviderServer, error) { //nolint:unparam
			return providerServer, nil
		},
	}
}

func testAccCheckPartition(ctx context.Context, t *testing.T, p **schema.Provider, expectedPartition string) resource.TestCheckFunc { //nolint:unparam
	return func(s *terraform.State) error {
		if p == nil || *p == nil || (*p).Meta() == nil || (*p).Meta().(*conns.AWSClient) == nil {
			return fmt.Errorf("provider not initialized")
		}

		providerPartition := (*p).Meta().(*conns.AWSClient).Partition

		if providerPartition != expectedPartition {
			return fmt.Errorf("expected DNS Suffix (%s), got: %s", expectedPartition, providerPartition)
		}

		return nil
	}
}

func testAccCheckDNSSuffix(ctx context.Context, t *testing.T, p **schema.Provider, expectedDnsSuffix string) resource.TestCheckFunc { //nolint:unparam
	return func(s *terraform.State) error {
		if p == nil || *p == nil || (*p).Meta() == nil || (*p).Meta().(*conns.AWSClient) == nil {
			return fmt.Errorf("provider not initialized")
		}

		providerDnsSuffix := (*p).Meta().(*conns.AWSClient).DNSSuffix

		if providerDnsSuffix != expectedDnsSuffix {
			return fmt.Errorf("expected DNS Suffix (%s), got: %s", expectedDnsSuffix, providerDnsSuffix)
		}

		return nil
	}
}

func testAccCheckRegion(ctx context.Context, t *testing.T, p **schema.Provider, expectedRegion string) resource.TestCheckFunc { //nolint:unparam
	return func(s *terraform.State) error {
		if p == nil || *p == nil || (*p).Meta() == nil || (*p).Meta().(*conns.AWSClient) == nil {
			return fmt.Errorf("provider not initialized")
		}

		if got := (*p).Meta().(*conns.AWSClient).Region; got != expectedRegion {
			return fmt.Errorf("expected Region (%s), got: %s", expectedRegion, got)
		}

		return nil
	}
}

func testAccCheckSTSRegion(ctx context.Context, t *testing.T, p **schema.Provider, expectedRegion string) resource.TestCheckFunc { //nolint:unparam
	return func(s *terraform.State) error {
		if p == nil || *p == nil || (*p).Meta() == nil || (*p).Meta().(*conns.AWSClient) == nil {
			return fmt.Errorf("provider not initialized")
		}

		stsRegion := aws.StringValue((*p).Meta().(*conns.AWSClient).STSConn(ctx).Config.Region)

		if stsRegion != expectedRegion {
			return fmt.Errorf("expected STS Region (%s), got: %s", expectedRegion, stsRegion)
		}

		return nil
	}
}

func testAccCheckReverseDNSPrefix(ctx context.Context, t *testing.T, p **schema.Provider, expectedReverseDnsPrefix string) resource.TestCheckFunc { //nolint:unparam
	return func(s *terraform.State) error {
		if p == nil || *p == nil || (*p).Meta() == nil || (*p).Meta().(*conns.AWSClient) == nil {
			return fmt.Errorf("provider not initialized")
		}
		providerReverseDnsPrefix := (*p).Meta().(*conns.AWSClient).ReverseDNSPrefix

		if providerReverseDnsPrefix != expectedReverseDnsPrefix {
			return fmt.Errorf("expected DNS Suffix (%s), got: %s", expectedReverseDnsPrefix, providerReverseDnsPrefix)
		}

		return nil
	}
}

func testAccCheckIgnoreTagsKeyPrefixes(ctx context.Context, t *testing.T, p **schema.Provider, expectedKeyPrefixes []string) resource.TestCheckFunc { //nolint:unparam
	return func(s *terraform.State) error {
		if p == nil || *p == nil || (*p).Meta() == nil || (*p).Meta().(*conns.AWSClient) == nil {
			return fmt.Errorf("provider not initialized")
		}

		providerClient := (*p).Meta().(*conns.AWSClient)
		ignoreTagsConfig := providerClient.IgnoreTagsConfig

		if ignoreTagsConfig == nil || ignoreTagsConfig.KeyPrefixes == nil {
			if len(expectedKeyPrefixes) != 0 {
				return fmt.Errorf("expected key_prefixes (%d) length, got: 0", len(expectedKeyPrefixes))
			}

			return nil
		}

		actualKeyPrefixes := ignoreTagsConfig.KeyPrefixes.Keys()

		if len(actualKeyPrefixes) != len(expectedKeyPrefixes) {
			return fmt.Errorf("expected key_prefixes (%d) length, got: %d", len(expectedKeyPrefixes), len(actualKeyPrefixes))
		}

		for _, expectedElement := range expectedKeyPrefixes {
			var found bool

			for _, actualElement := range actualKeyPrefixes {
				if actualElement == expectedElement {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("expected key_prefixes element, but was missing: %s", expectedElement)
			}
		}

		for _, actualElement := range actualKeyPrefixes {
			var found bool

			for _, expectedElement := range expectedKeyPrefixes {
				if actualElement == expectedElement {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("unexpected key_prefixes element: %s", actualElement)
			}
		}

		return nil
	}
}

func testAccCheckIgnoreTagsKeys(ctx context.Context, t *testing.T, p **schema.Provider, expectedKeys []string) resource.TestCheckFunc { //nolint:unparam
	return func(s *terraform.State) error {
		if p == nil || *p == nil || (*p).Meta() == nil || (*p).Meta().(*conns.AWSClient) == nil {
			return fmt.Errorf("provider not initialized")
		}

		providerClient := (*p).Meta().(*conns.AWSClient)
		ignoreTagsConfig := providerClient.IgnoreTagsConfig

		if ignoreTagsConfig == nil || ignoreTagsConfig.Keys == nil {
			if len(expectedKeys) != 0 {
				return fmt.Errorf("expected keys (%d) length, got: 0", len(expectedKeys))
			}

			return nil
		}

		actualKeys := ignoreTagsConfig.Keys.Keys()

		if len(actualKeys) != len(expectedKeys) {
			return fmt.Errorf("expected keys (%d) length, got: %d", len(expectedKeys), len(actualKeys))
		}

		for _, expectedElement := range expectedKeys {
			var found bool

			for _, actualElement := range actualKeys {
				if actualElement == expectedElement {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("expected keys element, but was missing: %s", expectedElement)
			}
		}

		for _, actualElement := range actualKeys {
			var found bool

			for _, expectedElement := range expectedKeys {
				if actualElement == expectedElement {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("unexpected keys element: %s", actualElement)
			}
		}

		return nil
	}
}

func testAccCheckProviderDefaultTags_Tags(ctx context.Context, t *testing.T, p **schema.Provider, expectedTags map[string]string) resource.TestCheckFunc { //nolint:unparam
	return func(s *terraform.State) error {
		if p == nil || *p == nil || (*p).Meta() == nil || (*p).Meta().(*conns.AWSClient) == nil {
			return fmt.Errorf("provider not initialized")
		}

		providerClient := (*p).Meta().(*conns.AWSClient)
		defaultTagsConfig := providerClient.DefaultTagsConfig

		if defaultTagsConfig == nil || len(defaultTagsConfig.Tags) == 0 {
			if len(expectedTags) != 0 {
				return fmt.Errorf("expected keys (%d) length, got: 0", len(expectedTags))
			}

			return nil
		}

		actualTags := defaultTagsConfig.Tags

		if len(actualTags) != len(expectedTags) {
			return fmt.Errorf("expected tags (%d) length, got: %d", len(expectedTags), len(actualTags))
		}

		for _, expectedElement := range expectedTags {
			var found bool

			for _, actualElement := range actualTags {
				if aws.StringValue(actualElement.Value) == expectedElement {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("expected tags element, but was missing: %s", expectedElement)
			}
		}

		for _, actualElement := range actualTags {
			var found bool

			for _, expectedElement := range expectedTags {
				if aws.StringValue(actualElement.Value) == expectedElement {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("unexpected tags element: %s", actualElement)
			}
		}

		return nil
	}
}

func testAccCheckEndpoints(_ context.Context, p **schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if p == nil || *p == nil || (*p).Meta() == nil || (*p).Meta().(*conns.AWSClient) == nil {
			return fmt.Errorf("provider not initialized")
		}

		providerClient := (*p).Meta().(*conns.AWSClient)

		for _, serviceKey := range names.Aliases() {
			methodName := serviceConn(serviceKey)
			method := reflect.ValueOf(providerClient).MethodByName(methodName)
			if !method.IsValid() {
				continue
			}
			if method.Kind() != reflect.Func {
				return fmt.Errorf("value %q is not a function", methodName)
			}
			if !funcHasConnFuncSignature(method) {
				return fmt.Errorf("function %q does not match expected signature", methodName)
			}

			result := method.Call([]reflect.Value{
				reflect.ValueOf(context.Background()),
			})
			if l := len(result); l != 1 {
				return fmt.Errorf("expected 1 result, got %d", l)
			}
			providerClientField := result[0]

			if !providerClientField.IsValid() {
				return fmt.Errorf("unable to match conns.AWSClient struct field name for endpoint name: %s", serviceKey)
			}

			if !reflect.Indirect(providerClientField).FieldByName("Config").IsValid() {
				continue // currently unknown how to do this check for v2 clients
			}

			actualEndpoint := reflect.Indirect(reflect.Indirect(providerClientField).FieldByName("Config").FieldByName("Endpoint")).String()
			expectedEndpoint := fmt.Sprintf("http://%s", serviceKey)

			if actualEndpoint != expectedEndpoint {
				return fmt.Errorf("expected endpoint (%s) value (%s), got: %s", serviceKey, expectedEndpoint, actualEndpoint)
			}
		}

		return nil
	}
}

func testAccCheckUnusualEndpoints(_ context.Context, p **schema.Provider, unusual unusualEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if p == nil || *p == nil || (*p).Meta() == nil || (*p).Meta().(*conns.AWSClient) == nil {
			return fmt.Errorf("provider not initialized")
		}

		providerClient := (*p).Meta().(*conns.AWSClient)

		methodName := serviceConn(unusual.thing)
		method := reflect.ValueOf(providerClient).MethodByName(methodName)
		if method.Kind() != reflect.Func {
			return fmt.Errorf("value %q is not a function", methodName)
		}
		if !funcHasConnFuncSignature(method) {
			return fmt.Errorf("function %q does not match expected signature", methodName)
		}

		result := method.Call([]reflect.Value{
			reflect.ValueOf(context.Background()),
		})
		if l := len(result); l != 1 {
			return fmt.Errorf("expected 1 result, got %d", l)
		}
		providerClientField := result[0]

		if !providerClientField.IsValid() {
			return fmt.Errorf("unable to match conns.AWSClient struct field name for endpoint name: %s", unusual.thing)
		}

		actualEndpoint := reflect.Indirect(reflect.Indirect(providerClientField).FieldByName("Config").FieldByName("Endpoint")).String()
		expectedEndpoint := unusual.url

		if actualEndpoint != expectedEndpoint {
			return fmt.Errorf("expected endpoint (%s) value (%s), got: %s", unusual.thing, expectedEndpoint, actualEndpoint)
		}

		return nil
	}
}

func funcHasConnFuncSignature(method reflect.Value) bool {
	typ := method.Type()
	if typ.NumIn() != 1 {
		return false
	}

	fn := func(ctx context.Context) {}
	ftyp := reflect.TypeOf(fn)

	return typ.In(0) == ftyp.In(0)
}

func serviceConn(key string) string {
	serviceUpper := ""
	var err error
	if serviceUpper, err = names.ProviderNameUpper(key); err != nil {
		return ""
	}

	return fmt.Sprintf("%sConn", serviceUpper)
}

const testAccProviderConfig_assumeRoleEmpty = `
provider "aws" {
  assume_role {
  }
}

data "aws_caller_identity" "current" {}
` //lintignore:AT004

const testAccProviderConfig_base = `
data "aws_region" "provider_test" {}

# Required to initialize the provider.
data "aws_service" "provider_test" {
  region     = data.aws_region.provider_test.name
  service_id = "s3"
}
`

func testAccProviderConfig_endpoints(endpoints string) string {
	//lintignore:AT004
	return acctest.ConfigCompose(testAccProviderConfig_base, fmt.Sprintf(`
provider "aws" {
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    %[1]s
  }
}
`, endpoints))
}

func testAccProviderConfig_fipsEndpoint(endpoint, rName string) string {
	//lintignore:AT004
	return acctest.ConfigCompose(testAccProviderConfig_base, fmt.Sprintf(`
provider "aws" {
  endpoints {
    s3 = %[1]q
  }
  https_proxy = ""
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}
`, endpoint, rName))
}

func testAccProviderConfig_unusualEndpoints(unusual1, unusual2, unusual3 unusualEndpoint) string {
	//lintignore:AT004
	return acctest.ConfigCompose(testAccProviderConfig_base, fmt.Sprintf(`
provider "aws" {
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    %[1]s = %[2]q
    %[3]s = %[4]q
    %[5]s = %[6]q
  }
}
`, unusual1.fieldName, unusual1.url, unusual2.fieldName, unusual2.url, unusual3.fieldName, unusual3.url))
}

func testAccProviderConfig_ignoreTagsKeys0() string {
	//lintignore:AT004
	return acctest.ConfigCompose(testAccProviderConfig_base, `
provider "aws" {
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func testAccProviderConfig_ignoreTagsKeys1(tag1 string) string {
	//lintignore:AT004
	return acctest.ConfigCompose(testAccProviderConfig_base, fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    keys = [%[1]q]
  }

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`, tag1))
}

func testAccProviderConfig_ignoreTagsKeys2(tag1, tag2 string) string {
	//lintignore:AT004
	return acctest.ConfigCompose(testAccProviderConfig_base, fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    keys = [%[1]q, %[2]q]
  }

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`, tag1, tag2))
}

func testAccProviderConfig_ignoreTagsKeyPrefixes0() string {
	//lintignore:AT004
	return acctest.ConfigCompose(testAccProviderConfig_base, `
provider "aws" {
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func testAccProviderConfig_ignoreTagsKeyPrefixes3(tagPrefix1 string) string {
	//lintignore:AT004
	return acctest.ConfigCompose(testAccProviderConfig_base, fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    key_prefixes = [%[1]q]
  }

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`, tagPrefix1))
}

func testAccProviderConfig_ignoreTagsKeyPrefixes2(tagPrefix1, tagPrefix2 string) string {
	//lintignore:AT004
	return acctest.ConfigCompose(testAccProviderConfig_base, fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    key_prefixes = [%[1]q, %[2]q]
  }

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`, tagPrefix1, tagPrefix2))
}

func testAccProviderConfig_defaultTagsEmptyConfigurationBlock() string {
	//lintignore:AT004
	return acctest.ConfigCompose(testAccProviderConfig_base, `
provider "aws" {
  default_tags {}

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func testAccProviderConfig_defaultAndIgnoreTagsEmptyConfigurationBlock() string {
	//lintignore:AT004
	return acctest.ConfigCompose(testAccProviderConfig_base, `
provider "aws" {
  default_tags {}
  ignore_tags {}

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func testAccProviderConfig_ignoreTagsEmptyConfigurationBlock() string {
	//lintignore:AT004
	return acctest.ConfigCompose(testAccProviderConfig_base, `
provider "aws" {
  ignore_tags {}

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func testAccProviderConfig_region(region string) string {
	//lintignore:AT004
	return acctest.ConfigCompose(testAccProviderConfig_base, fmt.Sprintf(`
provider "aws" {
  region                      = %[1]q
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`, region))
}

func testAccProviderConfig_stsRegion(region, stsRegion string) string {
	//lintignore:AT004
	return acctest.ConfigCompose(testAccProviderConfig_base, fmt.Sprintf(`
provider "aws" {
  region                      = %[1]q
  sts_region                  = %[2]q
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`, region, stsRegion))
}
