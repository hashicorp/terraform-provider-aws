package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
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

			if got, want := ReverseDns(testCase.input), testCase.expected; got != want {
				t.Errorf("got: %s, expected: %s", got, want)
			}
		})
	}
}

func TestAccAWSProvider_DefaultTags_EmptyConfigurationBlock(t *testing.T) {
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

func TestAccAWSProvider_DefaultTags_Tags_None(t *testing.T) {
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

func TestAccAWSProvider_DefaultTags_Tags_One(t *testing.T) {
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

func TestAccAWSProvider_DefaultTags_Tags_Multiple(t *testing.T) {
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

func TestAccAWSProvider_DefaultAndIgnoreTags_EmptyConfigurationBlocks(t *testing.T) {
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

func TestAccAWSProvider_Endpoints(t *testing.T) {
	var providers []*schema.Provider
	var endpoints strings.Builder

	// Initialize each endpoint configuration with matching name and value
	for _, endpointServiceName := range endpointServiceNames {
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

func TestAccAWSProvider_IgnoreTags_EmptyConfigurationBlock(t *testing.T) {
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

func TestAccAWSProvider_IgnoreTags_KeyPrefixes_None(t *testing.T) {
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

func TestAccAWSProvider_IgnoreTags_KeyPrefixes_One(t *testing.T) {
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

func TestAccAWSProvider_IgnoreTags_KeyPrefixes_Multiple(t *testing.T) {
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

func TestAccAWSProvider_IgnoreTags_Keys_None(t *testing.T) {
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

func TestAccAWSProvider_IgnoreTags_Keys_One(t *testing.T) {
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

func TestAccAWSProvider_IgnoreTags_Keys_Multiple(t *testing.T) {
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

func TestAccAWSProvider_Region_AwsC2S(t *testing.T) {
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

func TestAccAWSProvider_Region_AwsChina(t *testing.T) {
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

func TestAccAWSProvider_Region_AwsCommercial(t *testing.T) {
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

func TestAccAWSProvider_Region_AwsGovCloudUs(t *testing.T) {
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

func TestAccAWSProvider_Region_AwsSC2S(t *testing.T) {
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

func TestAccAWSProvider_AssumeRole_Empty(t *testing.T) {
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
					testAccCheckAwsCallerIdentityAccountId("data.aws_caller_identity.current"),
				),
			},
		},
	})
}
