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
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccProvider_DefaultTags_emptyBlock(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_defaultTagsEmptyConfigurationBlock(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderDefaultTags_Tags(t, &provider, map[string]string{}),
				),
			},
		},
	})
}

func TestAccProvider_DefaultTagsTags_none(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: acctest.ConfigDefaultTags_Tags0(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderDefaultTags_Tags(t, &provider, map[string]string{}),
				),
			},
		},
	})
}

func TestAccProvider_DefaultTagsTags_one(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: acctest.ConfigDefaultTags_Tags1("test", "value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderDefaultTags_Tags(t, &provider, map[string]string{"test": "value"}),
				),
			},
		},
	})
}

func TestAccProvider_DefaultTagsTags_multiple(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: acctest.ConfigDefaultTags_Tags2("test1", "value1", "test2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderDefaultTags_Tags(t, &provider, map[string]string{
						"test1": "value1",
						"test2": "value2",
					}),
				),
			},
		},
	})
}

func TestAccProvider_DefaultAndIgnoreTags_emptyBlocks(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_defaultAndIgnoreTagsEmptyConfigurationBlock(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProviderDefaultTags_Tags(t, &provider, map[string]string{}),
					testAccCheckIgnoreTagsKeys(t, &provider, []string{}),
					testAccCheckIgnoreTagsKeyPrefixes(t, &provider, []string{}),
				),
			},
		},
	})
}

func TestAccProvider_endpoints(t *testing.T) {
	var provider *schema.Provider
	var endpoints strings.Builder

	// Initialize each endpoint configuration with matching name and value
	for _, serviceKey := range names.ProviderPackages() {
		endpoints.WriteString(fmt.Sprintf("%s = \"http://%s\"\n", serviceKey, serviceKey))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_endpoints(endpoints.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpoints(&provider),
				),
			},
		},
	})
}

func TestAccProvider_fipsEndpoint(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
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
	var provider *schema.Provider

	unusual1 := unusualEndpoint{"es", "elasticsearch", "http://notarealendpoint"}
	unusual2 := unusualEndpoint{"databasemigration", "dms", "http://alsonotarealendpoint"}
	unusual3 := unusualEndpoint{"lexmodelbuildingservice", "lexmodels", "http://kingofspain"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_unusualEndpoints(unusual1, unusual2, unusual3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUnusualEndpoints(&provider, unusual1),
					testAccCheckUnusualEndpoints(&provider, unusual2),
					testAccCheckUnusualEndpoints(&provider, unusual3),
				),
			},
		},
	})
}

func TestAccProvider_IgnoreTags_emptyBlock(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_ignoreTagsEmptyConfigurationBlock(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnoreTagsKeys(t, &provider, []string{}),
					testAccCheckIgnoreTagsKeyPrefixes(t, &provider, []string{}),
				),
			},
		},
	})
}

func TestAccProvider_IgnoreTagsKeyPrefixes_none(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_ignoreTagsKeyPrefixes0(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnoreTagsKeyPrefixes(t, &provider, []string{}),
				),
			},
		},
	})
}

func TestAccProvider_IgnoreTagsKeyPrefixes_one(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_ignoreTagsKeyPrefixes3("test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnoreTagsKeyPrefixes(t, &provider, []string{"test"}),
				),
			},
		},
	})
}

func TestAccProvider_IgnoreTagsKeyPrefixes_multiple(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_ignoreTagsKeyPrefixes2("test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnoreTagsKeyPrefixes(t, &provider, []string{"test1", "test2"}),
				),
			},
		},
	})
}

func TestAccProvider_IgnoreTagsKeys_none(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_ignoreTagsKeys0(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnoreTagsKeys(t, &provider, []string{}),
				),
			},
		},
	})
}

func TestAccProvider_IgnoreTagsKeys_one(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_ignoreTagsKeys1("test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnoreTagsKeys(t, &provider, []string{"test"}),
				),
			},
		},
	})
}

func TestAccProvider_IgnoreTagsKeys_multiple(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_ignoreTagsKeys2("test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIgnoreTagsKeys(t, &provider, []string{"test1", "test2"}),
				),
			},
		},
	})
}

func TestAccProvider_Region_c2s(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_region(endpoints.UsIsoEast1RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSuffix(t, &provider, "c2s.ic.gov"),
					testAccCheckPartition(t, &provider, endpoints.AwsIsoPartitionID),
					testAccCheckReverseDNSPrefix(t, &provider, "gov.ic.c2s"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccProvider_Region_china(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_region(endpoints.CnNorthwest1RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSuffix(t, &provider, "amazonaws.com.cn"),
					testAccCheckPartition(t, &provider, endpoints.AwsCnPartitionID),
					testAccCheckReverseDNSPrefix(t, &provider, "cn.com.amazonaws"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccProvider_Region_commercial(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_region(endpoints.UsWest2RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSuffix(t, &provider, "amazonaws.com"),
					testAccCheckPartition(t, &provider, endpoints.AwsPartitionID),
					testAccCheckReverseDNSPrefix(t, &provider, "com.amazonaws"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccProvider_Region_govCloud(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_region(endpoints.UsGovWest1RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSuffix(t, &provider, "amazonaws.com"),
					testAccCheckPartition(t, &provider, endpoints.AwsUsGovPartitionID),
					testAccCheckReverseDNSPrefix(t, &provider, "com.amazonaws"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccProvider_Region_sc2s(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_region(endpoints.UsIsobEast1RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSuffix(t, &provider, "sc2s.sgov.gov"),
					testAccCheckPartition(t, &provider, endpoints.AwsIsoBPartitionID),
					testAccCheckReverseDNSPrefix(t, &provider, "gov.sgov.sc2s"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccProvider_Region_stsRegion(t *testing.T) {
	var provider *schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(t, &provider),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig_stsRegion(endpoints.UsEast1RegionID, endpoints.UsWest2RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegion(t, &provider, endpoints.UsEast1RegionID),
					testAccCheckSTSRegion(t, &provider, endpoints.UsWest2RegionID),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccProvider_AssumeRole_empty(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
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

func testAccProtoV5ProviderFactoriesInternal(t *testing.T, v **schema.Provider) map[string]func() (tfprotov5.ProviderServer, error) {
	providerServerFactory, p, err := provider.ProtoV5ProviderServerFactory(context.Background())

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

func testAccCheckPartition(t *testing.T, p **schema.Provider, expectedPartition string) resource.TestCheckFunc { //nolint:unparam
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

func testAccCheckDNSSuffix(t *testing.T, p **schema.Provider, expectedDnsSuffix string) resource.TestCheckFunc { //nolint:unparam
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

func testAccCheckRegion(t *testing.T, p **schema.Provider, expectedRegion string) resource.TestCheckFunc { //nolint:unparam
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

func testAccCheckSTSRegion(t *testing.T, p **schema.Provider, expectedRegion string) resource.TestCheckFunc { //nolint:unparam
	return func(s *terraform.State) error {
		if p == nil || *p == nil || (*p).Meta() == nil || (*p).Meta().(*conns.AWSClient) == nil {
			return fmt.Errorf("provider not initialized")
		}

		stsRegion := aws.StringValue((*p).Meta().(*conns.AWSClient).STSConn().Config.Region)

		if stsRegion != expectedRegion {
			return fmt.Errorf("expected STS Region (%s), got: %s", expectedRegion, stsRegion)
		}

		return nil
	}
}

func testAccCheckReverseDNSPrefix(t *testing.T, p **schema.Provider, expectedReverseDnsPrefix string) resource.TestCheckFunc { //nolint:unparam
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

func testAccCheckIgnoreTagsKeyPrefixes(t *testing.T, p **schema.Provider, expectedKeyPrefixes []string) resource.TestCheckFunc { //nolint:unparam
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

func testAccCheckIgnoreTagsKeys(t *testing.T, p **schema.Provider, expectedKeys []string) resource.TestCheckFunc { //nolint:unparam
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

func testAccCheckProviderDefaultTags_Tags(t *testing.T, p **schema.Provider, expectedTags map[string]string) resource.TestCheckFunc { //nolint:unparam
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

func testAccCheckEndpoints(p **schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if p == nil || *p == nil || (*p).Meta() == nil || (*p).Meta().(*conns.AWSClient) == nil {
			return fmt.Errorf("provider not initialized")
		}

		providerClient := (*p).Meta().(*conns.AWSClient)

		for _, serviceKey := range names.ProviderPackages() {
			method := reflect.ValueOf(providerClient).MethodByName(serviceConn(serviceKey))
			if !method.IsValid() {
				continue
			}
			result := method.Call([]reflect.Value{})
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

func testAccCheckUnusualEndpoints(p **schema.Provider, unusual unusualEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if p == nil || *p == nil || (*p).Meta() == nil || (*p).Meta().(*conns.AWSClient) == nil {
			return fmt.Errorf("provider not initialized")
		}

		providerClient := (*p).Meta().(*conns.AWSClient)

		result := reflect.ValueOf(providerClient).MethodByName(serviceConn(unusual.thing)).Call([]reflect.Value{})
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
  skip_get_ec2_platforms      = true
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
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}
`, endpoint, rName))
}

func testAccProviderConfig_unusualEndpoints(unusual1, unusual2, unusual3 unusualEndpoint) string {
	//lintignore:AT004
	return acctest.ConfigCompose(testAccProviderConfig_base, fmt.Sprintf(`
provider "aws" {
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
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
  skip_get_ec2_platforms      = true
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
  skip_get_ec2_platforms      = true
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
  skip_get_ec2_platforms      = true
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
  skip_get_ec2_platforms      = true
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
  skip_get_ec2_platforms      = true
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
  skip_get_ec2_platforms      = true
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
  skip_get_ec2_platforms      = true
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
  skip_get_ec2_platforms      = true
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
  skip_get_ec2_platforms      = true
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
  skip_get_ec2_platforms      = true
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
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`, region, stsRegion))
}
