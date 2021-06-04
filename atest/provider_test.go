package atest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	awsprovider "github.com/terraform-providers/terraform-provider-aws/provider"
)

func TestProvider(t *testing.T) {
	if err := awsprovider.Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = awsprovider.Provider()
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

			if got, want := awsprovider.ReverseDNS(testCase.input), testCase.expected; got != want {
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderDefaultTagsEmptyConfigurationBlock(),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderDefaultTags_Tags0(),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderDefaultTags_Tags1("test", "value"),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderDefaultTags_Tags2("test1", "value1", "test2", "value2"),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderDefaultAndIgnoreTagsEmptyConfigurationBlock(),
				Check: resource.ComposeTestCheckFunc(
					CheckProviderDefaultTags_Tags(&providers, map[string]string{}),
					CheckProviderIgnoreTagsKeys(&providers, []string{}),
					CheckProviderIgnoreTagsKeyPrefixes(&providers, []string{}),
				),
			},
		},
	})
}

func TestAccAWSProvider_Endpoints(t *testing.T) {
	var providers []*schema.Provider
	var endpoints strings.Builder

	// Initialize each endpoint configuration with matching name and value
	for _, endpointServiceName := range awsprovider.EndpointServiceNames {
		endpoints.WriteString(fmt.Sprintf("%s = \"http://%s\"\n", endpointServiceName, endpointServiceName))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { PreCheck(t) },
		ErrorCheck:        ErrorCheck(t),
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderEndpoints(endpoints.String()),
				Check: resource.ComposeTestCheckFunc(
					CheckProviderEndpoints(&providers),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderIgnoreTagsEmptyConfigurationBlock(),
				Check: resource.ComposeTestCheckFunc(
					CheckProviderIgnoreTagsKeys(&providers, []string{}),
					CheckProviderIgnoreTagsKeyPrefixes(&providers, []string{}),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderIgnoreTagsKeyPrefixes0(),
				Check: resource.ComposeTestCheckFunc(
					CheckProviderIgnoreTagsKeyPrefixes(&providers, []string{}),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ProviderConfigIgnoreTagsKeyPrefixes1WithSkip("test"),
				Check: resource.ComposeTestCheckFunc(
					CheckProviderIgnoreTagsKeyPrefixes(&providers, []string{"test"}),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderIgnoreTagsKeyPrefixes2("test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					CheckProviderIgnoreTagsKeyPrefixes(&providers, []string{"test1", "test2"}),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderIgnoreTagsKeys0(),
				Check: resource.ComposeTestCheckFunc(
					CheckProviderIgnoreTagsKeys(&providers, []string{}),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ProviderConfigIgnoreTagsKeys1WithSkip("test"),
				Check: resource.ComposeTestCheckFunc(
					CheckProviderIgnoreTagsKeys(&providers, []string{"test"}),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderIgnoreTagsKeys2("test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					CheckProviderIgnoreTagsKeys(&providers, []string{"test1", "test2"}),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderRegion("us-iso-east-1"), // lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					CheckProviderDnsSuffix(&providers, "c2s.ic.gov"),
					CheckProviderPartition(&providers, "aws-iso"),
					CheckProviderReverseDnsPrefix(&providers, "gov.ic.c2s"),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderRegion("cn-northwest-1"), // lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					CheckProviderDnsSuffix(&providers, "amazonaws.com.cn"),
					CheckProviderPartition(&providers, "aws-cn"),
					CheckProviderReverseDnsPrefix(&providers, "cn.com.amazonaws"),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderRegion("us-west-2"), // lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					CheckProviderDnsSuffix(&providers, "amazonaws.com"),
					CheckProviderPartition(&providers, "aws"),
					CheckProviderReverseDnsPrefix(&providers, "com.amazonaws"),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderRegion("us-gov-west-1"), // lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					CheckProviderDnsSuffix(&providers, "amazonaws.com"),
					CheckProviderPartition(&providers, "aws-us-gov"),
					CheckProviderReverseDnsPrefix(&providers, "com.amazonaws"),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: ConfigProviderRegion("us-isob-east-1"), // lintignore:AWSAT003
				Check: resource.ComposeTestCheckFunc(
					CheckProviderDnsSuffix(&providers, "sc2s.sgov.gov"),
					CheckProviderPartition(&providers, "aws-iso-b"),
					CheckProviderReverseDnsPrefix(&providers, "gov.sgov.sc2s"),
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
		ProviderFactories: ProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: CheckAWSProviderConfigAssumeRoleEmpty,
				Check: resource.ComposeTestCheckFunc(
					CheckAwsCallerIdentityAccountId("data.aws_caller_identity.current"),
				),
			},
		},
	})
}
