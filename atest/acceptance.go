package atest

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
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/envvar"
	organizationsfinder "github.com/terraform-providers/terraform-provider-aws/aws/service/organizations/finder"
	stsfinder "github.com/terraform-providers/terraform-provider-aws/aws/service/sts/finder"
	"github.com/terraform-providers/terraform-provider-aws/provider"
	awsprovider "github.com/terraform-providers/terraform-provider-aws/provider"
)

const (
	// Provider name for single configuration testing
	ProviderNameAws = "aws"

	// Provider name for alternate configuration testing
	ProviderNameAwsAlternate = "awsalternate"

	// Provider name for alternate account and alternate region configuration testing
	ProviderNameAwsAlternateAccountAlternateRegion = "awsalternateaccountalternateregion"

	// Provider name for alternate account and same region configuration testing
	ProviderNameAwsAlternateAccountSameRegion = "awsalternateaccountsameregion"

	// Provider name for same account and alternate region configuration testing
	ProviderNameAwsSameAccountAlternateRegion = "awssameaccountalternateregion"

	// Provider name for third configuration testing
	ProviderNameAwsThird = "awsthird"
)

const RFC3339RegexPattern = `^[0-9]{4}-(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[01])[Tt]([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](\.[0-9]+)?([Zz]|([+-]([01][0-9]|2[0-3]):[0-5][0-9]))$`

// Skip implements a wrapper for (*testing.T).Skip() to prevent unused linting reports
//
// Reference: https://github.com/dominikh/go-tools/issues/633#issuecomment-606560616
var Skip = func(t *testing.T, message string) {
	t.Skip(message)
}

// Providers is a static map containing only the main provider instance.
//
// Deprecated: Terraform Plugin SDK version 2 uses TestCase.ProviderFactories
// but supports this value in TestCase.Providers for backwards compatibility.
// In the future Providers: Providers will be changed to
// ProviderFactories: ProviderFactories
var Providers map[string]*schema.Provider

// ProviderFactories is a static map containing only the main provider instance
//
// Use other ProviderFactories functions, such as ProviderFactoriesAlternate,
// for tests requiring special provider configurations.
var ProviderFactories map[string]func() (*schema.Provider, error)

// Provider is the "main" provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// PreCheck(t) must be called before using this provider instance.
var Provider *schema.Provider

// ProviderConfigure ensures Provider is only configured once
//
// The PreCheck(t) function is invoked for every test and this prevents
// extraneous reconfiguration to the same values each time. However, this does
// not prevent reconfiguration that may happen should the address of
// Provider be errantly reused in ProviderFactories.
var ProviderConfigure sync.Once

func init() {
	Provider = awsprovider.Provider()

	Providers = map[string]*schema.Provider{
		ProviderNameAws: Provider,
	}

	// Always allocate a new provider instance each invocation, otherwise gRPC
	// ProviderConfigure() can overwrite configuration during concurrent testing.
	ProviderFactories = map[string]func() (*schema.Provider, error){
		ProviderNameAws: func() (*schema.Provider, error) { return awsprovider.Provider(), nil }, //nolint:unparam
	}
}

// ProviderFactoriesInit creates ProviderFactories for the provider under testing.
func ProviderFactoriesInit(providers *[]*schema.Provider, providerNames []string) map[string]func() (*schema.Provider, error) {
	var factories = make(map[string]func() (*schema.Provider, error), len(providerNames))

	for _, name := range providerNames {
		p := awsprovider.Provider()

		factories[name] = func() (*schema.Provider, error) { //nolint:unparam
			return p, nil
		}

		if providers != nil {
			*providers = append(*providers, p)
		}
	}

	return factories
}

// atest.ProviderFactoriesInternal creates ProviderFactories for provider configuration testing
//
// This should only be used for AWSProvider_ tests which need to
// reference the provider instance itself. Other testing should use
// ProviderFactories or other related functions.
func ProviderFactoriesInternal(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error) {
	return ProviderFactoriesInit(providers, []string{ProviderNameAws})
}

// ProviderFactoriesAlternate creates ProviderFactories for cross-account and cross-region configurations
//
// For cross-region testing: Typically paired with MultipleRegionPreCheck and AlternateRegionProviderConfig.
//
// For cross-account testing: Typically paired with AlternateAccountPreCheck and AlternateAccountProviderConfig.
func ProviderFactoriesAlternate(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error) {
	return ProviderFactoriesInit(providers, []string{
		ProviderNameAws,
		ProviderNameAwsAlternate,
	})
}

// ProviderFactoriesAlternateAccountAndAlternateRegion creates ProviderFactories for cross-account and cross-region configurations
//
// Usage typically paired with MultipleRegionPreCheck, AlternateAccountPreCheck,
// and AlternateAccountAndAlternateRegionProviderConfig.
func ProviderFactoriesAlternateAccountAndAlternateRegion(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error) {
	return ProviderFactoriesInit(providers, []string{
		ProviderNameAws,
		ProviderNameAwsAlternateAccountAlternateRegion,
		ProviderNameAwsAlternateAccountSameRegion,
		ProviderNameAwsSameAccountAlternateRegion,
	})
}

// ProviderFactoriesMultipleRegion creates ProviderFactories for the number of region configurations
//
// Usage typically paired with MultipleRegionPreCheck and MultipleRegionProviderConfig.
func ProviderFactoriesMultipleRegion(providers *[]*schema.Provider, regions int) map[string]func() (*schema.Provider, error) {
	providerNames := []string{
		ProviderNameAws,
		ProviderNameAwsAlternate,
	}

	if regions >= 3 {
		providerNames = append(providerNames, ProviderNameAwsThird)
	}

	return ProviderFactoriesInit(providers, providerNames)
}

// PreCheck verifies and sets required provider testing configuration
//
// This PreCheck function should be present in every acceptance test. It allows
// test configurations to omit a provider configuration with region and ensures
// testing functions that attempt to call AWS APIs are previously configured.
//
// These verifications and configuration are preferred at this level to prevent
// provider developers from experiencing less clear errors for every test.
func PreCheck(t *testing.T) {
	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	ProviderConfigure.Do(func() {
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
		region := Region()
		os.Setenv(envvar.AwsDefaultRegion, region)

		err := Provider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
		if err != nil {
			t.Fatal(err)
		}
	})
}

// AwsProviderAccountID returns the account ID of an AWS provider
func ProviderAccountID(provider *schema.Provider) string {
	if provider == nil {
		log.Print("[DEBUG] Unable to read account ID from test provider: empty provider")
		return ""
	}
	if provider.Meta() == nil {
		log.Print("[DEBUG] Unable to read account ID from test provider: unconfigured provider")
		return ""
	}
	client, ok := provider.Meta().(*awsprovider.AWSClient)
	if !ok {
		log.Print("[DEBUG] Unable to read account ID from test provider: non-AWS or unconfigured AWS provider")
		return ""
	}
	return client.AccountID
}

// CheckResourceAttrAccountID ensures the Terraform state exactly matches the account ID
func CheckAttrAccountID(resourceName, attributeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestCheckResourceAttr(resourceName, attributeName, AccountID())(s)
	}
}

// CheckResourceAttrRegionalARN ensures the Terraform state exactly matches a formatted ARN with region
func CheckAttrRegionalARN(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			AccountID: AccountID(),
			Partition: Partition(),
			Region:    Region(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// CheckResourceAttrRegionalARNNoAccount ensures the Terraform state exactly matches a formatted ARN with region but without account ID
func CheckAttrRegionalARNNoAccount(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			Partition: Partition(),
			Region:    Region(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// CheckResourceAttrRegionalARNAccountID ensures the Terraform state exactly matches a formatted ARN with region and specific account ID
func CheckAttrRegionalARNAccountID(resourceName, attributeName, arnService, accountID, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			AccountID: accountID,
			Partition: Partition(),
			Region:    Region(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// CheckResourceAttrRegionalHostname ensures the Terraform state exactly matches a formatted DNS hostname with region and partition DNS suffix
func CheckAttrRegionalHostname(resourceName, attributeName, serviceName, hostnamePrefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hostname := fmt.Sprintf("%s.%s.%s.%s", hostnamePrefix, serviceName, Region(), PartitionDNSSuffix())

		return resource.TestCheckResourceAttr(resourceName, attributeName, hostname)(s)
	}
}

// CheckResourceAttrRegionalHostnameService ensures the Terraform state exactly matches a service DNS hostname with region and partition DNS suffix
//
// For example: ec2.us-west-2.amazonaws.com
func CheckAttrRegionalHostnameService(resourceName, attributeName, serviceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hostname := fmt.Sprintf("%s.%s.%s", serviceName, Region(), PartitionDNSSuffix())

		return resource.TestCheckResourceAttr(resourceName, attributeName, hostname)(s)
	}
}

// CheckResourceAttrRegionalReverseDnsService ensures the Terraform state exactly matches a service reverse DNS hostname with region and partition DNS suffix
//
// For example: com.amazonaws.us-west-2.s3
func CheckAttrRegionalReverseDnsService(resourceName, attributeName, serviceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		reverseDns := fmt.Sprintf("%s.%s.%s", PartitionReverseDNSPrefix(), Region(), serviceName)

		return resource.TestCheckResourceAttr(resourceName, attributeName, reverseDns)(s)
	}
}

// CheckResourceAttrHostnameWithPort ensures the Terraform state regexp matches a formatted DNS hostname with prefix, partition DNS suffix, and given port
func CheckAttrHostnameWithPort(resourceName, attributeName, serviceName, hostnamePrefix string, port int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// kafka broker example: "ec2-12-345-678-901.compute-1.amazonaws.com:2345"
		hostname := fmt.Sprintf("%s.%s.%s:%d", hostnamePrefix, serviceName, PartitionDNSSuffix(), port)

		return resource.TestCheckResourceAttr(resourceName, attributeName, hostname)(s)
	}
}

// CheckResourceAttrPrivateDnsName ensures the Terraform state exactly matches a private DNS name
//
// For example: ip-172-16-10-100.us-west-2.compute.internal
func CheckAttrPrivateDnsName(resourceName, attributeName string, privateIpAddress **string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		privateDnsName := fmt.Sprintf("ip-%s.%s", awsprovider.EC2DashIP(**privateIpAddress), awsprovider.EC2RegionalPrivateDnsSuffix(Region()))

		return resource.TestCheckResourceAttr(resourceName, attributeName, privateDnsName)(s)
	}
}

// MatchResourceAttrAccountID ensures the Terraform state regexp matches an account ID
func MatchAttrAccountID(resourceName, attributeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestMatchResourceAttr(resourceName, attributeName, regexp.MustCompile(`^\d{12}$`))(s)
	}
}

// MatchResourceAttrRegionalARN ensures the Terraform state regexp matches a formatted ARN with region
func MatchAttrRegionalARN(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: AccountID(),
			Partition: Partition(),
			Region:    Region(),
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

// MatchResourceAttrRegionalARNNoAccount ensures the Terraform state regexp matches a formatted ARN with region but without account ID
func MatchAttrRegionalARNNoAccount(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			Partition: Partition(),
			Region:    Region(),
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

// MatchResourceAttrRegionalARNAccountID ensures the Terraform state regexp matches a formatted ARN with region and specific account ID
func MatchAttrRegionalARNAccountID(resourceName, attributeName, arnService, accountID string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: accountID,
			Partition: Partition(),
			Region:    Region(),
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

// MatchResourceAttrRegionalHostname ensures the Terraform state regexp matches a formatted DNS hostname with region and partition DNS suffix
func MatchAttrRegionalHostname(resourceName, attributeName, serviceName string, hostnamePrefixRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hostnameRegexpPattern := fmt.Sprintf("%s\\.%s\\.%s\\.%s$", hostnamePrefixRegexp.String(), serviceName, Region(), PartitionDNSSuffix())

		hostnameRegexp, err := regexp.Compile(hostnameRegexpPattern)

		if err != nil {
			return fmt.Errorf("Unable to compile hostname regexp (%s): %w", hostnameRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, hostnameRegexp)(s)
	}
}

// CheckResourceAttrGlobalARN ensures the Terraform state exactly matches a formatted ARN without region
func CheckAttrGlobalARN(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			AccountID: AccountID(),
			Partition: Partition(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// CheckResourceAttrGlobalARNNoAccount ensures the Terraform state exactly matches a formatted ARN without region or account ID
func CheckAttrGlobalARNNoAccount(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			Partition: Partition(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// CheckResourceAttrGlobalARNAccountID ensures the Terraform state exactly matches a formatted ARN without region and with specific account ID
func CheckAttrGlobalARNAccountID(resourceName, attributeName, accountID, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			AccountID: accountID,
			Partition: Partition(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// MatchResourceAttrGlobalARN ensures the Terraform state regexp matches a formatted ARN without region
func MatchAttrGlobalARN(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: AccountID(),
			Partition: Partition(),
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

// CheckResourceAttrRegionalARNIgnoreRegionAndAccount ensures the Terraform state exactly matches a formatted ARN with region without specifying the region or account
func CheckAttrRegionalARNIgnoreRegionAndAccount(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: awsprovider.AccountIDRegexpInternalPattern,
			Partition: Partition(),
			Region:    awsprovider.RegionRegexpInternalPattern,
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

// MatchResourceAttrGlobalARNNoAccount ensures the Terraform state regexp matches a formatted ARN without region or account ID
func MatchAttrGlobalARNNoAccount(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			Partition: Partition(),
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

// CheckResourceAttrRfc3339 ensures the Terraform state matches a RFC3339 value
// This TestCheckFunc will likely be moved to the Terraform Plugin SDK in the future.
func CheckAttrRfc3339(resourceName, attributeName string) resource.TestCheckFunc {
	return resource.TestMatchResourceAttr(resourceName, attributeName, regexp.MustCompile(RFC3339RegexPattern))
}

// CheckListHasSomeElementAttrPair is a TestCheckFunc which validates that the collection on the left has an element with an attribute value
// matching the value on the left
// Based on TestCheckResourceAttrPair from the Terraform SDK testing framework
func CheckListHasSomeElementAttrPair(nameFirst string, resourceAttr string, elementAttr string, nameSecond string, keySecond string) resource.TestCheckFunc {
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

// CheckResourceAttrEquivalentJSON is a TestCheckFunc that compares a JSON value with an expected value. Both JSON
// values are normalized before being compared.
func CheckAttrEquivalentJSON(resourceName, attributeName, expectedJSON string) resource.TestCheckFunc {
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

// GetAccountID returns the account ID of Provider
// Must be used within a resource.TestCheckFunc
func AccountID() string {
	return ProviderAccountID(Provider)
}

func Region() string {
	return envvar.GetWithDefault(envvar.AwsDefaultRegion, endpoints.UsWest2RegionID)
}

func AlternateRegion() string {
	return envvar.GetWithDefault(envvar.AwsAlternateRegion, endpoints.UsEast1RegionID)
}

func ThirdRegion() string {
	return envvar.GetWithDefault(envvar.AwsThirdRegion, endpoints.UsEast2RegionID)
}

func Partition() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), Region()); ok {
		return partition.ID()
	}
	return "aws"
}

func PartitionDNSSuffix() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), Region()); ok {
		return partition.DNSSuffix()
	}
	return "amazonaws.com"
}

func PartitionReverseDNSPrefix() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), Region()); ok {
		return awsprovider.ReverseDNS(partition.DNSSuffix())
	}

	return "com.amazonaws"
}

func AlternateRegionPartition() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), AlternateRegion()); ok {
		return partition.ID()
	}
	return "aws"
}

func ThirdRegionPartition() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), ThirdRegion()); ok {
		return partition.ID()
	}
	return "aws"
}

func PreCheckAlternateAccount(t *testing.T) {
	envvar.TestSkipIfAllEmpty(t, []string{envvar.AwsAlternateProfile, envvar.AwsAlternateAccessKeyId}, "credentials for running acceptance testing in alternate AWS account")

	if os.Getenv(envvar.AwsAlternateAccessKeyId) != "" {
		envvar.TestSkipIfEmpty(t, envvar.AwsAlternateSecretAccessKey, "static credentials value when using "+envvar.AwsAlternateAccessKeyId)
	}
}

func PreCheckEC2VPCOnly(t *testing.T) {
	client := Provider.Meta().(*awsprovider.AWSClient)
	platforms := client.SupportedPlatforms
	region := client.Region
	if awsprovider.HasEC2Classic(platforms) {
		t.Skipf("This test can only in regions without EC2 Classic, platforms available in %s: %q",
			region, platforms)
	}
}

func PreCheckPartitionService(serviceId string, t *testing.T) {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), Region()); ok {
		if _, ok := partition.Services()[serviceId]; !ok {
			t.Skip(fmt.Sprintf("skipping tests; partition %s does not support %s service", partition.ID(), serviceId))
		}
	}
}

func PreCheckMultipleRegion(t *testing.T, regions int) {
	if Region() == AlternateRegion() {
		t.Fatalf("%s and %s must be set to different values for acceptance tests", envvar.AwsDefaultRegion, envvar.AwsAlternateRegion)
	}

	if Partition() != AlternateRegionPartition() {
		t.Fatalf("%s partition (%s) does not match %s partition (%s)", envvar.AwsAlternateRegion, AlternateRegionPartition(), envvar.AwsDefaultRegion, Partition())
	}

	if regions >= 3 {
		if Region() == ThirdRegion() {
			t.Fatalf("%s and %s must be set to different values for acceptance tests", envvar.AwsDefaultRegion, envvar.AwsThirdRegion)
		}

		if AlternateRegion() == ThirdRegion() {
			t.Fatalf("%s and %s must be set to different values for acceptance tests", envvar.AwsAlternateRegion, envvar.AwsThirdRegion)
		}

		if Partition() != ThirdRegionPartition() {
			t.Fatalf("%s partition (%s) does not match %s partition (%s)", envvar.AwsThirdRegion, ThirdRegionPartition(), envvar.AwsDefaultRegion, Partition())
		}
	}

	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), Region()); ok {
		if len(partition.Regions()) < regions {
			t.Skipf("skipping tests; partition includes %d regions, %d expected", len(partition.Regions()), regions)
		}
	}
}

// RegionPreCheck checks that the test region is the specified region.
func PreCheckRegion(t *testing.T, region string) {
	if Region() != region {
		t.Skipf("skipping tests; %s (%s) does not equal %s", envvar.AwsDefaultRegion, Region(), region)
	}
}

// PartitionPreCheck checks that the test partition is the specified partition.
func PreCheckPartition(partition string, t *testing.T) {
	if Partition() != partition {
		t.Skipf("skipping tests; current partition (%s) does not equal %s", Partition(), partition)
	}
}

func PreCheckOrganizationsAccount(t *testing.T) {
	conn := Provider.Meta().(*awsprovider.AWSClient).OrganizationsConn
	input := &organizations.DescribeOrganizationInput{}
	_, err := conn.DescribeOrganization(input)
	if tfawserr.ErrMessageContains(err, organizations.ErrCodeAWSOrganizationsNotInUseException, "") {
		return
	}
	if err != nil {
		t.Fatalf("error describing AWS Organization: %s", err)
	}
	t.Skip("skipping tests; this AWS account must not be an existing member of an AWS Organization")
}

func PreCheckOrganizationsEnabled(t *testing.T) {
	conn := Provider.Meta().(*awsprovider.AWSClient).OrganizationsConn
	input := &organizations.DescribeOrganizationInput{}
	_, err := conn.DescribeOrganization(input)
	if tfawserr.ErrMessageContains(err, organizations.ErrCodeAWSOrganizationsNotInUseException, "") {
		t.Skip("this AWS account must be an existing member of an AWS Organization")
	}
	if err != nil {
		t.Fatalf("error describing AWS Organization: %s", err)
	}
}

func PreCheckOrganizationManagementAccount(t *testing.T) {
	organization, err := organizationsfinder.Organization(Provider.Meta().(*awsprovider.AWSClient).OrganizationsConn)

	if err != nil {
		t.Fatalf("error describing AWS Organization: %s", err)
	}

	callerIdentity, err := stsfinder.CallerIdentity(Provider.Meta().(*awsprovider.AWSClient).STSConn)

	if err != nil {
		t.Fatalf("error getting current identity: %s", err)
	}

	if aws.StringValue(organization.MasterAccountId) != aws.StringValue(callerIdentity.Account) {
		t.Skip("this AWS account must be the management account of an AWS Organization")
	}
}

func PreCheckIamServiceLinkedRole(t *testing.T, pathPrefix string) {
	conn := Provider.Meta().(*awsprovider.AWSClient).IAMConn

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

	if PreCheckSkipError(err) {
		t.Skipf("skipping tests: %s", err)
	}

	if err != nil {
		t.Fatalf("error listing IAM roles: %s", err)
	}

	if role == nil {
		t.Skipf("skipping tests; missing IAM service-linked role %s. Please create the role and retry", pathPrefix)
	}
}

func ConfigProviderAlternateAccount() string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "awsalternate" {
  access_key = %[1]q
  profile    = %[2]q
  secret_key = %[3]q
}
`, os.Getenv(envvar.AwsAlternateAccessKeyId), os.Getenv(envvar.AwsAlternateProfile), os.Getenv(envvar.AwsAlternateSecretAccessKey))
}

func ConfigProviderAlternateAccountAlternateRegion() string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "awsalternate" {
  access_key = %[1]q
  profile    = %[2]q
  region     = %[3]q
  secret_key = %[4]q
}
`, os.Getenv(envvar.AwsAlternateAccessKeyId), os.Getenv(envvar.AwsAlternateProfile), AlternateRegion(), os.Getenv(envvar.AwsAlternateSecretAccessKey))
}

// When testing needs to distinguish a second region and second account in the same region
// e.g. cross-region functionality with RAM shared subnets
func ConfigProviderAlternateAccountAndAlternateRegion() string {
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
`, os.Getenv(envvar.AwsAlternateAccessKeyId), os.Getenv(envvar.AwsAlternateProfile), AlternateRegion(), os.Getenv(envvar.AwsAlternateSecretAccessKey))
}

// Deprecated: Use MultipleRegionProviderConfig instead
func ConfigProviderAlternateRegion() string {
	return ConfigProviderNamedRegional(ProviderNameAwsAlternate, AlternateRegion())
}

func ConfigProviderMultipleRegion(regions int) string {
	var config strings.Builder

	config.WriteString(ConfigProviderNamedRegional(ProviderNameAwsAlternate, AlternateRegion()))

	if regions >= 3 {
		config.WriteString(ConfigProviderNamedRegional(ProviderNameAwsThird, ThirdRegion()))
	}

	return config.String()
}

func ConfigProviderDefaultAndIgnoreTagsKeyPrefixes1(key1, value1, keyPrefix1 string) string {
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

func ConfigProviderDefaultAndIgnoreTagsKeys1(key1, value1 string) string {
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

func ConfigProviderIgnoreTagsKeyPrefixes1(keyPrefix1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    key_prefixes = [%[1]q]
  }
}
`, keyPrefix1)
}

func ConfigProviderIgnoreTagsKeys1(key1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    keys = [%[1]q]
  }
}
`, key1)
}

// NamedRegionalProviderConfig creates a new provider named configuration with a region.
//
// This can be used to build multiple provider configuration testing.
func ConfigProviderNamedRegional(providerName string, region string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider %[1]q {
  region = %[2]q
}
`, providerName, region)
}

// RegionalProviderConfig creates a new provider configuration with a region.
//
// This can only be used for single provider configuration testing as it
// overwrites the "aws" provider configuration.
func ConfigProviderRegional(region string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  region = %[1]q
}
`, region)
}

func RegionProviderFunc(region string, providers *[]*schema.Provider) func() *schema.Provider {
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

			// Ignore if Meta is not awsprovider.AWSClient, this will happen for other providers
			client, ok := provider.Meta().(*awsprovider.AWSClient)
			if !ok {
				log.Printf("[DEBUG] Skipping non-AWS provider")
				continue
			}

			clientRegion := client.Region
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

func DeleteResource(resource *schema.Resource, d *schema.ResourceData, meta interface{}) error {
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

func CheckDisappears(provider *schema.Provider, resource *schema.Resource, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if resourceState.Primary.ID == "" {
			return fmt.Errorf("resource ID missing: %s", resourceName)
		}

		return DeleteResource(resource, resource.Data(resourceState.Primary), provider.Meta())
	}
}

func CheckWithProviders(f func(*terraform.State, *schema.Provider) error, providers *[]*schema.Provider) resource.TestCheckFunc {
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

// ErrorCheckSkipMessagesContaining skips tests based on error messages that indicate unsupported features
func ErrorCheckSkipMessagesContaining(t *testing.T, messages ...string) resource.ErrorCheckFunc {
	return func(err error) error {
		if err == nil {
			return nil
		}

		for _, message := range messages {
			errorMessage := err.Error()
			if strings.Contains(errorMessage, message) {
				t.Skipf("skipping test for %s/%s: %s", Partition(), Region(), errorMessage)
			}
		}

		return err
	}
}

type ServiceErrorCheckFunc func(*testing.T) resource.ErrorCheckFunc

var serviceErrorCheckFuncs map[string]ServiceErrorCheckFunc

func RegisterServiceErrorCheckFunc(endpointID string, f ServiceErrorCheckFunc) {
	if serviceErrorCheckFuncs == nil {
		serviceErrorCheckFuncs = make(map[string]ServiceErrorCheckFunc)
	}

	if _, ok := serviceErrorCheckFuncs[endpointID]; ok {
		// already registered
		panic(fmt.Sprintf("Cannot re-register a service! ServiceErrorCheckFunc exists for %s", endpointID)) //lintignore:R009
	}

	serviceErrorCheckFuncs[endpointID] = f
}

func ErrorCheck(t *testing.T, endpointIDs ...string) resource.ErrorCheckFunc {
	return func(err error) error {
		if err == nil {
			return nil
		}

		for _, endpointID := range endpointIDs {
			if f, ok := serviceErrorCheckFuncs[endpointID]; ok {
				ef := f(t)
				err = ef(err)
			}

			if err == nil {
				break
			}
		}

		if ErrorCheckCommon(err) {
			t.Skipf("skipping test for %s/%s: %s", Partition(), Region(), err.Error())
		}

		return err
	}
}

// NOTE: This function cannot use the standard tfawserr helpers
// as it is receiving error strings from the SDK testing framework,
// not actual error types from the resource logic.
func ErrorCheckCommon(err error) bool {
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
func PreCheckSkipError(err error) bool {
	// GovCloud has endpoints that respond with (no message provided after the error code):
	// AccessDeniedException:
	// Ignore these API endpoints that exist but are not officially enabled
	if tfawserr.ErrMessageContains(err, "AccessDeniedException", "") {
		return true
	}
	// Ignore missing API endpoints
	if tfawserr.ErrMessageContains(err, "RequestError", "send request failed") {
		return true
	}
	// Ignore unsupported API calls
	if tfawserr.ErrMessageContains(err, "UnknownOperationException", "") {
		return true
	}
	if tfawserr.ErrMessageContains(err, "UnsupportedOperation", "") {
		return true
	}
	if tfawserr.ErrMessageContains(err, "InvalidInputException", "Unknown operation") {
		return true
	}
	if tfawserr.ErrMessageContains(err, "InvalidAction", "is not valid") {
		return true
	}
	if tfawserr.ErrMessageContains(err, "InvalidAction", "Unavailable Operation") {
		return true
	}
	return false
}

func CheckProviderDnsSuffix(providers *[]*schema.Provider, expectedDnsSuffix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*awsprovider.AWSClient) == nil {
				continue
			}

			providerDnsSuffix := provider.Meta().(*awsprovider.AWSClient).DNSSuffix

			if providerDnsSuffix != expectedDnsSuffix {
				return fmt.Errorf("expected DNS Suffix (%s), got: %s", expectedDnsSuffix, providerDnsSuffix)
			}
		}

		return nil
	}
}

func CheckProviderEndpoints(providers *[]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		// Match awsprovider.AWSClient struct field names to endpoint configuration names
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
			if provider == nil || provider.Meta() == nil || provider.Meta().(*awsprovider.AWSClient) == nil {
				continue
			}

			providerClient := provider.Meta().(*awsprovider.AWSClient)

			for _, endpointServiceName := range awsprovider.EndpointServiceNames {
				providerClientField := reflect.Indirect(reflect.ValueOf(providerClient)).FieldByNameFunc(endpointFieldNameF(endpointServiceName))

				if !providerClientField.IsValid() {
					return fmt.Errorf("unable to match awsprovider.AWSClient struct field name for endpoint name: %s", endpointServiceName)
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

func CheckProviderIgnoreTagsKeyPrefixes(providers *[]*schema.Provider, expectedKeyPrefixes []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*awsprovider.AWSClient) == nil {
				continue
			}

			providerClient := provider.Meta().(*awsprovider.AWSClient)
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

func CheckProviderIgnoreTagsKeys(providers *[]*schema.Provider, expectedKeys []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*awsprovider.AWSClient) == nil {
				continue
			}

			providerClient := provider.Meta().(*awsprovider.AWSClient)
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

func CheckProviderDefaultTags_Tags(providers *[]*schema.Provider, expectedTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*awsprovider.AWSClient) == nil {
				continue
			}

			providerClient := provider.Meta().(*awsprovider.AWSClient)
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

func CheckProviderPartition(providers *[]*schema.Provider, expectedPartition string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*awsprovider.AWSClient) == nil {
				continue
			}

			providerPartition := provider.Meta().(*awsprovider.AWSClient).Partition

			if providerPartition != expectedPartition {
				return fmt.Errorf("expected DNS Suffix (%s), got: %s", expectedPartition, providerPartition)
			}
		}

		return nil
	}
}

func CheckProviderReverseDnsPrefix(providers *[]*schema.Provider, expectedReverseDnsPrefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*awsprovider.AWSClient) == nil {
				continue
			}

			providerReverseDnsPrefix := provider.Meta().(*awsprovider.AWSClient).ReverseDnsPrefix

			if providerReverseDnsPrefix != expectedReverseDnsPrefix {
				return fmt.Errorf("expected DNS Suffix (%s), got: %s", expectedReverseDnsPrefix, providerReverseDnsPrefix)
			}
		}

		return nil
	}
}

// PreCheckEc2ClassicOrHasDefaultVpcWithDefaultSubnets checks that the test region has either
// - The EC2-Classic platform available, or
// - A default VPC with default subnets.
// This check is useful to ensure that an instance can be launched without specifying a subnet.
func PreCheckEc2ClassicOrHasDefaultVpcWithDefaultSubnets(t *testing.T) {
	client := Provider.Meta().(*awsprovider.AWSClient)

	if !awsprovider.HasEC2Classic(client.SupportedPlatforms) && !(HasDefaultVpc(t) && DefaultSubnetCount(t) > 0) {
		t.Skipf("skipping tests; %s does not have EC2-Classic or a default VPC with default subnets", client.Region)
	}
}

// HasDefaultVpc returns whether the current AWS region has a default VPC.
func HasDefaultVpc(t *testing.T) bool {
	conn := Provider.Meta().(*awsprovider.AWSClient).EC2Conn

	resp, err := conn.DescribeAccountAttributes(&ec2.DescribeAccountAttributesInput{
		AttributeNames: aws.StringSlice([]string{ec2.AccountAttributeNameDefaultVpc}),
	})
	if PreCheckSkipError(err) ||
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

// DefaultSubnetCount returns the number of default subnets in the current region's default VPC.
func DefaultSubnetCount(t *testing.T) int {
	conn := Provider.Meta().(*awsprovider.AWSClient).EC2Conn

	input := &ec2.DescribeSubnetsInput{
		Filters: provider.BuildAttributeFilterList(map[string]string{
			"defaultForAz": "true",
		}),
	}
	output, err := conn.DescribeSubnets(input)
	if PreCheckSkipError(err) {
		return 0
	}
	if err != nil {
		t.Fatalf("error describing default subnets: %s", err)
	}

	return len(output.Subnets)
}

func ConfigProviderDefaultTags_Tags0() string {
	//lintignore:AT004
	return ComposeConfig(
		ProviderConfigBase,
		`
provider "aws" {
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func ConfigProviderDefaultTags_Tags1(tag1, value1 string) string {
	//lintignore:AT004
	return ComposeConfig(
		ProviderConfigBase,
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

func ConfigProviderDefaultTags_Tags2(tag1, value1, tag2, value2 string) string {
	//lintignore:AT004
	return ComposeConfig(
		ProviderConfigBase,
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

func ConfigProviderDefaultTagsEmptyConfigurationBlock() string {
	//lintignore:AT004
	return ComposeConfig(
		ProviderConfigBase,
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

func ConfigProviderDefaultAndIgnoreTagsEmptyConfigurationBlock() string {
	//lintignore:AT004
	return ComposeConfig(
		ProviderConfigBase,
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

func ConfigProviderEndpoints(endpoints string) string {
	//lintignore:AT004
	return ComposeConfig(
		ProviderConfigBase,
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

func ConfigProviderIgnoreTagsEmptyConfigurationBlock() string {
	//lintignore:AT004
	return ComposeConfig(
		ProviderConfigBase,
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

func ConfigProviderIgnoreTagsKeyPrefixes0() string {
	//lintignore:AT004
	return ComposeConfig(
		ProviderConfigBase,
		`
provider "aws" {
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func ProviderConfigIgnoreTagsKeyPrefixes1WithSkip(tagPrefix1 string) string {
	//lintignore:AT004
	return ComposeConfig(
		ProviderConfigBase,
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

func ConfigProviderIgnoreTagsKeyPrefixes2(tagPrefix1, tagPrefix2 string) string {
	//lintignore:AT004
	return ComposeConfig(
		ProviderConfigBase,
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

func ConfigProviderIgnoreTagsKeys0() string {
	//lintignore:AT004
	return ComposeConfig(
		ProviderConfigBase,
		`
provider "aws" {
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func ProviderConfigIgnoreTagsKeys1WithSkip(tag1 string) string {
	//lintignore:AT004
	return ComposeConfig(
		ProviderConfigBase,
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

func ConfigProviderIgnoreTagsKeys2(tag1, tag2 string) string {
	//lintignore:AT004
	return ComposeConfig(
		ProviderConfigBase,
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

func ConfigProviderRegion(region string) string {
	//lintignore:AT004
	return ComposeConfig(
		ProviderConfigBase,
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

func PreCheckAssumeRoleARN(t *testing.T) {
	envvar.TestSkipIfEmpty(t, envvar.TfAccAssumeRoleArn, "Amazon Resource Name (ARN) of existing IAM Role to assume for testing restricted permissions")
}

func ConfigProviderAssumeRolePolicy(policy string) string {
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

const CheckAWSProviderConfigAssumeRoleEmpty = `
provider "aws" {
  assume_role {
  }
}

data "aws_caller_identity" "current" {}
` //lintignore:AT004

const ProviderConfigBase = `
data "aws_partition" "provider_test" {}

# Required to initialize the provider
data "aws_arn" "test" {
  arn = "arn:${data.aws_partition.provider_test.partition}:s3:::test"
}
`

func CheckSortedCSV(resourceName, attributeName string) resource.TestCheckFunc {
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

// ComposeConfig can be called to concatenate multiple strings to build test configurations
func ComposeConfig(config ...string) string {
	var str strings.Builder

	for _, conf := range config {
		str.WriteString(conf)
	}

	return str.String()
}

func CheckAwsCallerIdentityAccountId(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find AccountID resource: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Account Id resource ID not set.")
		}

		expected := Provider.Meta().(*awsprovider.AWSClient).AccountID
		if rs.Primary.Attributes["account_id"] != expected {
			return fmt.Errorf("Incorrect Account ID: expected %q, got %q", expected, rs.Primary.Attributes["account_id"])
		}

		if rs.Primary.Attributes["user_id"] == "" {
			return fmt.Errorf("UserID expected to not be nil")
		}

		if rs.Primary.Attributes["arn"] == "" {
			return fmt.Errorf("ARN expected to not be nil")
		}

		return nil
	}
}
