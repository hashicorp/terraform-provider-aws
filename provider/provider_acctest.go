package aws

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/envvar"
)

const (
	// Provider name for single configuration testing
	HelperProviderNameAws = "aws"

	// Provider name for alternate configuration testing
	HelperProviderNameAwsAlternate = "awsalternate"

	// Provider name for alternate account and alternate region configuration testing
	HelperProviderNameAwsAlternateAccountAlternateRegion = "awsalternateaccountalternateregion"

	// Provider name for alternate account and same region configuration testing
	HelperProviderNameAwsAlternateAccountSameRegion = "awsalternateaccountsameregion"

	// Provider name for same account and alternate region configuration testing
	HelperProviderNameAwsSameAccountAlternateRegion = "awssameaccountalternateregion"

	// Provider name for third configuration testing
	HelperProviderNameAwsThird = "awsthird"
)

const HelperRFC3339RegexPattern = `^[0-9]{4}-(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[01])[Tt]([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](\.[0-9]+)?([Zz]|([+-]([01][0-9]|2[0-3]):[0-5][0-9]))$`

// HelperAccTestProviders is a static map containing only the main provider instance.
//
// Deprecated: Terraform Plugin SDK version 2 uses TestCase.ProviderFactories
// but supports this value in TestCase.Providers for backwards compatibility.
// In the future Providers: HelperAccTestProviders will be changed to
// ProviderFactories: HelperAccTestProviderFactories
var HelperAccTestProviders map[string]*schema.Provider

// HelperAccTestProviderFactories is a static map containing only the main provider instance
//
// Use other HelperAccTestProviderFactories functions, such as HelperAccTestProviderFactoriesAlternate,
// for tests requiring special provider configurations.
var HelperAccTestProviderFactories map[string]func() (*schema.Provider, error)

// HelperAccTestProvider is the "main" provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// HelperAccTestPreCheck(t) must be called before using this provider instance.
var HelperAccTestProvider *schema.Provider

// HelperAccTestProviderConfigure ensures HelperAccTestProvider is only configured once
//
// The HelperAccTestPreCheck(t) function is invoked for every test and this prevents
// extraneous reconfiguration to the same values each time. However, this does
// not prevent reconfiguration that may happen should the address of
// HelperAccTestProvider be errantly reused in ProviderFactories.
var HelperAccTestProviderConfigure sync.Once

func init() {
	HelperAccTestProvider = Provider()

	HelperAccTestProviders = map[string]*schema.Provider{
		HelperProviderNameAws: HelperAccTestProvider,
	}

	HelperAccTestProviders = map[string]*schema.Provider{
		HelperProviderNameAws: HelperAccTestProvider,
	}

	// Always allocate a new provider instance each invocation, otherwise gRPC
	// ProviderConfigure() can overwrite configuration during concurrent testing.
	HelperAccTestProviderFactories = map[string]func() (*schema.Provider, error){
		HelperProviderNameAws: func() (*schema.Provider, error) { return Provider(), nil }, //nolint:unparam
	}
}

// HelperAccTestProviderFactoriesInit creates ProviderFactories for the provider under testing.
func HelperAccTestProviderFactoriesInit(providers *[]*schema.Provider, providerNames []string) map[string]func() (*schema.Provider, error) {
	var factories = make(map[string]func() (*schema.Provider, error), len(providerNames))

	for _, name := range providerNames {
		p := Provider()

		factories[name] = func() (*schema.Provider, error) { //nolint:unparam
			return p, nil
		}

		if providers != nil {
			*providers = append(*providers, p)
		}
	}

	return factories
}

// HelperAccTestProviderFactoriesInternal creates ProviderFactories for provider configuration testing
//
// This should only be used for TestAccAWSProvider_ tests which need to
// reference the provider instance itself. Other testing should use
// HelperAccTestProviderFactories or other related functions.
func HelperAccTestProviderFactoriesInternal(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error) {
	return HelperAccTestProviderFactoriesInit(providers, []string{HelperProviderNameAws})
}

// HelperAccTestProviderFactoriesAlternate creates ProviderFactories for cross-account and cross-region configurations
//
// For cross-region testing: Typically paired with HelperAccTestMultipleRegionPreCheck and HelperAccTestAlternateRegionProviderConfig.
//
// For cross-account testing: Typically paired with HelperAccTestAlternateAccountPreCheck and HelperAccTestAlternateAccountProviderConfig.
func HelperAccTestProviderFactoriesAlternate(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error) {
	return HelperAccTestProviderFactoriesInit(providers, []string{
		HelperProviderNameAws,
		HelperProviderNameAwsAlternate,
	})
}

// HelperAccTestProviderFactoriesAlternateAccountAndAlternateRegion creates ProviderFactories for cross-account and cross-region configurations
//
// Usage typically paired with HelperAccTestMultipleRegionPreCheck, HelperAccTestAlternateAccountPreCheck,
// and HelperAccTestAlternateAccountAndAlternateRegionProviderConfig.
func HelperAccTestProviderFactoriesAlternateAccountAndAlternateRegion(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error) {
	return HelperAccTestProviderFactoriesInit(providers, []string{
		HelperProviderNameAws,
		HelperProviderNameAwsAlternateAccountAlternateRegion,
		HelperProviderNameAwsAlternateAccountSameRegion,
		HelperProviderNameAwsSameAccountAlternateRegion,
	})
}

// HelperAccTestProviderFactoriesMultipleRegion creates ProviderFactories for the number of region configurations
//
// Usage typically paired with HelperAccTestMultipleRegionPreCheck and HelperAccTestMultipleRegionProviderConfig.
func HelperAccTestProviderFactoriesMultipleRegion(providers *[]*schema.Provider, regions int) map[string]func() (*schema.Provider, error) {
	providerNames := []string{
		HelperProviderNameAws,
		HelperProviderNameAwsAlternate,
	}

	if regions >= 3 {
		providerNames = append(providerNames, HelperProviderNameAwsThird)
	}

	return HelperAccTestProviderFactoriesInit(providers, providerNames)
}

// HelperAccTestPreCheck verifies and sets required provider testing configuration
//
// This PreCheck function should be present in every acceptance test. It allows
// test configurations to omit a provider configuration with region and ensures
// testing functions that attempt to call AWS APIs are previously configured.
//
// These verifications and configuration are preferred at this level to prevent
// provider developers from experiencing less clear errors for every test.
func HelperAccTestPreCheck(t *testing.T) {
	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	HelperAccTestProviderConfigure.Do(func() {
		envvar.TestFailIfAllEmpty(t, []string{envvar.AwsProfile, envvar.AwsAccessKeyId, envvar.AwsContainerCredentialsFullUri}, "credentials for running acceptance testing")

		if os.Getenv(envvar.AwsAccessKeyId) != "" {
			envvar.TestFailIfEmpty(t, envvar.AwsSecretAccessKey, "static credentials value when using "+envvar.AwsAccessKeyId)
		}

		// Setting the AWS_DEFAULT_REGION environment variable here allows all tests to omit
		// a provider configuration with a region. This defaults to us-west-2 for provider
		// developer simplicity and has been in the codebase for a very long time.
		//
		// This handling must be preserved until either:
		//   * AWS_DEFAULT_REGION is required and checked above (should mention us-west-2 default)
		//   * Region is automatically handled via shared AWS configuration file and still verified
		region := HelperAccTestGetRegion()
		os.Setenv(envvar.AwsDefaultRegion, region)

		err := HelperAccTestProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
		if err != nil {
			t.Fatal(err)
		}
	})
}

// HelperAccTestAwsProviderAccountID returns the account ID of an AWS provider
func HelperAccTestAwsProviderAccountID(provider *schema.Provider) string {
	if provider == nil {
		log.Print("[DEBUG] Unable to read account ID from test provider: empty provider")
		return ""
	}
	if provider.Meta() == nil {
		log.Print("[DEBUG] Unable to read account ID from test provider: unconfigured provider")
		return ""
	}
	client, ok := provider.Meta().(*AWSClient)
	if !ok {
		log.Print("[DEBUG] Unable to read account ID from test provider: non-AWS or unconfigured AWS provider")
		return ""
	}
	return client.accountid
}

// HelperAccTestCheckResourceAttrAccountID ensures the Terraform state exactly matches the account ID
func HelperAccTestCheckResourceAttrAccountID(resourceName, attributeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestCheckResourceAttr(resourceName, attributeName, HelperAccTestGetAccountID())(s)
	}
}

// HelperAccTestCheckResourceAttrRegionalARN ensures the Terraform state exactly matches a formatted ARN with region
func HelperAccTestCheckResourceAttrRegionalARN(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			AccountID: HelperAccTestGetAccountID(),
			Partition: HelperAccTestGetPartition(),
			Region:    HelperAccTestGetRegion(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// HelperAccTestCheckResourceAttrRegionalARNNoAccount ensures the Terraform state exactly matches a formatted ARN with region but without account ID
func HelperAccTestCheckResourceAttrRegionalARNNoAccount(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			Partition: HelperAccTestGetPartition(),
			Region:    HelperAccTestGetRegion(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// HelperAccTestCheckResourceAttrRegionalARNAccountID ensures the Terraform state exactly matches a formatted ARN with region and specific account ID
func HelperAccTestCheckResourceAttrRegionalARNAccountID(resourceName, attributeName, arnService, accountID, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			AccountID: accountID,
			Partition: HelperAccTestGetPartition(),
			Region:    HelperAccTestGetRegion(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// HelperAccTestCheckResourceAttrRegionalHostname ensures the Terraform state exactly matches a formatted DNS hostname with region and partition DNS suffix
func HelperAccTestCheckResourceAttrRegionalHostname(resourceName, attributeName, serviceName, hostnamePrefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hostname := fmt.Sprintf("%s.%s.%s.%s", hostnamePrefix, serviceName, HelperAccTestGetRegion(), HelperAccTestGetPartitionDNSSuffix())

		return resource.TestCheckResourceAttr(resourceName, attributeName, hostname)(s)
	}
}

// HelperAccTestCheckResourceAttrRegionalHostnameService ensures the Terraform state exactly matches a service DNS hostname with region and partition DNS suffix
//
// For example: ec2.us-west-2.amazonaws.com
func HelperAccTestCheckResourceAttrRegionalHostnameService(resourceName, attributeName, serviceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hostname := fmt.Sprintf("%s.%s.%s", serviceName, HelperAccTestGetRegion(), HelperAccTestGetPartitionDNSSuffix())

		return resource.TestCheckResourceAttr(resourceName, attributeName, hostname)(s)
	}
}

// HelperAccTestCheckResourceAttrRegionalReverseDnsService ensures the Terraform state exactly matches a service reverse DNS hostname with region and partition DNS suffix
//
// For example: com.amazonaws.us-west-2.s3
func HelperAccTestCheckResourceAttrRegionalReverseDnsService(resourceName, attributeName, serviceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		reverseDns := fmt.Sprintf("%s.%s.%s", HelperAccTestGetPartitionReverseDNSPrefix(), HelperAccTestGetRegion(), serviceName)

		return resource.TestCheckResourceAttr(resourceName, attributeName, reverseDns)(s)
	}
}

// HelperAccTestCheckResourceAttrHostnameWithPort ensures the Terraform state regexp matches a formatted DNS hostname with prefix, partition DNS suffix, and given port
func HelperAccTestCheckResourceAttrHostnameWithPort(resourceName, attributeName, serviceName, hostnamePrefix string, port int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// kafka broker example: "ec2-12-345-678-901.compute-1.amazonaws.com:2345"
		hostname := fmt.Sprintf("%s.%s.%s:%d", hostnamePrefix, serviceName, HelperAccTestGetPartitionDNSSuffix(), port)

		return resource.TestCheckResourceAttr(resourceName, attributeName, hostname)(s)
	}
}

// HelperAccTestMatchResourceAttrAccountID ensures the Terraform state regexp matches an account ID
func HelperAccTestMatchResourceAttrAccountID(resourceName, attributeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestMatchResourceAttr(resourceName, attributeName, regexp.MustCompile(`^\d{12}$`))(s)
	}
}

// HelperAccTestMatchResourceAttrRegionalARN ensures the Terraform state regexp matches a formatted ARN with region
func HelperAccTestMatchResourceAttrRegionalARN(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: HelperAccTestGetAccountID(),
			Partition: HelperAccTestGetPartition(),
			Region:    HelperAccTestGetRegion(),
			Resource:  arnResourceRegexp.String(),
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("Unable to compile ARN regexp (%s): %w", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// HelperAccTestMatchResourceAttrRegionalARNNoAccount ensures the Terraform state regexp matches a formatted ARN with region but without account ID
func HelperAccTestMatchResourceAttrRegionalARNNoAccount(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			Partition: HelperAccTestGetPartition(),
			Region:    HelperAccTestGetRegion(),
			Resource:  arnResourceRegexp.String(),
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("Unable to compile ARN regexp (%s): %s", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// HelperAccTestMatchResourceAttrRegionalARNAccountID ensures the Terraform state regexp matches a formatted ARN with region and specific account ID
func HelperAccTestMatchResourceAttrRegionalARNAccountID(resourceName, attributeName, arnService, accountID string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: accountID,
			Partition: HelperAccTestGetPartition(),
			Region:    HelperAccTestGetRegion(),
			Resource:  arnResourceRegexp.String(),
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("Unable to compile ARN regexp (%s): %w", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// HelperAccTestMatchResourceAttrRegionalHostname ensures the Terraform state regexp matches a formatted DNS hostname with region and partition DNS suffix
func HelperAccTestMatchResourceAttrRegionalHostname(resourceName, attributeName, serviceName string, hostnamePrefixRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hostnameRegexpPattern := fmt.Sprintf("%s\\.%s\\.%s\\.%s$", hostnamePrefixRegexp.String(), serviceName, HelperAccTestGetRegion(), HelperAccTestGetPartitionDNSSuffix())

		hostnameRegexp, err := regexp.Compile(hostnameRegexpPattern)

		if err != nil {
			return fmt.Errorf("Unable to compile hostname regexp (%s): %w", hostnameRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, hostnameRegexp)(s)
	}
}

// HelperAccTestCheckResourceAttrGlobalARN ensures the Terraform state exactly matches a formatted ARN without region
func HelperAccTestCheckResourceAttrGlobalARN(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			AccountID: HelperAccTestGetAccountID(),
			Partition: HelperAccTestGetPartition(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// HelperAccTestCheckResourceAttrGlobalARNNoAccount ensures the Terraform state exactly matches a formatted ARN without region or account ID
func HelperAccTestCheckResourceAttrGlobalARNNoAccount(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			Partition: HelperAccTestGetPartition(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// HelperAccTestCheckResourceAttrGlobalARNAccountID ensures the Terraform state exactly matches a formatted ARN without region and with specific account ID
func HelperAccTestCheckResourceAttrGlobalARNAccountID(resourceName, attributeName, accountID, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			AccountID: accountID,
			Partition: HelperAccTestGetPartition(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// HelperAccTestMatchResourceAttrGlobalARN ensures the Terraform state regexp matches a formatted ARN without region
func HelperAccTestMatchResourceAttrGlobalARN(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: HelperAccTestGetAccountID(),
			Partition: HelperAccTestGetPartition(),
			Resource:  arnResourceRegexp.String(),
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("Unable to compile ARN regexp (%s): %w", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// HelperAccTestCheckResourceAttrRegionalARNIgnoreRegionAndAccount ensures the Terraform state exactly matches a formatted ARN with region without specifying the region or account
func HelperAccTestCheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: awsAccountIDRegexpInternalPattern,
			Partition: HelperAccTestGetPartition(),
			Region:    awsRegionRegexpInternalPattern,
			Resource:  arnResource,
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("Unable to compile ARN regexp (%s): %w", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// HelperAccTestMatchResourceAttrGlobalARNNoAccount ensures the Terraform state regexp matches a formatted ARN without region or account ID
func HelperAccTestMatchResourceAttrGlobalARNNoAccount(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			Partition: HelperAccTestGetPartition(),
			Resource:  arnResourceRegexp.String(),
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("Unable to compile ARN regexp (%s): %s", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// HelperAccTestCheckResourceAttrRfc3339 ensures the Terraform state matches a RFC3339 value
// This TestCheckFunc will likely be moved to the Terraform Plugin SDK in the future.
func HelperAccTestCheckResourceAttrRfc3339(resourceName, attributeName string) resource.TestCheckFunc {
	return resource.TestMatchResourceAttr(resourceName, attributeName, regexp.MustCompile(HelperRFC3339RegexPattern))
}

// HelperAccTestCheckListHasSomeElementAttrPair is a TestCheckFunc which validates that the collection on the left has an element with an attribute value
// matching the value on the left
// Based on TestCheckResourceAttrPair from the Terraform SDK testing framework
func HelperAccTestCheckListHasSomeElementAttrPair(nameFirst string, resourceAttr string, elementAttr string, nameSecond string, keySecond string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		isFirst, err := PrimaryInstanceState(s, nameFirst)
		if err != nil {
			return err
		}

		isSecond, err := PrimaryInstanceState(s, nameSecond)
		if err != nil {
			return err
		}

		vSecond, ok := isSecond.Attributes[keySecond]
		if !ok {
			return fmt.Errorf("%s: No attribute %q found", nameSecond, keySecond)
		} else if vSecond == "" {
			return fmt.Errorf("%s: No value was set on attribute %q", nameSecond, keySecond)
		}

		attrsFirst := make([]string, 0, len(isFirst.Attributes))
		attrMatcher := regexp.MustCompile(fmt.Sprintf("%s\\.\\d+\\.%s", resourceAttr, elementAttr))
		for k := range isFirst.Attributes {
			if attrMatcher.MatchString(k) {
				attrsFirst = append(attrsFirst, k)
			}
		}

		found := false
		for _, attrName := range attrsFirst {
			vFirst := isFirst.Attributes[attrName]
			if vFirst == vSecond {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s: No element of %q found with attribute %q matching value %q set on %q of %s", nameFirst, resourceAttr, elementAttr, vSecond, keySecond, nameSecond)
		}

		return nil
	}
}

// HelperAccTestCheckResourceAttrEquivalentJSON is a TestCheckFunc that compares a JSON value with an expected value. Both JSON
// values are normalized before being compared.
func HelperAccTestCheckResourceAttrEquivalentJSON(resourceName, attributeName, expectedJSON string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		is, err := PrimaryInstanceState(s, resourceName)
		if err != nil {
			return err
		}

		v, ok := is.Attributes[attributeName]
		if !ok {
			return fmt.Errorf("%s: No attribute %q found", resourceName, attributeName)
		}

		vNormal, err := structure.NormalizeJsonString(v)
		if err != nil {
			return fmt.Errorf("%s: Error normalizing JSON in %q: %w", resourceName, attributeName, err)
		}

		expectedNormal, err := structure.NormalizeJsonString(expectedJSON)
		if err != nil {
			return fmt.Errorf("Error normalizing expected JSON: %w", err)
		}

		if vNormal != expectedNormal {
			return fmt.Errorf("%s: Attribute %q expected\n%s\ngot\n%s", resourceName, attributeName, expectedJSON, v)
		}
		return nil
	}
}

// Copied and inlined from the SDK testing code
func PrimaryInstanceState(s *terraform.State, name string) (*terraform.InstanceState, error) {
	rs, ok := s.RootModule().Resources[name]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", name)
	}

	is := rs.Primary
	if is == nil {
		return nil, fmt.Errorf("No primary instance: %s", name)
	}

	return is, nil
}

// HelperAccTestGetAccountID returns the account ID of HelperAccTestProvider
// Must be used within a resource.TestCheckFunc
func HelperAccTestGetAccountID() string {
	return HelperAccTestAwsProviderAccountID(HelperAccTestProvider)
}

func HelperAccTestGetRegion() string {
	return envvar.GetWithDefault(envvar.AwsDefaultRegion, endpoints.UsWest2RegionID)
}

func HelperAccTestGetAlternateRegion() string {
	return envvar.GetWithDefault(envvar.AwsAlternateRegion, endpoints.UsEast1RegionID)
}

func HelperAccTestGetThirdRegion() string {
	return envvar.GetWithDefault(envvar.AwsThirdRegion, endpoints.UsEast2RegionID)
}

func HelperAccTestGetPartition() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), HelperAccTestGetRegion()); ok {
		return partition.ID()
	}
	return "aws"
}

func HelperAccTestGetPartitionDNSSuffix() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), HelperAccTestGetRegion()); ok {
		return partition.DNSSuffix()
	}
	return "amazonaws.com"
}

func HelperAccTestGetPartitionReverseDNSPrefix() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), HelperAccTestGetRegion()); ok {
		return ReverseDns(partition.DNSSuffix())
	}

	return "com.amazonaws"
}

func HelperAccTestGetAlternateRegionPartition() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), HelperAccTestGetAlternateRegion()); ok {
		return partition.ID()
	}
	return "aws"
}

func HelperAccTestGetThirdRegionPartition() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), HelperAccTestGetThirdRegion()); ok {
		return partition.ID()
	}
	return "aws"
}

func HelperAccTestAlternateAccountPreCheck(t *testing.T) {
	envvar.TestFailIfAllEmpty(t, []string{envvar.AwsAlternateProfile, envvar.AwsAlternateAccessKeyId}, "credentials for running acceptance testing in alternate AWS account")

	if os.Getenv(envvar.AwsAlternateAccessKeyId) != "" {
		envvar.TestFailIfEmpty(t, envvar.AwsAlternateSecretAccessKey, "static credentials value when using "+envvar.AwsAlternateAccessKeyId)
	}
}

func HelperAccTestEC2VPCOnlyPreCheck(t *testing.T) {
	client := HelperAccTestProvider.Meta().(*AWSClient)
	platforms := client.supportedplatforms
	region := client.region
	if hasEc2Classic(platforms) {
		t.Skipf("This test can only in regions without EC2 Classic, platforms available in %s: %q",
			region, platforms)
	}
}

func HelperAccTestPartitionHasServicePreCheck(serviceId string, t *testing.T) {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), HelperAccTestGetRegion()); ok {
		if _, ok := partition.Services()[serviceId]; !ok {
			t.Skip(fmt.Sprintf("skipping tests; partition %s does not support %s service", partition.ID(), serviceId))
		}
	}
}

func HelperAccTestMultipleRegionPreCheck(t *testing.T, regions int) {
	if HelperAccTestGetRegion() == HelperAccTestGetAlternateRegion() {
		t.Fatalf("%s and %s must be set to different values for acceptance tests", envvar.AwsDefaultRegion, envvar.AwsAlternateRegion)
	}

	if HelperAccTestGetPartition() != HelperAccTestGetAlternateRegionPartition() {
		t.Fatalf("%s partition (%s) does not match %s partition (%s)", envvar.AwsAlternateRegion, HelperAccTestGetAlternateRegionPartition(), envvar.AwsDefaultRegion, HelperAccTestGetPartition())
	}

	if regions >= 3 {
		if HelperAccTestGetRegion() == HelperAccTestGetThirdRegion() {
			t.Fatalf("%s and %s must be set to different values for acceptance tests", envvar.AwsDefaultRegion, envvar.AwsThirdRegion)
		}

		if HelperAccTestGetAlternateRegion() == HelperAccTestGetThirdRegion() {
			t.Fatalf("%s and %s must be set to different values for acceptance tests", envvar.AwsAlternateRegion, envvar.AwsThirdRegion)
		}

		if HelperAccTestGetPartition() != HelperAccTestGetThirdRegionPartition() {
			t.Fatalf("%s partition (%s) does not match %s partition (%s)", envvar.AwsThirdRegion, HelperAccTestGetThirdRegionPartition(), envvar.AwsDefaultRegion, HelperAccTestGetPartition())
		}
	}

	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), HelperAccTestGetRegion()); ok {
		if len(partition.Regions()) < regions {
			t.Skipf("skipping tests; partition includes %d regions, %d expected", len(partition.Regions()), regions)
		}
	}
}

// HelperAccTestRegionPreCheck checks that the test region is the specified region.
func HelperAccTestRegionPreCheck(t *testing.T, region string) {
	if HelperAccTestGetRegion() != region {
		t.Skipf("skipping tests; %s (%s) does not equal %s", envvar.AwsDefaultRegion, HelperAccTestGetRegion(), region)
	}
}

// HelperAccTestPartitionPreCheck checks that the test partition is the specified partition.
func HelperAccTestPartitionPreCheck(partition string, t *testing.T) {
	if HelperAccTestGetPartition() != partition {
		t.Skipf("skipping tests; current partition (%s) does not equal %s", HelperAccTestGetPartition(), partition)
	}
}

func HelperAccTestOrganizationsAccountPreCheck(t *testing.T) {
	conn := HelperAccTestProvider.Meta().(*AWSClient).organizationsconn
	input := &organizations.DescribeOrganizationInput{}
	_, err := conn.DescribeOrganization(input)
	if isAWSErr(err, organizations.ErrCodeAWSOrganizationsNotInUseException, "") {
		return
	}
	if err != nil {
		t.Fatalf("error describing AWS Organization: %s", err)
	}
	t.Skip("skipping tests; this AWS account must not be an existing member of an AWS Organization")
}

func HelperAccTestOrganizationsEnabledPreCheck(t *testing.T) {
	conn := HelperAccTestProvider.Meta().(*AWSClient).organizationsconn
	input := &organizations.DescribeOrganizationInput{}
	_, err := conn.DescribeOrganization(input)
	if isAWSErr(err, organizations.ErrCodeAWSOrganizationsNotInUseException, "") {
		t.Skip("this AWS account must be an existing member of an AWS Organization")
	}
	if err != nil {
		t.Fatalf("error describing AWS Organization: %s", err)
	}
}

func HelperAccTestPreCheckIamServiceLinkedRole(t *testing.T, pathPrefix string) {
	conn := HelperAccTestProvider.Meta().(*AWSClient).iamconn

	input := &iam.ListRolesInput{
		PathPrefix: aws.String(pathPrefix),
	}

	var role *iam.Role
	err := conn.ListRolesPages(input, func(page *iam.ListRolesOutput, lastPage bool) bool {
		for _, r := range page.Roles {
			role = r
			break
		}

		return !lastPage
	})

	if HelperAccTestPreCheckSkipError(err) {
		t.Skipf("skipping tests: %s", err)
	}

	if err != nil {
		t.Fatalf("error listing IAM roles: %s", err)
	}

	if role == nil {
		t.Skipf("skipping tests; missing IAM service-linked role %s. Please create the role and retry", pathPrefix)
	}
}

func HelperAccTestAlternateAccountProviderConfig() string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "awsalternate" {
  access_key = %[1]q
  profile    = %[2]q
  secret_key = %[3]q
}
`, os.Getenv(envvar.AwsAlternateAccessKeyId), os.Getenv(envvar.AwsAlternateProfile), os.Getenv(envvar.AwsAlternateSecretAccessKey))
}

func HelperAccTestAlternateAccountAlternateRegionProviderConfig() string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "awsalternate" {
  access_key = %[1]q
  profile    = %[2]q
  region     = %[3]q
  secret_key = %[4]q
}
`, os.Getenv(envvar.AwsAlternateAccessKeyId), os.Getenv(envvar.AwsAlternateProfile), HelperAccTestGetAlternateRegion(), os.Getenv(envvar.AwsAlternateSecretAccessKey))
}

// When testing needs to distinguish a second region and second account in the same region
// e.g. cross-region functionality with RAM shared subnets
func HelperAccTestAlternateAccountAndAlternateRegionProviderConfig() string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "awsalternateaccountalternateregion" {
  access_key = %[1]q
  profile    = %[2]q
  region     = %[3]q
  secret_key = %[4]q
}

provider "awsalternateaccountsameregion" {
  access_key = %[1]q
  profile    = %[2]q
  secret_key = %[4]q
}

provider "awssameaccountalternateregion" {
  region = %[3]q
}
`, os.Getenv(envvar.AwsAlternateAccessKeyId), os.Getenv(envvar.AwsAlternateProfile), HelperAccTestGetAlternateRegion(), os.Getenv(envvar.AwsAlternateSecretAccessKey))
}

// Deprecated: Use HelperAccTestMultipleRegionProviderConfig instead
func HelperAccTestAlternateRegionProviderConfig() string {
	return HelperAccTestNamedRegionalProviderConfig(HelperProviderNameAwsAlternate, HelperAccTestGetAlternateRegion())
}

func HelperAccTestMultipleRegionProviderConfig(regions int) string {
	var config strings.Builder

	config.WriteString(HelperAccTestNamedRegionalProviderConfig(HelperProviderNameAwsAlternate, HelperAccTestGetAlternateRegion()))

	if regions >= 3 {
		config.WriteString(HelperAccTestNamedRegionalProviderConfig(HelperProviderNameAwsThird, HelperAccTestGetThirdRegion()))
	}

	return config.String()
}

func HelperAccTestProviderConfigDefaultAndIgnoreTagsKeyPrefixes1(key1, value1, keyPrefix1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  default_tags {
    tags = {
      %q = %q
    }
  }
  ignore_tags {
    key_prefixes = [%q]
  }
}
`, key1, value1, keyPrefix1)
}

func HelperAccTestProviderConfigDefaultAndIgnoreTagsKeys1(key1, value1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  default_tags {
    tags = {
      %[1]q = %q
    }
  }
  ignore_tags {
    keys = [%[1]q]
  }
}
`, key1, value1)
}

func HelperAccTestProviderConfigIgnoreTagsKeyPrefixes1(keyPrefix1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    key_prefixes = [%[1]q]
  }
}
`, keyPrefix1)
}

func HelperAccTestProviderConfigIgnoreTagsKeys1(key1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    keys = [%[1]q]
  }
}
`, key1)
}

// HelperAccTestNamedRegionalProviderConfig creates a new provider named configuration with a region.
//
// This can be used to build multiple provider configuration testing.
func HelperAccTestNamedRegionalProviderConfig(providerName string, region string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider %[1]q {
  region = %[2]q
}
`, providerName, region)
}

// HelperAccTestRegionalProviderConfig creates a new provider configuration with a region.
//
// This can only be used for single provider configuration testing as it
// overwrites the "aws" provider configuration.
func HelperAccTestRegionalProviderConfig(region string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  region = %[1]q
}
`, region)
}

func HelperAccTestAwsRegionProviderFunc(region string, providers *[]*schema.Provider) func() *schema.Provider {
	return func() *schema.Provider {
		if region == "" {
			log.Println("[DEBUG] No region given")
			return nil
		}
		if providers == nil {
			log.Println("[DEBUG] No providers given")
			return nil
		}

		log.Printf("[DEBUG] Checking providers for AWS region: %s", region)
		for _, provider := range *providers {
			// Ignore if Meta is empty, this can happen for validation providers
			if provider == nil || provider.Meta() == nil {
				log.Printf("[DEBUG] Skipping empty provider")
				continue
			}

			// Ignore if Meta is not AWSClient, this will happen for other providers
			client, ok := provider.Meta().(*AWSClient)
			if !ok {
				log.Printf("[DEBUG] Skipping non-AWS provider")
				continue
			}

			clientRegion := client.region
			log.Printf("[DEBUG] Checking AWS provider region %q against %q", clientRegion, region)
			if clientRegion == region {
				log.Printf("[DEBUG] Found AWS provider with region: %s", region)
				return provider
			}
		}

		log.Printf("[DEBUG] No suitable provider found for %q in %d providers", region, len(*providers))
		return nil
	}
}

func HelperAccTestDeleteResource(resource *schema.Resource, d *schema.ResourceData, meta interface{}) error {
	if resource.DeleteContext != nil || resource.DeleteWithoutTimeout != nil {
		var diags diag.Diagnostics

		if resource.DeleteContext != nil {
			diags = resource.DeleteContext(context.Background(), d, meta)
		} else {
			diags = resource.DeleteWithoutTimeout(context.Background(), d, meta)
		}

		for i := range diags {
			if diags[i].Severity == diag.Error {
				return fmt.Errorf("error deleting resource: %s", diags[i].Summary)
			}
		}

		return nil
	}

	return resource.Delete(d, meta)
}

func HelperAccTestCheckResourceDisappears(provider *schema.Provider, resource *schema.Resource, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if resourceState.Primary.ID == "" {
			return fmt.Errorf("resource ID missing: %s", resourceName)
		}

		return HelperAccTestDeleteResource(resource, resource.Data(resourceState.Primary), provider.Meta())
	}
}

func HelperAccTestCheckWithProviders(f func(*terraform.State, *schema.Provider) error, providers *[]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		numberOfProviders := len(*providers)
		for i, provider := range *providers {
			if provider.Meta() == nil {
				log.Printf("[DEBUG] Skipping empty provider %d (total: %d)", i, numberOfProviders)
				continue
			}
			log.Printf("[DEBUG] Calling check with provider %d (total: %d)", i, numberOfProviders)
			if err := f(s, provider); err != nil {
				return err
			}
		}
		return nil
	}
}

// HelperAccTestErrorCheckSkipMessagesContaining skips tests based on error messages that indicate unsupported features
func HelperAccTestErrorCheckSkipMessagesContaining(t *testing.T, messages ...string) resource.ErrorCheckFunc {
	return func(err error) error {
		if err == nil {
			return nil
		}

		for _, message := range messages {
			errorMessage := err.Error()
			if strings.Contains(errorMessage, message) {
				t.Skipf("skipping test for %s/%s: %s", HelperAccTestGetPartition(), HelperAccTestGetRegion(), errorMessage)
			}
		}

		return err
	}
}

type HelperAccTestServiceErrorCheckFunc func(*testing.T) resource.ErrorCheckFunc

var HelperAccTestServiceErrorCheckFuncs map[string]HelperAccTestServiceErrorCheckFunc

func HelperAccTestRegisterServiceErrorCheckFunc(endpointID string, f HelperAccTestServiceErrorCheckFunc) {
	if HelperAccTestServiceErrorCheckFuncs == nil {
		HelperAccTestServiceErrorCheckFuncs = make(map[string]HelperAccTestServiceErrorCheckFunc)
	}

	if _, ok := HelperAccTestServiceErrorCheckFuncs[endpointID]; ok {
		// already registered
		panic(fmt.Sprintf("Cannot re-register a service! ServiceErrorCheckFunc exists for %s", endpointID)) //lintignore:R009
	}

	HelperAccTestServiceErrorCheckFuncs[endpointID] = f
}

func HelperAccTestErrorCheck(t *testing.T, endpointIDs ...string) resource.ErrorCheckFunc {
	return func(err error) error {
		if err == nil {
			return nil
		}

		for _, endpointID := range endpointIDs {
			if f, ok := HelperAccTestServiceErrorCheckFuncs[endpointID]; ok {
				ef := f(t)
				err = ef(err)
			}

			if err == nil {
				break
			}
		}

		if HelperAccTestErrorCheckCommon(err) {
			t.Skipf("skipping test for %s/%s: %s", HelperAccTestGetPartition(), HelperAccTestGetRegion(), err.Error())
		}

		return err
	}
}

// NOTE: This function cannot use the standard tfawserr helpers
// as it is receiving error strings from the SDK testing framework,
// not actual error types from the resource logic.
func HelperAccTestErrorCheckCommon(err error) bool {
	if strings.Contains(err.Error(), "is not supported in this") {
		return true
	}

	if strings.Contains(err.Error(), "is currently not supported") {
		return true
	}

	if strings.Contains(err.Error(), "InvalidAction") {
		return true
	}

	if strings.Contains(err.Error(), "Unknown operation") {
		return true
	}

	if strings.Contains(err.Error(), "UnknownOperationException") {
		return true
	}

	if strings.Contains(err.Error(), "UnsupportedOperation") {
		return true
	}

	return false
}

// Check service API call error for reasons to skip acceptance testing
// These include missing API endpoints and unsupported API calls
func HelperAccTestPreCheckSkipError(err error) bool {
	// GovCloud has endpoints that respond with (no message provided after the error code):
	// AccessDeniedException:
	// Ignore these API endpoints that exist but are not officially enabled
	if isAWSErr(err, "AccessDeniedException", "") {
		return true
	}
	// Ignore missing API endpoints
	if isAWSErr(err, "RequestError", "send request failed") {
		return true
	}
	// Ignore unsupported API calls
	if isAWSErr(err, "UnknownOperationException", "") {
		return true
	}
	if isAWSErr(err, "UnsupportedOperation", "") {
		return true
	}
	if isAWSErr(err, "InvalidInputException", "Unknown operation") {
		return true
	}
	if isAWSErr(err, "InvalidAction", "is not valid") {
		return true
	}
	if isAWSErr(err, "InvalidAction", "Unavailable Operation") {
		return true
	}
	return false
}

func HelperAccTestCheckAWSProviderDnsSuffix(providers *[]*schema.Provider, expectedDnsSuffix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*AWSClient) == nil {
				continue
			}

			providerDnsSuffix := provider.Meta().(*AWSClient).dnsSuffix

			if providerDnsSuffix != expectedDnsSuffix {
				return fmt.Errorf("expected DNS Suffix (%s), got: %s", expectedDnsSuffix, providerDnsSuffix)
			}
		}

		return nil
	}
}

func HelperAccTestCheckAWSProviderEndpoints(providers *[]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		// Match AWSClient struct field names to endpoint configuration names
		endpointFieldNameF := func(endpoint string) func(string) bool {
			return func(name string) bool {
				switch endpoint {
				case "applicationautoscaling":
					endpoint = "appautoscaling"
				case "budgets":
					endpoint = "budget"
				case "cloudformation":
					endpoint = "cf"
				case "cloudhsm":
					endpoint = "cloudhsmv2"
				case "cognitoidentity":
					endpoint = "cognito"
				case "configservice":
					endpoint = "config"
				case "cur":
					endpoint = "costandusagereport"
				case "directconnect":
					endpoint = "dx"
				case "lexmodels":
					endpoint = "lexmodel"
				case "route53":
					endpoint = "r53"
				case "sdb":
					endpoint = "simpledb"
				case "serverlessrepo":
					endpoint = "serverlessapplicationrepository"
				case "servicecatalog":
					endpoint = "sc"
				case "servicediscovery":
					endpoint = "sd"
				case "stepfunctions":
					endpoint = "sfn"
				}

				return name == fmt.Sprintf("%sconn", endpoint)
			}
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*AWSClient) == nil {
				continue
			}

			providerClient := provider.Meta().(*AWSClient)

			for _, endpointServiceName := range endpointServiceNames {
				providerClientField := reflect.Indirect(reflect.ValueOf(providerClient)).FieldByNameFunc(endpointFieldNameF(endpointServiceName))

				if !providerClientField.IsValid() {
					return fmt.Errorf("unable to match AWSClient struct field name for endpoint name: %s", endpointServiceName)
				}

				actualEndpoint := reflect.Indirect(reflect.Indirect(providerClientField).FieldByName("Config").FieldByName("Endpoint")).String()
				expectedEndpoint := fmt.Sprintf("http://%s", endpointServiceName)

				if actualEndpoint != expectedEndpoint {
					return fmt.Errorf("expected endpoint (%s) value (%s), got: %s", endpointServiceName, expectedEndpoint, actualEndpoint)
				}
			}
		}

		return nil
	}
}

func HelperAccTestCheckAWSProviderIgnoreTagsKeyPrefixes(providers *[]*schema.Provider, expectedKeyPrefixes []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*AWSClient) == nil {
				continue
			}

			providerClient := provider.Meta().(*AWSClient)
			ignoreTagsConfig := providerClient.IgnoreTagsConfig

			if ignoreTagsConfig == nil || ignoreTagsConfig.KeyPrefixes == nil {
				if len(expectedKeyPrefixes) != 0 {
					return fmt.Errorf("expected key_prefixes (%d) length, got: 0", len(expectedKeyPrefixes))
				}

				continue
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
		}

		return nil
	}
}

func HelperAccTestCheckAWSProviderIgnoreTagsKeys(providers *[]*schema.Provider, expectedKeys []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*AWSClient) == nil {
				continue
			}

			providerClient := provider.Meta().(*AWSClient)
			ignoreTagsConfig := providerClient.IgnoreTagsConfig

			if ignoreTagsConfig == nil || ignoreTagsConfig.Keys == nil {
				if len(expectedKeys) != 0 {
					return fmt.Errorf("expected keys (%d) length, got: 0", len(expectedKeys))
				}

				continue
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
		}

		return nil
	}
}

func HelperAccTestCheckProviderDefaultTags_Tags(providers *[]*schema.Provider, expectedTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*AWSClient) == nil {
				continue
			}

			providerClient := provider.Meta().(*AWSClient)
			defaultTagsConfig := providerClient.DefaultTagsConfig

			if defaultTagsConfig == nil || len(defaultTagsConfig.Tags) == 0 {
				if len(expectedTags) != 0 {
					return fmt.Errorf("expected keys (%d) length, got: 0", len(expectedTags))
				}

				continue
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
		}

		return nil
	}
}

func HelperAccTestCheckAWSProviderPartition(providers *[]*schema.Provider, expectedPartition string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*AWSClient) == nil {
				continue
			}

			providerPartition := provider.Meta().(*AWSClient).partition

			if providerPartition != expectedPartition {
				return fmt.Errorf("expected DNS Suffix (%s), got: %s", expectedPartition, providerPartition)
			}
		}

		return nil
	}
}

func HelperAccTestCheckAWSProviderReverseDnsPrefix(providers *[]*schema.Provider, expectedReverseDnsPrefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*AWSClient) == nil {
				continue
			}

			providerReverseDnsPrefix := provider.Meta().(*AWSClient).reverseDnsPrefix

			if providerReverseDnsPrefix != expectedReverseDnsPrefix {
				return fmt.Errorf("expected DNS Suffix (%s), got: %s", expectedReverseDnsPrefix, providerReverseDnsPrefix)
			}
		}

		return nil
	}
}

// HelperAccTestPreCheckEc2ClassicOrHasDefaultVpcWithDefaultSubnets checks that the test region has either
// - The EC2-Classic platform available, or
// - A default VPC with default subnets.
// This check is useful to ensure that an instance can be launched without specifying a subnet.
func HelperAccTestPreCheckEc2ClassicOrHasDefaultVpcWithDefaultSubnets(t *testing.T) {
	client := HelperAccTestProvider.Meta().(*AWSClient)

	if !hasEc2Classic(client.supportedplatforms) && !(HelperAccTestHasDefaultVpc(t) && HelperAccTestDefaultSubnetCount(t) > 0) {
		t.Skipf("skipping tests; %s does not have EC2-Classic or a default VPC with default subnets", client.region)
	}
}

// HelperAccTestHasDefaultVpc returns whether the current AWS region has a default VPC.
func HelperAccTestHasDefaultVpc(t *testing.T) bool {
	conn := HelperAccTestProvider.Meta().(*AWSClient).ec2conn

	resp, err := conn.DescribeAccountAttributes(&ec2.DescribeAccountAttributesInput{
		AttributeNames: aws.StringSlice([]string{ec2.AccountAttributeNameDefaultVpc}),
	})
	if HelperAccTestPreCheckSkipError(err) ||
		len(resp.AccountAttributes) == 0 ||
		len(resp.AccountAttributes[0].AttributeValues) == 0 ||
		aws.StringValue(resp.AccountAttributes[0].AttributeValues[0].AttributeValue) == "none" {
		return false
	}
	if err != nil {
		t.Fatalf("error describing EC2 account attributes: %s", err)
	}

	return true
}

// HelperAccTestDefaultSubnetCount returns the number of default subnets in the current region's default VPC.
func HelperAccTestDefaultSubnetCount(t *testing.T) int {
	conn := HelperAccTestProvider.Meta().(*AWSClient).ec2conn

	input := &ec2.DescribeSubnetsInput{
		Filters: buildEC2AttributeFilterList(map[string]string{
			"defaultForAz": "true",
		}),
	}
	output, err := conn.DescribeSubnets(input)
	if HelperAccTestPreCheckSkipError(err) {
		return 0
	}
	if err != nil {
		t.Fatalf("error describing default subnets: %s", err)
	}

	return len(output.Subnets)
}

func HelperAccTestAWSProviderConfigDefaultTags_Tags0() string {
	//lintignore:AT004
	return HelperAccTestComposeConfig(
		HelperAccTestProviderConfigBase,
		`
provider "aws" {
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func HelperAccTestAWSProviderConfigDefaultTags_Tags1(tag1, value1 string) string {
	//lintignore:AT004
	return HelperAccTestComposeConfig(
		HelperAccTestProviderConfigBase,
		fmt.Sprintf(`
provider "aws" {
  default_tags {
    tags = {
      %q = %q
    }
  }

  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`, tag1, value1))
}

func HelperAccTestAWSProviderConfigDefaultTags_Tags2(tag1, value1, tag2, value2 string) string {
	//lintignore:AT004
	return HelperAccTestComposeConfig(
		HelperAccTestProviderConfigBase,
		fmt.Sprintf(`
provider "aws" {
  default_tags {
    tags = {
      %q = %q
      %q = %q
    }
  }

  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`, tag1, value1, tag2, value2))
}

func HelperAccTestAWSProviderConfigDefaultTagsEmptyConfigurationBlock() string {
	//lintignore:AT004
	return HelperAccTestComposeConfig(
		HelperAccTestProviderConfigBase,
		`
provider "aws" {
  default_tags {}

  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func HelperAccTestAWSProviderConfigDefaultAndIgnoreTagsEmptyConfigurationBlock() string {
	//lintignore:AT004
	return HelperAccTestComposeConfig(
		HelperAccTestProviderConfigBase,
		`
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

func HelperAccTestAWSProviderConfigEndpoints(endpoints string) string {
	//lintignore:AT004
	return HelperAccTestComposeConfig(
		HelperAccTestProviderConfigBase,
		fmt.Sprintf(`
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

func HelperAccTestAWSProviderConfigIgnoreTagsEmptyConfigurationBlock() string {
	//lintignore:AT004
	return HelperAccTestComposeConfig(
		HelperAccTestProviderConfigBase,
		`
provider "aws" {
  ignore_tags {}

  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func HelperAccTestAWSProviderConfigIgnoreTagsKeyPrefixes0() string {
	//lintignore:AT004
	return HelperAccTestComposeConfig(
		HelperAccTestProviderConfigBase,
		`
provider "aws" {
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func HelperAccTestAWSProviderConfigIgnoreTagsKeyPrefixes1(tagPrefix1 string) string {
	//lintignore:AT004
	return HelperAccTestComposeConfig(
		HelperAccTestProviderConfigBase,
		fmt.Sprintf(`
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

func HelperAccTestAWSProviderConfigIgnoreTagsKeyPrefixes2(tagPrefix1, tagPrefix2 string) string {
	//lintignore:AT004
	return HelperAccTestComposeConfig(
		HelperAccTestProviderConfigBase,
		fmt.Sprintf(`
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

func HelperAccTestAWSProviderConfigIgnoreTagsKeys0() string {
	//lintignore:AT004
	return HelperAccTestComposeConfig(
		HelperAccTestProviderConfigBase,
		`
provider "aws" {
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func HelperAccTestAWSProviderConfigIgnoreTagsKeys1(tag1 string) string {
	//lintignore:AT004
	return HelperAccTestComposeConfig(
		HelperAccTestProviderConfigBase,
		fmt.Sprintf(`
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

func HelperAccTestAWSProviderConfigIgnoreTagsKeys2(tag1, tag2 string) string {
	//lintignore:AT004
	return HelperAccTestComposeConfig(
		HelperAccTestProviderConfigBase,
		fmt.Sprintf(`
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

func HelperAccTestAWSProviderConfigRegion(region string) string {
	//lintignore:AT004
	return HelperAccTestComposeConfig(
		HelperAccTestProviderConfigBase,
		fmt.Sprintf(`
provider "aws" {
  region                      = %[1]q
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`, region))
}

func HelperAccTestAssumeRoleARNPreCheck(t *testing.T) {
	envvar.TestSkipIfEmpty(t, envvar.TfAccAssumeRoleArn, "Amazon Resource Name (ARN) of existing IAM Role to assume for testing restricted permissions")
}

func HelperAccTestProviderConfigAssumeRolePolicy(policy string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  assume_role {
    role_arn = %q
    policy   = %q
  }
}
`, os.Getenv(envvar.TfAccAssumeRoleArn), policy)
}

const HelperAccTestCheckAWSProviderConfigAssumeRoleEmpty = `
provider "aws" {
  assume_role {
  }
}

data "aws_caller_identity" "current" {}
` //lintignore:AT004

const HelperAccTestProviderConfigBase = `
data "aws_partition" "provider_test" {}

# Required to initialize the provider
data "aws_arn" "test" {
  arn = "arn:${data.aws_partition.provider_test.partition}:s3:::test"
}
`

func HelperAccTestCheckResourceAttrIsSortedCsv(resourceName, attributeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		is, err := PrimaryInstanceState(s, resourceName)
		if err != nil {
			return err
		}

		v, ok := is.Attributes[attributeName]
		if !ok {
			return fmt.Errorf("%s: No attribute %q found", resourceName, attributeName)
		}

		splitV := strings.Split(v, ",")
		if !sort.StringsAreSorted(splitV) {
			return fmt.Errorf("%s: Expected attribute %q to be sorted, got %q", resourceName, attributeName, v)
		}

		return nil
	}
}

// HelperAccTestComposeConfig can be called to concatenate multiple strings to build test configurations
func HelperAccTestComposeConfig(config ...string) string {
	var str strings.Builder

	for _, conf := range config {
		str.WriteString(conf)
	}

	return str.String()
}
