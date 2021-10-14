package acctest

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/envvar"
	organizationsfinder "github.com/hashicorp/terraform-provider-aws/aws/internal/service/organizations/finder"
	stsfinder "github.com/hashicorp/terraform-provider-aws/aws/internal/service/sts/finder"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

const (
	// Provider name for single configuration testing
	ProviderName = "aws"

	// Provider name for alternate configuration testing
	ProviderNameAlternate = "awsalternate"

	// Provider name for alternate account and alternate region configuration testing
	ProviderNameAwsAlternateAccountAlternateRegion = "awsalternateaccountalternateregion"

	// Provider name for alternate account and same region configuration testing
	ProviderNameAwsAlternateAccountSameRegion = "awsalternateaccountsameregion"

	// Provider name for same account and alternate region configuration testing
	ProviderNameAwsSameAccountAlternateRegion = "awssameaccountalternateregion"

	// Provider name for third configuration testing
	ProviderNameThird = "awsthird"

	ResourcePrefix = "tf-acc-test"
)

const RFC3339RegexPattern = `^[0-9]{4}-(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[01])[Tt]([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](\.[0-9]+)?([Zz]|([+-]([01][0-9]|2[0-3]):[0-5][0-9]))$`
const awsRegionRegexp = `[a-z]{2}(-[a-z]+)+-\d`
const awsAccountIDRegexp = `(aws|\d{12})`

// Skip implements a wrapper for (*testing.T).Skip() to prevent unused linting reports
//
// Reference: https://github.com/dominikh/go-tools/issues/633#issuecomment-606560616
func Skip(t *testing.T, message string) {
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
// Use other ProviderFactories functions, such as FactoriesAlternate,
// for tests requiring special provider configurations.
var ProviderFactories map[string]func() (*schema.Provider, error)

// Provider is the "main" provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// PreCheck(t) must be called before using this provider instance.
var Provider *schema.Provider

// testAccProviderConfigure ensures Provider is only configured once
//
// The PreCheck(t) function is invoked for every test and this prevents
// extraneous reconfiguration to the same values each time. However, this does
// not prevent reconfiguration that may happen should the address of
// Provider be errantly reused in ProviderFactories.
var testAccProviderConfigure sync.Once

func init() {
	Provider = provider.Provider()

	Providers = map[string]*schema.Provider{
		ProviderName: Provider,
	}

	// Always allocate a new provider instance each invocation, otherwise gRPC
	// ProviderConfigure() can overwrite configuration during concurrent testing.
	ProviderFactories = map[string]func() (*schema.Provider, error){
		ProviderName: func() (*schema.Provider, error) { return provider.Provider(), nil }, //nolint:unparam
	}
}

// FactoriesInit creates ProviderFactories for the provider under testing.
func FactoriesInit(providers *[]*schema.Provider, providerNames []string) map[string]func() (*schema.Provider, error) {
	var factories = make(map[string]func() (*schema.Provider, error), len(providerNames))

	for _, name := range providerNames {
		p := provider.Provider()

		factories[name] = func() (*schema.Provider, error) { //nolint:unparam
			return p, nil
		}

		if providers != nil {
			*providers = append(*providers, p)
		}
	}

	return factories
}

// FactoriesInternal creates ProviderFactories for provider configuration testing
//
// This should only be used for TestAccAWSProvider_ tests which need to
// reference the provider instance itself. Other testing should use
// ProviderFactories or other related functions.
func FactoriesInternal(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error) {
	return FactoriesInit(providers, []string{ProviderName})
}

// FactoriesAlternate creates ProviderFactories for cross-account and cross-region configurations
//
// For cross-region testing: Typically paired with PreCheckMultipleRegion and ConfigAlternateRegionProvider.
//
// For cross-account testing: Typically paired with PreCheckAlternateAccount and ConfigAlternateAccountProvider.
func FactoriesAlternate(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error) {
	return FactoriesInit(providers, []string{
		ProviderName,
		ProviderNameAlternate,
	})
}

// FactoriesAlternateAccountAndAlternateRegion creates ProviderFactories for cross-account and cross-region configurations
//
// Usage typically paired with PreCheckMultipleRegion, PreCheckAlternateAccount,
// and ConfigAlternateAccountAndAlternateRegionProvider.
func FactoriesAlternateAccountAndAlternateRegion(providers *[]*schema.Provider) map[string]func() (*schema.Provider, error) {
	return FactoriesInit(providers, []string{
		ProviderName,
		ProviderNameAwsAlternateAccountAlternateRegion,
		ProviderNameAwsAlternateAccountSameRegion,
		ProviderNameAwsSameAccountAlternateRegion,
	})
}

// FactoriesMultipleRegion creates ProviderFactories for the number of region configurations
//
// Usage typically paired with PreCheckMultipleRegion and ConfigMultipleRegionProvider.
func FactoriesMultipleRegion(providers *[]*schema.Provider, regions int) map[string]func() (*schema.Provider, error) {
	providerNames := []string{
		ProviderName,
		ProviderNameAlternate,
	}

	if regions >= 3 {
		providerNames = append(providerNames, ProviderNameThird)
	}

	return FactoriesInit(providers, providerNames)
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
	testAccProviderConfigure.Do(func() {
		conns.FailIfAllEnvVarEmpty(t, []string{conns.EnvVarProfile, conns.EnvVarAccessKeyId, conns.EnvVarContainerCredentialsFullUri}, "credentials for running acceptance testing")

		if os.Getenv(conns.EnvVarAccessKeyId) != "" {
			conns.FailIfEnvVarEmpty(t, conns.EnvVarSecretAccessKey, "static credentials value when using "+conns.EnvVarAccessKeyId)
		}

		// Setting the AWS_DEFAULT_REGION environment variable here allows all tests to omit
		// a provider configuration with a region. This defaults to us-west-2 for provider
		// developer simplicity and has been in the codebase for a very long time.
		//
		// This handling must be preserved until either:
		//   * AWS_DEFAULT_REGION is required and checked above (should mention us-west-2 default)
		//   * Region is automatically handled via shared AWS configuration file and still verified
		region := Region()
		os.Setenv(conns.EnvVarDefaultRegion, region)

		err := Provider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
		if err != nil {
			t.Fatal(err)
		}
	})
}

// ProviderAccountID returns the account ID of an AWS provider
func ProviderAccountID(provo *schema.Provider) string {
	if provo == nil {
		log.Print("[DEBUG] Unable to read account ID from test provider: empty provider")
		return ""
	}
	if provo.Meta() == nil {
		log.Print("[DEBUG] Unable to read account ID from test provider: unconfigured provider")
		return ""
	}
	client, ok := provo.Meta().(*conns.AWSClient)
	if !ok {
		log.Print("[DEBUG] Unable to read account ID from test provider: non-AWS or unconfigured AWS provider")
		return ""
	}
	return client.AccountID
}

// CheckResourceAttrAccountID ensures the Terraform state exactly matches the account ID
func CheckResourceAttrAccountID(resourceName, attributeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestCheckResourceAttr(resourceName, attributeName, AccountID())(s)
	}
}

// CheckResourceAttrRegionalARN ensures the Terraform state exactly matches a formatted ARN with region
func CheckResourceAttrRegionalARN(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
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
func CheckResourceAttrRegionalARNNoAccount(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
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
func CheckResourceAttrRegionalARNAccountID(resourceName, attributeName, arnService, accountID, arnResource string) resource.TestCheckFunc {
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
func CheckResourceAttrRegionalHostname(resourceName, attributeName, serviceName, hostnamePrefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hostname := fmt.Sprintf("%s.%s.%s.%s", hostnamePrefix, serviceName, Region(), PartitionDNSSuffix())

		return resource.TestCheckResourceAttr(resourceName, attributeName, hostname)(s)
	}
}

// CheckResourceAttrRegionalHostnameService ensures the Terraform state exactly matches a service DNS hostname with region and partition DNS suffix
//
// For example: ec2.us-west-2.amazonaws.com
func CheckResourceAttrRegionalHostnameService(resourceName, attributeName, serviceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hostname := fmt.Sprintf("%s.%s.%s", serviceName, Region(), PartitionDNSSuffix())

		return resource.TestCheckResourceAttr(resourceName, attributeName, hostname)(s)
	}
}

// CheckResourceAttrRegionalReverseDNSService ensures the Terraform state exactly matches a service reverse DNS hostname with region and partition DNS suffix
//
// For example: com.amazonaws.us-west-2.s3
func CheckResourceAttrRegionalReverseDNSService(resourceName, attributeName, serviceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		reverseDns := fmt.Sprintf("%s.%s.%s", PartitionReverseDNSPrefix(), Region(), serviceName)

		return resource.TestCheckResourceAttr(resourceName, attributeName, reverseDns)(s)
	}
}

/*
// CheckResourceAttrHostnameWithPort ensures the Terraform state regexp matches a formatted DNS hostname with prefix, partition DNS suffix, and given port
func CheckResourceAttrHostnameWithPort(resourceName, attributeName, serviceName, hostnamePrefix string, port int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// kafka broker example: "ec2-12-345-678-901.compute-1.amazonaws.com:2345"
		hostname := fmt.Sprintf("%s.%s.%s:%d", hostnamePrefix, serviceName, PartitionDNSSuffix(), port)

		return resource.TestCheckResourceAttr(resourceName, attributeName, hostname)(s)
	}
}
*/

// CheckResourceAttrPrivateDNSName ensures the Terraform state exactly matches a private DNS name
//
// For example: ip-172-16-10-100.us-west-2.compute.internal
func CheckResourceAttrPrivateDNSName(resourceName, attributeName string, privateIpAddress **string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		privateDnsName := fmt.Sprintf("ip-%s.%s", convertIPToDashIP(**privateIpAddress), regionalPrivateDNSSuffix(Region()))

		return resource.TestCheckResourceAttr(resourceName, attributeName, privateDnsName)(s)
	}
}

// MatchResourceAttrAccountID ensures the Terraform state regexp matches an account ID
func MatchResourceAttrAccountID(resourceName, attributeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestMatchResourceAttr(resourceName, attributeName, regexp.MustCompile(`^\d{12}$`))(s)
	}
}

// MatchResourceAttrRegionalARN ensures the Terraform state regexp matches a formatted ARN with region
func MatchResourceAttrRegionalARN(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
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
func MatchResourceAttrRegionalARNNoAccount(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
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
func MatchResourceAttrRegionalARNAccountID(resourceName, attributeName, arnService, accountID string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
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
func MatchResourceAttrRegionalHostname(resourceName, attributeName, serviceName string, hostnamePrefixRegexp *regexp.Regexp) resource.TestCheckFunc {
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
func CheckResourceAttrGlobalARN(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
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
func CheckResourceAttrGlobalARNNoAccount(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
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
func CheckResourceAttrGlobalARNAccountID(resourceName, attributeName, accountID, arnService, arnResource string) resource.TestCheckFunc {
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
func MatchResourceAttrGlobalARN(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
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
func CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: awsAccountIDRegexp,
			Partition: Partition(),
			Region:    awsRegionRegexp,
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
func MatchResourceAttrGlobalARNNoAccount(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
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

// CheckResourceAttrRFC3339 ensures the Terraform state matches a RFC3339 value
// This TestCheckFunc will likely be moved to the Terraform Plugin SDK in the future.
func CheckResourceAttrRFC3339(resourceName, attributeName string) resource.TestCheckFunc {
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
func CheckResourceAttrEquivalentJSON(resourceName, attributeName, expectedJSON string) resource.TestCheckFunc {
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

// AccountID returns the account ID of Provider
// Must be used within a resource.TestCheckFunc
func AccountID() string {
	return ProviderAccountID(Provider)
}

func Region() string {
	return conns.GetEnvVarWithDefault(conns.EnvVarDefaultRegion, endpoints.UsWest2RegionID)
}

func AlternateRegion() string {
	return conns.GetEnvVarWithDefault(conns.EnvVarAlternateRegion, endpoints.UsEast1RegionID)
}

func ThirdRegion() string {
	return conns.GetEnvVarWithDefault(conns.EnvVarThirdRegion, endpoints.UsEast2RegionID)
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
		return conns.ReverseDNS(partition.DNSSuffix())
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
	conns.SkipIfAllEnvVarEmpty(t, []string{conns.EnvVarAlternateProfile, conns.EnvVarAlternateAccessKeyId}, "credentials for running acceptance testing in alternate AWS account")

	if os.Getenv(conns.EnvVarAlternateAccessKeyId) != "" {
		conns.SkipIfEnvVarEmpty(t, conns.EnvVarAlternateSecretAccessKey, "static credentials value when using "+conns.EnvVarAlternateAccessKeyId)
	}
}

func PreCheckEC2VPCOnly(t *testing.T) {
	client := Provider.Meta().(*conns.AWSClient)
	platforms := client.SupportedPlatforms
	region := client.Region
	if conns.HasEC2Classic(platforms) {
		t.Skipf("This test can only in regions without EC2 Classic, platforms available in %s: %q",
			region, platforms)
	}
}

func PreCheckPartitionHasService(serviceId string, t *testing.T) {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), Region()); ok {
		if _, ok := partition.Services()[serviceId]; !ok {
			t.Skip(fmt.Sprintf("skipping tests; partition %s does not support %s service", partition.ID(), serviceId))
		}
	}
}

func PreCheckMultipleRegion(t *testing.T, regions int) {
	if Region() == AlternateRegion() {
		t.Fatalf("%s and %s must be set to different values for acceptance tests", conns.EnvVarDefaultRegion, conns.EnvVarAlternateRegion)
	}

	if Partition() != AlternateRegionPartition() {
		t.Fatalf("%s partition (%s) does not match %s partition (%s)", conns.EnvVarAlternateRegion, AlternateRegionPartition(), conns.EnvVarDefaultRegion, Partition())
	}

	if regions >= 3 {
		if Region() == ThirdRegion() {
			t.Fatalf("%s and %s must be set to different values for acceptance tests", conns.EnvVarDefaultRegion, conns.EnvVarThirdRegion)
		}

		if AlternateRegion() == ThirdRegion() {
			t.Fatalf("%s and %s must be set to different values for acceptance tests", conns.EnvVarAlternateRegion, conns.EnvVarThirdRegion)
		}

		if Partition() != ThirdRegionPartition() {
			t.Fatalf("%s partition (%s) does not match %s partition (%s)", conns.EnvVarThirdRegion, ThirdRegionPartition(), conns.EnvVarDefaultRegion, Partition())
		}
	}

	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), Region()); ok {
		if len(partition.Regions()) < regions {
			t.Skipf("skipping tests; partition includes %d regions, %d expected", len(partition.Regions()), regions)
		}
	}
}

// PreCheckRegion checks that the test region is the specified region.
func PreCheckRegion(t *testing.T, region string) {
	if Region() != region {
		t.Skipf("skipping tests; %s (%s) does not equal %s", conns.EnvVarDefaultRegion, Region(), region)
	}
}

// PreCheckPartition checks that the test partition is the specified partition.
func PreCheckPartition(partition string, t *testing.T) {
	if Partition() != partition {
		t.Skipf("skipping tests; current partition (%s) does not equal %s", Partition(), partition)
	}
}

func PreCheckOrganizationsAccount(t *testing.T) {
	conn := Provider.Meta().(*conns.AWSClient).OrganizationsConn
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
	conn := Provider.Meta().(*conns.AWSClient).OrganizationsConn
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
	organization, err := organizationsfinder.Organization(Provider.Meta().(*conns.AWSClient).OrganizationsConn)

	if err != nil {
		t.Fatalf("error describing AWS Organization: %s", err)
	}

	callerIdentity, err := stsfinder.CallerIdentity(Provider.Meta().(*conns.AWSClient).STSConn)

	if err != nil {
		t.Fatalf("error getting current identity: %s", err)
	}

	if aws.StringValue(organization.MasterAccountId) != aws.StringValue(callerIdentity.Account) {
		t.Skip("this AWS account must be the management account of an AWS Organization")
	}
}

func PreCheckHasIAMRole(t *testing.T, roleName string) {
	conn := Provider.Meta().(*conns.AWSClient).IAMConn

	input := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}
	_, err := conn.GetRole(input)

	if tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
		t.Skipf("skipping acceptance test: required IAM role \"%s\" is not present", roleName)
	}
	if PreCheckSkipError(err) {
		t.Skipf("skipping acceptance test: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func PreCheckIAMServiceLinkedRole(t *testing.T, pathPrefix string) {
	conn := Provider.Meta().(*conns.AWSClient).IAMConn

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

func ConfigAlternateAccountProvider() string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "awsalternate" {
  access_key = %[1]q
  profile    = %[2]q
  secret_key = %[3]q
}
`, os.Getenv(conns.EnvVarAlternateAccessKeyId), os.Getenv(conns.EnvVarAlternateProfile), os.Getenv(conns.EnvVarAlternateSecretAccessKey))
}

func ConfigAlternateAccountAlternateRegionProvider() string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "awsalternate" {
  access_key = %[1]q
  profile    = %[2]q
  region     = %[3]q
  secret_key = %[4]q
}
`, os.Getenv(conns.EnvVarAlternateAccessKeyId), os.Getenv(conns.EnvVarAlternateProfile), AlternateRegion(), os.Getenv(conns.EnvVarAlternateSecretAccessKey))
}

// When testing needs to distinguish a second region and second account in the same region
// e.g. cross-region functionality with RAM shared subnets
func ConfigAlternateAccountAndAlternateRegionProvider() string {
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
`, os.Getenv(conns.EnvVarAlternateAccessKeyId), os.Getenv(conns.EnvVarAlternateProfile), AlternateRegion(), os.Getenv(conns.EnvVarAlternateSecretAccessKey))
}

// Deprecated: Use ConfigMultipleRegionProvider instead
func ConfigAlternateRegionProvider() string {
	return ConfigNamedRegionalProvider(ProviderNameAlternate, AlternateRegion())
}

func ConfigMultipleRegionProvider(regions int) string {
	var config strings.Builder

	config.WriteString(ConfigNamedRegionalProvider(ProviderNameAlternate, AlternateRegion()))

	if regions >= 3 {
		config.WriteString(ConfigNamedRegionalProvider(ProviderNameThird, ThirdRegion()))
	}

	return config.String()
}

func ConfigDefaultAndIgnoreTagsKeyPrefixes1(key1, value1, keyPrefix1 string) string {
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

func ConfigDefaultAndIgnoreTagsKeys1(key1, value1 string) string {
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

func ConfigIgnoreTagsKeyPrefixes1(keyPrefix1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    key_prefixes = [%[1]q]
  }
}
`, keyPrefix1)
}

func ConfigIgnoreTagsKeys(key1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    keys = [%[1]q]
  }
}
`, key1)
}

// ConfigNamedRegionalProvider creates a new provider named configuration with a region.
//
// This can be used to build multiple provider configuration testing.
func ConfigNamedRegionalProvider(providerName string, region string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider %[1]q {
  region = %[2]q
}
`, providerName, region)
}

// ConfigRegionalProvider creates a new provider configuration with a region.
//
// This can only be used for single provider configuration testing as it
// overwrites the "aws" provider configuration.
func ConfigRegionalProvider(region string) string {
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
		for _, provo := range *providers {
			// Ignore if Meta is empty, this can happen for validation providers
			if provo == nil || provo.Meta() == nil {
				log.Printf("[DEBUG] Skipping empty provider")
				continue
			}

			// Ignore if Meta is not conns.AWSClient, this will happen for other providers
			client, ok := provo.Meta().(*conns.AWSClient)
			if !ok {
				log.Printf("[DEBUG] Skipping non-AWS provider")
				continue
			}

			clientRegion := client.Region
			log.Printf("[DEBUG] Checking AWS provider region %q against %q", clientRegion, region)
			if clientRegion == region {
				log.Printf("[DEBUG] Found AWS provider with region: %s", region)
				return provo
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

func CheckResourceDisappears(provo *schema.Provider, resource *schema.Resource, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if resourceState.Primary.ID == "" {
			return fmt.Errorf("resource ID missing: %s", resourceName)
		}

		return DeleteResource(resource, resource.Data(resourceState.Primary), provo.Meta())
	}
}

func CheckWithProviders(f func(*terraform.State, *schema.Provider) error, providers *[]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		numberOfProviders := len(*providers)
		for i, provo := range *providers {
			if provo.Meta() == nil {
				log.Printf("[DEBUG] Skipping empty provider %d (total: %d)", i, numberOfProviders)
				continue
			}
			log.Printf("[DEBUG] Calling check with provider %d (total: %d)", i, numberOfProviders)
			if err := f(s, provo); err != nil {
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

func CheckDNSSuffix(providers *[]*schema.Provider, expectedDnsSuffix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provo := range *providers {
			if provo == nil || provo.Meta() == nil || provo.Meta().(*conns.AWSClient) == nil {
				continue
			}

			providerDnsSuffix := provo.Meta().(*conns.AWSClient).DNSSuffix

			if providerDnsSuffix != expectedDnsSuffix {
				return fmt.Errorf("expected DNS Suffix (%s), got: %s", expectedDnsSuffix, providerDnsSuffix)
			}
		}

		return nil
	}
}

func CheckEndpoints(providers *[]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		// Match conns.AWSClient struct field names to endpoint configuration names
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

		for _, provo := range *providers {
			if provo == nil || provo.Meta() == nil || provo.Meta().(*conns.AWSClient) == nil {
				continue
			}

			providerClient := provo.Meta().(*conns.AWSClient)

			for _, endpointServiceName := range provider.EndpointServiceNames {
				providerClientField := reflect.Indirect(reflect.ValueOf(providerClient)).FieldByNameFunc(endpointFieldNameF(endpointServiceName))

				if !providerClientField.IsValid() {
					return fmt.Errorf("unable to match conns.AWSClient struct field name for endpoint name: %s", endpointServiceName)
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

func CheckIgnoreTagsKeyPrefixes(providers *[]*schema.Provider, expectedKeyPrefixes []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provo := range *providers {
			if provo == nil || provo.Meta() == nil || provo.Meta().(*conns.AWSClient) == nil {
				continue
			}

			providerClient := provo.Meta().(*conns.AWSClient)
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

func CheckIgnoreTagsKeys(providers *[]*schema.Provider, expectedKeys []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provo := range *providers {
			if provo == nil || provo.Meta() == nil || provo.Meta().(*conns.AWSClient) == nil {
				continue
			}

			providerClient := provo.Meta().(*conns.AWSClient)
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

		for _, provo := range *providers {
			if provo == nil || provo.Meta() == nil || provo.Meta().(*conns.AWSClient) == nil {
				continue
			}

			providerClient := provo.Meta().(*conns.AWSClient)
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

func CheckPartition(providers *[]*schema.Provider, expectedPartition string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provo := range *providers {
			if provo == nil || provo.Meta() == nil || provo.Meta().(*conns.AWSClient) == nil {
				continue
			}

			providerPartition := provo.Meta().(*conns.AWSClient).Partition

			if providerPartition != expectedPartition {
				return fmt.Errorf("expected DNS Suffix (%s), got: %s", expectedPartition, providerPartition)
			}
		}

		return nil
	}
}

func CheckReverseDNSPrefix(providers *[]*schema.Provider, expectedReverseDnsPrefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		for _, provo := range *providers {
			if provo == nil || provo.Meta() == nil || provo.Meta().(*conns.AWSClient) == nil {
				continue
			}

			providerReverseDnsPrefix := provo.Meta().(*conns.AWSClient).ReverseDNSPrefix

			if providerReverseDnsPrefix != expectedReverseDnsPrefix {
				return fmt.Errorf("expected DNS Suffix (%s), got: %s", expectedReverseDnsPrefix, providerReverseDnsPrefix)
			}
		}

		return nil
	}
}

// PreCheckEC2ClassicOrHasDefaultVPCWithDefaultSubnets checks that the test region has either
// - The EC2-Classic platform available, or
// - A default VPC with default subnets.
// This check is useful to ensure that an instance can be launched without specifying a subnet.
func PreCheckEC2ClassicOrHasDefaultVPCWithDefaultSubnets(t *testing.T) {
	client := Provider.Meta().(*conns.AWSClient)

	if !conns.HasEC2Classic(client.SupportedPlatforms) && !(HasDefaultVPC(t) && DefaultSubnetCount(t) > 0) {
		t.Skipf("skipping tests; %s does not have EC2-Classic or a default VPC with default subnets", client.Region)
	}
}

// HasDefaultVPC returns whether the current AWS region has a default VPC.
func HasDefaultVPC(t *testing.T) bool {
	conn := Provider.Meta().(*conns.AWSClient).EC2Conn

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
	conn := Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeSubnetsInput{
		Filters: buildAttributeFilterList(map[string]string{
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

func ConfigDefaultTags_Tags0() string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
		`
provider "aws" {
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func ConfigDefaultTags_Tags1(tag1, value1 string) string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
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

func ConfigDefaultTags_Tags2(tag1, value1, tag2, value2 string) string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
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

func ConfigDefaultTagsEmptyConfigurationBlock() string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
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

func ConfigDefaultAndIgnoreTagsEmptyConfigurationBlock() string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
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

func ConfigEndpoints(endpoints string) string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
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

func ConfigIgnoreTagsEmptyConfigurationBlock() string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
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

func ConfigIgnoreTagsKeyPrefixes0() string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
		`
provider "aws" {
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func ConfigIgnoreTagsKeyPrefixes3(tagPrefix1 string) string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
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

func ConfigIgnoreTagsKeyPrefixes2(tagPrefix1, tagPrefix2 string) string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
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

func ConfigIgnoreTagsKeys0() string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
		`
provider "aws" {
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func ConfigIgnoreTagsKeys1(tag1 string) string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
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

func ConfigIgnoreTagsKeys2(tag1, tag2 string) string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
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

func ConfigRegion(region string) string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
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
	conns.SkipIfEnvVarEmpty(t, conns.EnvVarAccAssumeRoleARN, "Amazon Resource Name (ARN) of existing IAM Role to assume for testing restricted permissions")
}

func ConfigAssumeRolePolicy(policy string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  assume_role {
    role_arn = %q
    policy   = %q
  }
}
`, os.Getenv(conns.EnvVarAccAssumeRoleARN), policy)
}

const testAccCheckAWSProviderConfigAssumeRoleEmpty = `
provider "aws" {
  assume_role {
  }
}

data "aws_caller_identity" "current" {}
` //lintignore:AT004

const testAccProviderConfigBase = `
data "aws_partition" "provider_test" {}

# Required to initialize the provider
data "aws_arn" "test" {
  arn = "arn:${data.aws_partition.provider_test.partition}:s3:::test"
}
`

func CheckResourceAttrIsSortedCSV(resourceName, attributeName string) resource.TestCheckFunc {
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

// ConfigCompose can be called to concatenate multiple strings to build test configurations
func ConfigCompose(config ...string) string {
	var str strings.Builder

	for _, conf := range config {
		str.WriteString(conf)
	}

	return str.String()
}

type domainName string

// The top level domain ".test" is reserved by IANA for testing purposes:
// https://datatracker.ietf.org/doc/html/rfc6761
const domainNameTestTopLevelDomain domainName = "test"

// RandomSubdomain creates a random three-level domain name in the form
// "<random>.<random>.test"
// The top level domain ".test" is reserved by IANA for testing purposes:
// https://datatracker.ietf.org/doc/html/rfc6761
func RandomSubdomain() string {
	return string(RandomDomain().RandomSubdomain())
}

// RandomDomainName creates a random two-level domain name in the form
// "<random>.test"
// The top level domain ".test" is reserved by IANA for testing purposes:
// https://datatracker.ietf.org/doc/html/rfc6761
func RandomDomainName() string {
	return string(RandomDomain())
}

// RandomFQDomainName creates a random fully-qualified two-level domain name in the form
// "<random>.test."
// The top level domain ".test" is reserved by IANA for testing purposes:
// https://datatracker.ietf.org/doc/html/rfc6761
func RandomFQDomainName() string {
	return string(RandomDomain().FQDN())
}

func (d domainName) Subdomain(name string) domainName {
	return domainName(fmt.Sprintf("%s.%s", name, d))
}

func (d domainName) RandomSubdomain() domainName {
	return d.Subdomain(sdkacctest.RandString(8))
}

func (d domainName) FQDN() domainName {
	return domainName(fmt.Sprintf("%s.", d))
}

func (d domainName) String() string {
	return string(d)
}

func RandomDomain() domainName {
	return domainNameTestTopLevelDomain.RandomSubdomain()
}

// DefaultEmailAddress is the default email address to set as a
// resource or data source parameter for acceptance tests.
const DefaultEmailAddress = "no-reply@hashicorp.com"

// RandomEmailAddress generates a random email address in the form
// "tf-acc-test-<random>@<domain>"
func RandomEmailAddress(domainName string) string {
	return fmt.Sprintf("%s@%s", sdkacctest.RandomWithPrefix(ResourcePrefix), domainName)
}

func PreCheckOutpostsOutposts(t *testing.T) {
	conn := Provider.Meta().(*conns.AWSClient).OutpostsConn

	input := &outposts.ListOutpostsInput{}

	output, err := conn.ListOutposts(input)

	if PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	// Ensure there is at least one Outpost
	if output == nil || len(output.Outposts) == 0 {
		t.Skip("skipping since no Outposts found")
	}
}

func ACMCertificateDomainFromEnv(t *testing.T) string {
	rootDomain := os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN")

	if rootDomain == "" {
		t.Skip(
			"Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set. " +
				"For DNS validation requests, this domain must be publicly " +
				"accessible and configurable via Route53 during the testing. " +
				"For email validation requests, you must have access to one of " +
				"the five standard email addresses used (admin|administrator|" +
				"hostmaster|postmaster|webmaster)@domain or one of the WHOIS " +
				"contact addresses.")
	}

	if len(rootDomain) >= 56 {
		t.Skip(
			"Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is too long. " +
				"The domain must be shorter than 56 characters to allow for " +
				"subdomain randomization in the testing.")
	}

	return rootDomain
}

// ACM domain names cannot be longer than 64 characters
// Other resources, e.g. Cognito User Pool Domains, limit this to 63
func ACMCertificateRandomSubDomain(rootDomain string) string {
	// Max length (63)
	// Subtract "tf-acc-" prefix (7)
	// Subtract "." between prefix and root domain (1)
	// Subtract length of root domain
	return fmt.Sprintf("tf-acc-%s.%s", sdkacctest.RandString(55-len(rootDomain)), rootDomain)
}

func CheckACMPCACertificateAuthorityActivateCA(certificateAuthority *acmpca.CertificateAuthority) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := Provider.Meta().(*conns.AWSClient).ACMPCAConn

		arn := aws.StringValue(certificateAuthority.Arn)

		getCsrResp, err := conn.GetCertificateAuthorityCsr(&acmpca.GetCertificateAuthorityCsrInput{
			CertificateAuthorityArn: aws.String(arn),
		})
		if err != nil {
			return fmt.Errorf("error getting ACM PCA Certificate Authority (%s) CSR: %s", arn, err)
		}

		issueCertResp, err := conn.IssueCertificate(&acmpca.IssueCertificateInput{
			CertificateAuthorityArn: aws.String(arn),
			Csr:                     []byte(aws.StringValue(getCsrResp.Csr)),
			IdempotencyToken:        aws.String(resource.UniqueId()),
			SigningAlgorithm:        certificateAuthority.CertificateAuthorityConfiguration.SigningAlgorithm,
			TemplateArn:             aws.String(fmt.Sprintf("arn:%s:acm-pca:::template/RootCACertificate/V1", Partition())),
			Validity: &acmpca.Validity{
				Type:  aws.String(acmpca.ValidityPeriodTypeYears),
				Value: aws.Int64(10),
			},
		})
		if err != nil {
			return fmt.Errorf("error issuing ACM PCA Certificate Authority (%s) Root CA certificate from CSR: %s", arn, err)
		}

		// Wait for certificate status to become ISSUED.
		err = conn.WaitUntilCertificateIssued(&acmpca.GetCertificateInput{
			CertificateAuthorityArn: aws.String(arn),
			CertificateArn:          issueCertResp.CertificateArn,
		})
		if err != nil {
			return fmt.Errorf("error waiting for ACM PCA Certificate Authority (%s) Root CA certificate to become ISSUED: %s", arn, err)
		}

		getCertResp, err := conn.GetCertificate(&acmpca.GetCertificateInput{
			CertificateAuthorityArn: aws.String(arn),
			CertificateArn:          issueCertResp.CertificateArn,
		})
		if err != nil {
			return fmt.Errorf("error getting ACM PCA Certificate Authority (%s) issued Root CA certificate: %s", arn, err)
		}

		_, err = conn.ImportCertificateAuthorityCertificate(&acmpca.ImportCertificateAuthorityCertificateInput{
			CertificateAuthorityArn: aws.String(arn),
			Certificate:             []byte(aws.StringValue(getCertResp.Certificate)),
		})
		if err != nil {
			return fmt.Errorf("error importing ACM PCA Certificate Authority (%s) Root CA certificate: %s", arn, err)
		}

		return err
	}
}

func CheckACMPCACertificateAuthorityDisableCA(certificateAuthority *acmpca.CertificateAuthority) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := Provider.Meta().(*conns.AWSClient).ACMPCAConn

		_, err := conn.UpdateCertificateAuthority(&acmpca.UpdateCertificateAuthorityInput{
			CertificateAuthorityArn: certificateAuthority.Arn,
			Status:                  aws.String(acmpca.CertificateAuthorityStatusDisabled),
		})

		return err
	}
}

func CheckACMPCACertificateAuthorityExists(resourceName string, certificateAuthority *acmpca.CertificateAuthority) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := Provider.Meta().(*conns.AWSClient).ACMPCAConn
		input := &acmpca.DescribeCertificateAuthorityInput{
			CertificateAuthorityArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeCertificateAuthority(input)

		if err != nil {
			return err
		}

		if output == nil || output.CertificateAuthority == nil {
			return fmt.Errorf("ACM PCA Certificate Authority %q does not exist", rs.Primary.ID)
		}

		*certificateAuthority = *output.CertificateAuthority

		return nil
	}
}

// PreCheckAPIGatewayTypeEDGE checks if endpoint config type EDGE can be used in a test and skips test if not (i.e., not in standard partition).
func PreCheckAPIGatewayTypeEDGE(t *testing.T) {
	if Partition() != endpoints.AwsPartitionID {
		t.Skipf("skipping test; Endpoint Configuration type EDGE is not supported in this partition (%s)", Partition())
	}
}

func PreCheckDirectoryService(t *testing.T) {
	conn := Provider.Meta().(*conns.AWSClient).DirectoryServiceConn

	input := &directoryservice.DescribeDirectoriesInput{}

	_, err := conn.DescribeDirectories(input)

	if PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// Certain regions such as AWS GovCloud (US) do not support Simple AD directories
// and we do not have a good read-only way to determine this situation. Here we
// opt to perform a creation that will fail so we can determine Simple AD support.
func PreCheckDirectoryServiceSimpleDirectory(t *testing.T) {
	conn := Provider.Meta().(*conns.AWSClient).DirectoryServiceConn

	input := &directoryservice.CreateDirectoryInput{
		Name:     aws.String("corp.example.com"),
		Password: aws.String("PreCheck123"),
		Size:     aws.String(directoryservice.DirectorySizeSmall),
	}

	_, err := conn.CreateDirectory(input)

	if tfawserr.ErrMessageContains(err, directoryservice.ErrCodeClientException, "Simple AD directory creation is currently not supported in this region") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil && !tfawserr.ErrMessageContains(err, directoryservice.ErrCodeInvalidParameterException, "VpcSettings must be specified") {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func ConfigAvailableAZsNoOptIn() string {
	return `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
`
}

func ConfigAvailableAZsNoOptInDefaultExclude() string {
	// Exclude usw2-az4 (us-west-2d) as it has limited instance types.
	return ConfigAvailableAZsNoOptInExclude("usw2-az4", "usgw1-az2")
}

func ConfigAvailableAZsNoOptInExclude(excludeZoneIds ...string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  exclude_zone_ids = ["%[1]s"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
`, strings.Join(excludeZoneIds, "\", \""))
}

// AvailableEC2InstanceTypeForAvailabilityZone returns the configuration for a data source that describes
// the first available EC2 instance type offering in the specified availability zone from a list of preferred instance types.
// The first argument is either an Availability Zone name or Terraform configuration reference to one, e.g.
//   * data.aws_availability_zones.available.names[0]
//   * aws_subnet.test.availability_zone
//   * us-west-2a
// The data source is named 'available'.
func AvailableEC2InstanceTypeForAvailabilityZone(availabilityZoneName string, preferredInstanceTypes ...string) string {
	if !strings.Contains(availabilityZoneName, ".") {
		availabilityZoneName = strconv.Quote(availabilityZoneName)
	}

	return fmt.Sprintf(`
data "aws_ec2_instance_type_offering" "available" {
  filter {
    name   = "instance-type"
    values = ["%[2]s"]
  }

  filter {
    name   = "location"
    values = [%[1]s]
  }

  location_type            = "availability-zone"
  preferred_instance_types = ["%[2]s"]
}
`, availabilityZoneName, strings.Join(preferredInstanceTypes, "\", \""))
}

// AvailableEC2InstanceTypeForRegion returns the configuration for a data source that describes
// the first available EC2 instance type offering in the current region from a list of preferred instance types.
// The data source is named 'available'.
func AvailableEC2InstanceTypeForRegion(preferredInstanceTypes ...string) string {
	return AvailableEC2InstanceTypeForRegionNamed("available", preferredInstanceTypes...)
}

// AvailableEC2InstanceTypeForRegionNamed returns the configuration for a data source that describes
// the first available EC2 instance type offering in the current region from a list of preferred instance types.
// The data source name is configurable.
func AvailableEC2InstanceTypeForRegionNamed(name string, preferredInstanceTypes ...string) string {
	return fmt.Sprintf(`
data "aws_ec2_instance_type_offering" "%[1]s" {
  filter {
    name   = "instance-type"
    values = ["%[2]s"]
  }

  preferred_instance_types = ["%[2]s"]
}
`, name, strings.Join(preferredInstanceTypes, "\", \""))
}

// ConfigLatestAmazonLinuxHVMEBSAMI returns the configuration for a data source that
// describes the latest Amazon Linux AMI using HVM virtualization and an EBS root device.
// The data source is named 'amzn-ami-minimal-hvm-ebs'.
func ConfigLatestAmazonLinuxHVMEBSAMI() string {
	return `
data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
}
`
}

func ConfigLambdaBase(policyName, roleName, sgName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = "%s"
  role = aws_iam_role.iam_for_lambda.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "SNS:Publish"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "xray:PutTraceSegments"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_vpc" "vpc_for_lambda" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lambda-function"
  }
}

resource "aws_subnet" "subnet_for_lambda" {
  vpc_id            = aws_vpc.vpc_for_lambda.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-lambda-function-1"
  }
}

# This is defined here, rather than only in test cases where it's needed is to
# prevent a timeout issue when fully removing Lambda Filesystems
resource "aws_subnet" "subnet_for_lambda_az2" {
  vpc_id            = aws_vpc.vpc_for_lambda.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-lambda-function-2"
  }
}

resource "aws_security_group" "sg_for_lambda" {
  name        = "%s"
  description = "Allow all inbound traffic for lambda test"
  vpc_id      = aws_vpc.vpc_for_lambda.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, policyName, roleName, sgName)
}

func CheckVPCExists(n string, vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC ID is set")
		}

		conn := Provider.Meta().(*conns.AWSClient).EC2Conn
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
		if err != nil {
			return err
		}
		if len(resp.Vpcs) == 0 || resp.Vpcs[0] == nil {
			return fmt.Errorf("VPC not found")
		}

		*vpc = *resp.Vpcs[0]

		return nil
	}
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

		expected := Provider.Meta().(*conns.AWSClient).AccountID
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

func convertIPToDashIP(ip string) string {
	return strings.Replace(ip, ".", "-", -1)
}

func regionalPrivateDNSSuffix(region string) string {
	if region == endpoints.UsEast1RegionID {
		return "ec2.internal"
	}

	return fmt.Sprintf("%s.compute.internal", region)
}

// buildAttributeFilterList takes a flat map of scalar attributes (most
// likely values extracted from a *schema.ResourceData on an EC2-querying
// data source) and produces a []*ec2.Filter representing an exact match
// for each of the given non-empty attributes.
//
// The keys of the given attributes map are the attribute names expected
// by the EC2 API, which are usually either in camelcase or with dash-separated
// words. We conventionally map these to underscore-separated identifiers
// with the same words when presenting these as data source query attributes
// in Terraform.
//
// It's the callers responsibility to transform any non-string values into
// the appropriate string serialization required by the AWS API when
// encoding the given filter. Any attributes given with empty string values
// are ignored, assuming that the user wishes to leave that attribute
// unconstrained while filtering.
//
// The purpose of this function is to create values to pass in
// for the "Filters" attribute on most of the "Describe..." API functions in
// the EC2 API, to aid in the implementation of Terraform data sources that
// retrieve data about EC2 objects.
func buildAttributeFilterList(attrs map[string]string) []*ec2.Filter {
	var filters []*ec2.Filter

	// sort the filters by name to make the output deterministic
	var names []string
	for filterName := range attrs {
		names = append(names, filterName)
	}

	sort.Strings(names)

	for _, filterName := range names {
		value := attrs[filterName]
		if value == "" {
			continue
		}

		filters = append(filters, &ec2.Filter{
			Name:   aws.String(filterName),
			Values: []*string{aws.String(value)},
		})
	}

	return filters
}
