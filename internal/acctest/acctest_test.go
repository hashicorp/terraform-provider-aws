package acctest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

func TestProvider(t *testing.T) {
	if err := provider.Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = provider.Provider()
}

func TestReverseDns(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "amazonaws.com",
			input:    "amazonaws.com",
			expected: "com.amazonaws",
		},
		{
			name:     "amazonaws.com.cn",
			input:    "amazonaws.com.cn",
			expected: "cn.com.amazonaws",
		},
		{
			name:     "sc2s.sgov.gov",
			input:    "sc2s.sgov.gov",
			expected: "gov.sgov.sc2s",
		},
		{
			name:     "c2s.ic.gov",
			input:    "c2s.ic.gov",
			expected: "gov.ic.c2s",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			if got, want := conns.ReverseDNS(testCase.input), testCase.expected; got != want {
				t.Errorf("got: %s, expected: %s", got, want)
			}
		})
	}
}

func TestAccNASAcctest_ProviderDefaultTags_emptyBlock(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigDefaultTagsEmptyConfigurationBlock(),
				Check: resource.ComposeTestCheckFunc(
					CheckProviderDefaultTags_Tags(&providers, map[string]string{}),
				),
			},
		},
	})
}

func TestAccNASAcctest_ProviderDefaultTagsTags_none(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigDefaultTags_Tags0(),
				Check: resource.ComposeTestCheckFunc(
					CheckProviderDefaultTags_Tags(&providers, map[string]string{}),
				),
			},
		},
	})
}

func TestAccNASAcctest_ProviderDefaultTagsTags_one(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigDefaultTags_Tags1("test", "value"),
				Check: resource.ComposeTestCheckFunc(
					CheckProviderDefaultTags_Tags(&providers, map[string]string{"test": "value"}),
				),
			},
		},
	})
}

func TestAccNASAcctest_ProviderDefaultTagsTags_multiple(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigDefaultTags_Tags2("test1", "value1", "test2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					CheckProviderDefaultTags_Tags(&providers, map[string]string{
						"test1": "value1",
						"test2": "value2",
					}),
				),
			},
		},
	})
}

func TestAccNASAcctest_ProviderDefaultAndIgnoreTags_emptyBlocks(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigDefaultAndIgnoreTagsEmptyConfigurationBlock(),
				Check: resource.ComposeTestCheckFunc(
					CheckProviderDefaultTags_Tags(&providers, map[string]string{}),
					CheckIgnoreTagsKeys(&providers, []string{}),
					CheckIgnoreTagsKeyPrefixes(&providers, []string{}),
				),
			},
		},
	})
}

func TestAccNASAcctest_Provider_endpoints(t *testing.T) {
	var providers []*schema.Provider
	var endpoints strings.Builder

	// Initialize each endpoint configuration with matching name and value
	for _, endpointServiceName := range provider.EndpointServiceNames {
		endpoints.WriteString(fmt.Sprintf("%s = \"http://%s\"\n", endpointServiceName, endpointServiceName))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigEndpoints(endpoints.String()),
				Check: resource.ComposeTestCheckFunc(
					CheckEndpoints(&providers),
				),
			},
		},
	})
}

func TestAccNASAcctest_ProviderIgnoreTags_emptyBlock(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigIgnoreTagsEmptyConfigurationBlock(),
				Check: resource.ComposeTestCheckFunc(
					CheckIgnoreTagsKeys(&providers, []string{}),
					CheckIgnoreTagsKeyPrefixes(&providers, []string{}),
				),
			},
		},
	})
}

func TestAccNASAcctest_ProviderIgnoreTagsKeyPrefixes_none(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigIgnoreTagsKeyPrefixes0(),
				Check: resource.ComposeTestCheckFunc(
					CheckIgnoreTagsKeyPrefixes(&providers, []string{}),
				),
			},
		},
	})
}

func TestAccNASAcctest_ProviderIgnoreTagsKeyPrefixes_one(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigIgnoreTagsKeyPrefixes3("test"),
				Check: resource.ComposeTestCheckFunc(
					CheckIgnoreTagsKeyPrefixes(&providers, []string{"test"}),
				),
			},
		},
	})
}

func TestAccNASAcctest_ProviderIgnoreTagsKeyPrefixes_multiple(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigIgnoreTagsKeyPrefixes2("test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					CheckIgnoreTagsKeyPrefixes(&providers, []string{"test1", "test2"}),
				),
			},
		},
	})
}

func TestAccNASAcctest_ProviderIgnoreTagsKeys_none(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigIgnoreTagsKeys0(),
				Check: resource.ComposeTestCheckFunc(
					CheckIgnoreTagsKeys(&providers, []string{}),
				),
			},
		},
	})
}

func TestAccNASAcctest_ProviderIgnoreTagsKeys_one(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigIgnoreTagsKeys1("test"),
				Check: resource.ComposeTestCheckFunc(
					CheckIgnoreTagsKeys(&providers, []string{"test"}),
				),
			},
		},
	})
}

func TestAccNASAcctest_ProviderIgnoreTagsKeys_multiple(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigIgnoreTagsKeys2("test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					CheckIgnoreTagsKeys(&providers, []string{"test1", "test2"}),
				),
			},
		},
	})
}

func TestAccNASAcctest_ProviderRegion_awsC2S(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigRegion("us-iso-east-1"), // lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					CheckDNSSuffix(&providers, "c2s.ic.gov"),
					CheckPartition(&providers, "aws-iso"),
					CheckReverseDNSPrefix(&providers, "gov.ic.c2s"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccNASAcctest_ProviderRegion_awsChina(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigRegion("cn-northwest-1"), // lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					CheckDNSSuffix(&providers, "amazonaws.com.cn"),
					CheckPartition(&providers, "aws-cn"),
					CheckReverseDNSPrefix(&providers, "cn.com.amazonaws"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccNASAcctest_ProviderRegion_awsCommercial(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigRegion("us-west-2"), // lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					CheckDNSSuffix(&providers, "amazonaws.com"),
					CheckPartition(&providers, "aws"),
					CheckReverseDNSPrefix(&providers, "com.amazonaws"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccNASAcctest_ProviderRegion_awsGovCloudUs(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigRegion("us-gov-west-1"), // lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					CheckDNSSuffix(&providers, "amazonaws.com"),
					CheckPartition(&providers, "aws-us-gov"),
					CheckReverseDNSPrefix(&providers, "com.amazonaws"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccNASAcctest_ProviderRegion_awsSC2S(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigRegion("us-isob-east-1"), // lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					CheckDNSSuffix(&providers, "sc2s.sgov.gov"),
					CheckPartition(&providers, "aws-iso-b"),
					CheckReverseDNSPrefix(&providers, "gov.sgov.sc2s"),
				),
				PlanOnly: true,
			},
		},
	})
}

func TestAccNASAcctest_ProviderAssumeRole_empty(t *testing.T) {
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSProviderConfigAssumeRoleEmpty,
				Check: resource.ComposeTestCheckFunc(
					CheckAwsCallerIdentityAccountId("data.aws_caller_identity.current"),
				),
			},
		},
	})
}
