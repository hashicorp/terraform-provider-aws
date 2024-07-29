// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	accounttypes "github.com/aws/aws-sdk-go-v2/service/account/types"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	acmpcatypes "github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	dstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	inspector2types "github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	organizationstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	ssoadmintypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	terraformsdk "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tfaccount "github.com/hashicorp/terraform-provider-aws/internal/service/account"
	tfacmpca "github.com/hashicorp/terraform-provider-aws/internal/service/acmpca"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	tfsts "github.com/hashicorp/terraform-provider-aws/internal/service/sts"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/jmespath/go-jmespath"
	"github.com/mitchellh/mapstructure"
)

const (
	// Provider name for single configuration testing
	ProviderName = "aws"

	// Provider name for alternate configuration testing
	ProviderNameAlternate = "awsalternate"

	// Provider name for alternate account and alternate region configuration testing
	ProviderNameAlternateAccountAlternateRegion = "awsalternateaccountalternateregion"

	// Provider name for alternate account and same region configuration testing
	ProviderNameAlternateAccountSameRegion = "awsalternateaccountsameregion"

	// Provider name for same account and alternate region configuration testing
	ProviderNameSameAccountAlternateRegion = "awssameaccountalternateregion"

	// Provider name for third configuration testing
	ProviderNameThird = "awsthird"

	ResourcePrefix = "tf-acc-test"

	CertificateIssueTimeout = 5 * time.Minute
)

const RFC3339RegexPattern = `^[0-9]{4}-(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[01])[Tt]([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](\.[0-9]+)?([Zz]|([+-]([01][0-9]|2[0-3]):[0-5][0-9]))$`
const regionRegexp = `[a-z]{2}(-[a-z]+)+-\d`
const accountIDRegexp = `(aws|aws-managed|\d{12})`

// Skip implements a wrapper for (*testing.T).Skip() to prevent unused linting reports
//
// Reference: https://github.com/dominikh/go-tools/issues/633#issuecomment-606560616
func Skip(t *testing.T, message string) {
	t.Helper()

	t.Skip(message)
}

// ProtoV5ProviderFactories is a static map containing only the main provider instance
//
// Use other ProviderFactories functions, such as FactoriesAlternate,
// for tests requiring special provider configurations.
var (
	ProtoV5ProviderFactories map[string]func() (tfprotov5.ProviderServer, error) = protoV5ProviderFactoriesInit(context.Background(), ProviderName)
)

// Provider is the "main" provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// PreCheck(t) must be called before using this provider instance.
var (
	Provider *schema.Provider = errs.Must(provider.New(context.Background()))
)

type ProviderFunc func() *schema.Provider

// testAccProviderConfigure ensures Provider is only configured once
//
// The PreCheck(t) function is invoked for every test and this prevents
// extraneous reconfiguration to the same values each time. However, this does
// not prevent reconfiguration that may happen should the address of
// Provider be errantly reused in ProviderFactories.
var testAccProviderConfigure sync.Once

func protoV5ProviderFactoriesInit(ctx context.Context, providerNames ...string) map[string]func() (tfprotov5.ProviderServer, error) {
	factories := make(map[string]func() (tfprotov5.ProviderServer, error), len(providerNames))

	for _, name := range providerNames {
		factories[name] = func() (tfprotov5.ProviderServer, error) {
			providerServerFactory, _, err := provider.ProtoV5ProviderServerFactory(ctx)

			if err != nil {
				return nil, err
			}

			return providerServerFactory(), nil
		}
	}

	return factories
}

func protoV5ProviderFactoriesNamedInit(ctx context.Context, t *testing.T, providers map[string]*schema.Provider, providerNames ...string) map[string]func() (tfprotov5.ProviderServer, error) {
	t.Helper()

	factories := make(map[string]func() (tfprotov5.ProviderServer, error), len(providerNames))

	for _, name := range providerNames {
		providerServerFactory, p, err := provider.ProtoV5ProviderServerFactory(ctx)

		if err != nil {
			t.Fatal(err)
		}

		factories[name] = func() (tfprotov5.ProviderServer, error) { //nolint:unparam
			return providerServerFactory(), nil
		}

		providers[name] = p
	}

	return factories
}

func protoV5ProviderFactoriesPlusProvidersInit(ctx context.Context, t *testing.T, providers *[]*schema.Provider, providerNames ...string) map[string]func() (tfprotov5.ProviderServer, error) {
	t.Helper()

	factories := make(map[string]func() (tfprotov5.ProviderServer, error), len(providerNames))

	for _, name := range providerNames {
		providerServerFactory, p, err := provider.ProtoV5ProviderServerFactory(ctx)

		if err != nil {
			t.Fatal(err)
		}

		factories[name] = func() (tfprotov5.ProviderServer, error) { //nolint:unparam
			return providerServerFactory(), nil
		}

		if providers != nil {
			*providers = append(*providers, p)
		}
	}

	return factories
}

// ProtoV5FactoriesPlusProvidersAlternate creates ProtoV5ProviderFactories for cross-account and cross-region configurations
// and also returns Providers suitable for use with AWS APIs.
//
// For cross-region testing: Typically paired with PreCheckMultipleRegion and ConfigAlternateRegionProvider.
//
// For cross-account testing: Typically paired with PreCheckAlternateAccount and ConfigAlternateAccountProvider.
func ProtoV5FactoriesPlusProvidersAlternate(ctx context.Context, t *testing.T, providers *[]*schema.Provider) map[string]func() (tfprotov5.ProviderServer, error) {
	t.Helper()

	return protoV5ProviderFactoriesPlusProvidersInit(ctx, t, providers, ProviderName, ProviderNameAlternate)
}

func ProtoV5FactoriesNamedAlternate(ctx context.Context, t *testing.T, providers map[string]*schema.Provider) map[string]func() (tfprotov5.ProviderServer, error) {
	t.Helper()

	return ProtoV5FactoriesNamed(ctx, t, providers, ProviderName, ProviderNameAlternate)
}

func ProtoV5FactoriesNamed(ctx context.Context, t *testing.T, providers map[string]*schema.Provider, providerNames ...string) map[string]func() (tfprotov5.ProviderServer, error) {
	t.Helper()

	return protoV5ProviderFactoriesNamedInit(ctx, t, providers, providerNames...)
}

func ProtoV5FactoriesAlternate(ctx context.Context, t *testing.T) map[string]func() (tfprotov5.ProviderServer, error) {
	t.Helper()

	return protoV5ProviderFactoriesInit(ctx, ProviderName, ProviderNameAlternate)
}

// ProtoV5FactoriesAlternateAccountAndAlternateRegion creates ProtoV5ProviderFactories for cross-account and cross-region configurations
//
// Usage typically paired with PreCheckMultipleRegion, PreCheckAlternateAccount,
// and ConfigAlternateAccountAndAlternateRegionProvider.
func ProtoV5FactoriesAlternateAccountAndAlternateRegion(ctx context.Context, t *testing.T) map[string]func() (tfprotov5.ProviderServer, error) {
	t.Helper()

	return protoV5ProviderFactoriesInit(
		ctx,
		ProviderName,
		ProviderNameAlternateAccountAlternateRegion,
		ProviderNameAlternateAccountSameRegion,
		ProviderNameSameAccountAlternateRegion,
	)
}

// ProtoV5FactoriesMultipleRegions creates ProtoV5ProviderFactories for the specified number of region configurations
//
// Usage typically paired with PreCheckMultipleRegion and ConfigMultipleRegionProvider.
func ProtoV5FactoriesMultipleRegions(ctx context.Context, t *testing.T, n int) map[string]func() (tfprotov5.ProviderServer, error) {
	t.Helper()

	switch n {
	case 2:
		return protoV5ProviderFactoriesInit(ctx, ProviderName, ProviderNameAlternate)
	case 3:
		return protoV5ProviderFactoriesInit(ctx, ProviderName, ProviderNameAlternate, ProviderNameThird)
	default:
		t.Fatalf("invalid number of Region configurations: %d", n)
	}

	return nil
}

// PreCheck verifies and sets required provider testing configuration
//
// This PreCheck function should be present in every acceptance test. It allows
// test configurations to omit a provider configuration with region and ensures
// testing functions that attempt to call AWS APIs are previously configured.
//
// These verifications and configuration are preferred at this level to prevent
// provider developers from experiencing less clear errors for every test.
func PreCheck(ctx context.Context, t *testing.T) {
	t.Helper()

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderConfigure.Do(func() {
		envvar.FailIfAllEmpty(t, []string{envvar.Profile, envvar.AccessKeyId, envvar.ContainerCredentialsFullURI}, "credentials for running acceptance testing")

		if os.Getenv(envvar.AccessKeyId) != "" {
			envvar.FailIfEmpty(t, envvar.SecretAccessKey, "static credentials value when using "+envvar.AccessKeyId)
		}

		// Setting the AWS_DEFAULT_REGION environment variable here allows all tests to omit
		// a provider configuration with a region. This defaults to us-west-2 for provider
		// developer simplicity and has been in the codebase for a very long time.
		//
		// This handling must be preserved until either:
		//   * AWS_DEFAULT_REGION is required and checked above (should mention us-west-2 default)
		//   * Region is automatically handled via shared AWS configuration file and still verified
		region := Region()
		os.Setenv(envvar.DefaultRegion, region)

		diags := Provider.Configure(ctx, terraformsdk.NewResourceConfigRaw(nil))
		if err := sdkdiag.DiagnosticsError(diags); err != nil {
			t.Fatalf("configuring provider: %s", err)
		}
	})
}

// ProviderAccountID returns the account ID of an AWS provider
func ProviderAccountID(provider *schema.Provider) string {
	if provider == nil {
		log.Print("[DEBUG] Unable to read account ID from test provider: empty provider")
		return ""
	}
	if provider.Meta() == nil {
		log.Print("[DEBUG] Unable to read account ID from test provider: unconfigured provider")
		return ""
	}
	client, ok := provider.Meta().(*conns.AWSClient)
	if !ok {
		log.Print("[DEBUG] Unable to read account ID from test provider: non-AWS or unconfigured AWS provider")
		return ""
	}
	return client.AccountID
}

// CheckDestroyNoop is a TestCheckFunc to be used as a TestCase's CheckDestroy when no such check can be made.
func CheckDestroyNoop(*terraform.State) error {
	return nil
}

// CheckSleep returns a TestCheckFunc that pauses the current goroutine for at least the duration d.
func CheckSleep(t *testing.T, d time.Duration) resource.TestCheckFunc {
	t.Helper()

	return func(*terraform.State) error {
		time.Sleep(d)

		return nil
	}
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

// CheckResourceAttrRegionalHostnameService ensures the Terraform state exactly matches a service DNS hostname with region and partition DNS suffix
//
// For example: ec2.us-west-2.amazonaws.com
func CheckResourceAttrRegionalHostnameService(resourceName, attributeName, serviceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hostname := fmt.Sprintf("%s.%s.%s", serviceName, Region(), PartitionDNSSuffix())

		return resource.TestCheckResourceAttr(resourceName, attributeName, hostname)(s)
	}
}

// CheckResourceAttrNameFromPrefix verifies that the state attribute value matches name generated from given prefix
func CheckResourceAttrNameFromPrefix(resourceName string, attributeName string, prefix string) resource.TestCheckFunc {
	return CheckResourceAttrNameWithSuffixFromPrefix(resourceName, attributeName, prefix, "")
}

// Regexp for "<start-of-string>terraform-<26 lowercase hex digits><additional suffix><end-of-string>".
func resourceUniqueIDPrefixPlusAdditionalSuffixRegexp(prefix, suffix string) *regexp.Regexp {
	return regexache.MustCompile(fmt.Sprintf("^%s[[:xdigit:]]{%d}%s$", prefix, id.UniqueIDSuffixLength, suffix))
}

// CheckResourceAttrNameWithSuffixFromPrefix verifies that the state attribute value matches name with suffix generated from given prefix
func CheckResourceAttrNameWithSuffixFromPrefix(resourceName string, attributeName string, prefix string, suffix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeMatch := resourceUniqueIDPrefixPlusAdditionalSuffixRegexp(prefix, suffix)
		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// CheckResourceAttrNameGenerated verifies that the state attribute value matches name automatically generated without prefix
func CheckResourceAttrNameGenerated(resourceName string, attributeName string) resource.TestCheckFunc {
	return CheckResourceAttrNameWithSuffixGenerated(resourceName, attributeName, "")
}

// CheckResourceAttrNameGeneratedWithPrefix verifies that the state attribute value matches name automatically generated with prefix
func CheckResourceAttrNameGeneratedWithPrefix(resourceName string, attributeName string, prefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestMatchResourceAttr(resourceName, attributeName, resourceUniqueIDPrefixPlusAdditionalSuffixRegexp(prefix, ""))(s)
	}
}

// CheckResourceAttrNameWithSuffixGenerated verifies that the state attribute value matches name with suffix automatically generated without prefix
func CheckResourceAttrNameWithSuffixGenerated(resourceName string, attributeName string, suffix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestMatchResourceAttr(resourceName, attributeName, resourceUniqueIDPrefixPlusAdditionalSuffixRegexp(id.UniqueIdPrefix, suffix))(s)
	}
}

// MatchResourceAttrAccountID ensures the Terraform state regexp matches an account ID
func MatchResourceAttrAccountID(resourceName, attributeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestMatchResourceAttr(resourceName, attributeName, regexache.MustCompile(`^\d{12}$`))(s)
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
			return fmt.Errorf("unable to compile ARN regexp (%s): %w", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// MatchResourceAttrRegionalARNRegion ensures the Terraform state regexp matches a formatted ARN with the specified region
func MatchResourceAttrRegionalARNRegion(resourceName, attributeName, arnService, region string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: AccountID(),
			Partition: Partition(),
			Region:    region,
			Resource:  arnResourceRegexp.String(),
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("unable to compile ARN regexp (%s): %w", arnRegexp, err)
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
			return fmt.Errorf("unable to compile ARN regexp (%s): %s", arnRegexp, err)
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
			return fmt.Errorf("unable to compile ARN regexp (%s): %w", arnRegexp, err)
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
			return fmt.Errorf("unable to compile hostname regexp (%s): %w", hostnameRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, hostnameRegexp)(s)
	}
}

// MatchResourceAttrGlobalHostname ensures the Terraform state regexp matches a formatted DNS hostname with partition DNS suffix and without region
func MatchResourceAttrGlobalHostname(resourceName, attributeName, serviceName string, hostnamePrefixRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		hostnameRegexpPattern := fmt.Sprintf("%s\\.%s\\.%s$", hostnamePrefixRegexp.String(), serviceName, PartitionDNSSuffix())

		hostnameRegexp, err := regexp.Compile(hostnameRegexpPattern)

		if err != nil {
			return fmt.Errorf("unable to compile hostname regexp (%s): %w", hostnameRegexp, err)
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
			return fmt.Errorf("unable to compile ARN regexp (%s): %w", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// CheckResourceAttrRegionalARNIgnoreRegionAndAccount ensures the Terraform state exactly matches a formatted ARN with region without specifying the region or account
func CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: accountIDRegexp,
			Partition: Partition(),
			Region:    regionRegexp,
			Resource:  arnResource,
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("unable to compile ARN regexp (%s): %w", arnRegexp, err)
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
			return fmt.Errorf("unable to compile ARN regexp (%s): %s", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// CheckResourceAttrRFC3339 ensures the Terraform state matches a RFC3339 value
// This TestCheckFunc will likely be moved to the Terraform Plugin SDK in the future.
func CheckResourceAttrRFC3339(resourceName, attributeName string) resource.TestCheckFunc {
	return resource.TestMatchResourceAttr(resourceName, attributeName, regexache.MustCompile(RFC3339RegexPattern))
}

// CheckResourceAttrEquivalentJSON is a TestCheckFunc that compares a JSON value with an expected value.
// Both JSON values are normalized before being compared.
func CheckResourceAttrEquivalentJSON(n, key, expectedJSON string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(n, key, func(value string) error {
		vNormal, err := structure.NormalizeJsonString(value)
		if err != nil {
			return fmt.Errorf("%s: Error normalizing JSON in %q: %w", n, key, err)
		}

		expectedNormal, err := structure.NormalizeJsonString(expectedJSON)
		if err != nil {
			return fmt.Errorf("normalizing expected JSON: %w", err)
		}

		if vNormal != expectedNormal {
			return fmt.Errorf("%s: Attribute %q expected\n%s\ngot\n%s", n, key, expectedJSON, value)
		}
		return nil
	})
}

func CheckResourceAttrJMES(name, key, jmesPath, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		is, err := PrimaryInstanceState(s, name)
		if err != nil {
			return err
		}

		attr, ok := is.Attributes[key]
		if !ok {
			return fmt.Errorf("%s: Attribute %q not set", name, key)
		}

		var jsonData any
		err = json.Unmarshal([]byte(attr), &jsonData)
		if err != nil {
			return fmt.Errorf("%s: Expected attribute %q to be JSON: %w", name, key, err)
		}

		result, err := jmespath.Search(jmesPath, jsonData)
		if err != nil {
			return fmt.Errorf("invalid JMESPath %q: %w", jmesPath, err)
		}

		var v string
		switch x := result.(type) {
		case string:
			v = x
		case float64:
			v = strconv.FormatFloat(x, 'f', -1, 64)
		case bool:
			v = fmt.Sprint(x)
		default:
			return fmt.Errorf(`%[1]s: Attribute %[2]q, JMESPath %[3]q got "%#[4]v" (%[4]T)`, name, key, jmesPath, result)
		}

		if v != value {
			return fmt.Errorf("%s: Attribute %q, JMESPath %q expected %#v, got %#v", name, key, jmesPath, value, v)
		}

		return nil
	}
}

func CheckResourceAttrJMESPair(nameFirst, keyFirst, jmesPath, nameSecond, keySecond string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		first, err := PrimaryInstanceState(s, nameFirst)
		if err != nil {
			return err
		}

		second, err := PrimaryInstanceState(s, nameSecond)
		if err != nil {
			return err
		}

		vFirst, okFirst := first.Attributes[keyFirst]
		if !okFirst {
			return fmt.Errorf("%s: Attribute %q not set", nameFirst, keyFirst)
		}

		var jsonData any
		err = json.Unmarshal([]byte(vFirst), &jsonData)
		if err != nil {
			return fmt.Errorf("%s: Expected attribute %q to be JSON: %w", nameFirst, keyFirst, err)
		}

		result, err := jmespath.Search(jmesPath, jsonData)
		if err != nil {
			return fmt.Errorf("invalid JMESPath %q: %w", jmesPath, err)
		}

		var value string
		switch x := result.(type) {
		case string:
			value = x
		case float64:
			value = strconv.FormatFloat(x, 'f', -1, 64)
		case bool:
			value = fmt.Sprint(x)
		default:
			return fmt.Errorf(`%[1]s: Attribute %[2]q, JMESPath %[3]q got "%#[4]v" (%[4]T)`, nameFirst, keyFirst, jmesPath, result)
		}

		vSecond, okSecond := second.Attributes[keySecond]
		if !okSecond {
			return fmt.Errorf("%s: Attribute %q, JMESPath %q is %q, but %q is not set in %s", nameFirst, keyFirst, jmesPath, value, keySecond, nameSecond)
		}

		if value != vSecond {
			return fmt.Errorf("%s: Attribute %q, JMESPath %q, expected %q, got %q", nameFirst, keyFirst, jmesPath, vSecond, value)
		}

		return nil
	}
}

// CheckResourceAttrContains ensures the Terraform state value contains the specified substr.
func CheckResourceAttrContains(name, key, substr string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(name, key, func(value string) error {
		if strings.Contains(value, substr) {
			return nil
		}
		return fmt.Errorf("%s: Attribute '%s' expected contains %#v, got %#v", name, key, substr, value)
	})
}

// CheckResourceAttrHasPrefix ensures the Terraform state value has the specified prefix.
func CheckResourceAttrHasPrefix(name, key, prefix string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(name, key, func(value string) error {
		if strings.HasPrefix(value, prefix) {
			return nil
		}
		return fmt.Errorf("%s: Attribute '%s' expected prefix %#v, got %#v", name, key, prefix, value)
	})
}

// CheckResourceAttrHasSuffix ensures the Terraform state value has the specified suffix.
func CheckResourceAttrHasSuffix(name, key, suffix string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(name, key, func(value string) error {
		if strings.HasSuffix(value, suffix) {
			return nil
		}
		return fmt.Errorf("%s: Attribute '%s' expected suffix %#v, got %#v", name, key, suffix, value)
	})
}

// Copied and inlined from the SDK testing code
func PrimaryInstanceState(s *terraform.State, name string) (*terraform.InstanceState, error) {
	rs, ok := s.RootModule().Resources[name]
	if !ok {
		return nil, fmt.Errorf("not found: %s", name)
	}

	is := rs.Primary
	if is == nil {
		return nil, fmt.Errorf("no primary instance: %s", name)
	}

	return is, nil
}

// AccountID returns the account ID of Provider
// Must be used within a resource.TestCheckFunc
func AccountID() string {
	return ProviderAccountID(Provider)
}

func Region() string {
	return envvar.GetWithDefault(envvar.DefaultRegion, names.USWest2RegionID)
}

func AlternateRegion() string {
	return envvar.GetWithDefault(envvar.AlternateRegion, names.USEast1RegionID)
}

func ThirdRegion() string {
	return envvar.GetWithDefault(envvar.ThirdRegion, names.USEast2RegionID)
}

func Partition() string {
	return names.PartitionForRegion(Region())
}

func PartitionRegions() []string {
	return RegionsInPartition(Partition())
}

func PartitionDNSSuffix() string {
	return names.DNSSuffixForPartition(Partition())
}

func PartitionReverseDNSPrefix() string {
	return names.ReverseDNS(PartitionDNSSuffix())
}

func alternateRegionPartition() string {
	return names.PartitionForRegion(AlternateRegion())
}

func thirdRegionPartition() string {
	return names.PartitionForRegion(ThirdRegion())
}

func PreCheckAlternateAccount(t *testing.T) {
	t.Helper()

	envvar.SkipIfAllEmpty(t, []string{envvar.AlternateProfile, envvar.AlternateAccessKeyId}, "credentials for running acceptance testing in alternate AWS account")

	if os.Getenv(envvar.AlternateAccessKeyId) != "" {
		envvar.SkipIfEmpty(t, envvar.AlternateSecretAccessKey, "static credentials value when using "+envvar.AlternateAccessKeyId)
	}
}

func PreCheckThirdAccount(t *testing.T) {
	t.Helper()

	envvar.SkipIfAllEmpty(t, []string{envvar.ThirdProfile, envvar.ThirdAccessKeyId}, "credentials for running acceptance testing in third AWS account")

	if os.Getenv(envvar.ThirdAccessKeyId) != "" {
		envvar.SkipIfEmpty(t, envvar.ThirdSecretAccessKey, "static credentials value when using "+envvar.ThirdAccessKeyId)
	}
}

func PreCheckPartitionHasService(t *testing.T, serviceID string) {
	t.Helper()

	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), Region()); ok {
		if _, ok := partition.Services()[serviceID]; !ok {
			t.Skipf("skipping tests; partition %s does not support %s service", partition.ID(), serviceID)
		}
	}
}

func PreCheckMultipleRegion(t *testing.T, regions int) {
	t.Helper()

	if len(PartitionRegions()) <= 1 {
		t.Skipf("Skipping multiple region test as 1 or fewer regions detected in partion (%s)", Partition())
	}

	if Region() == AlternateRegion() {
		t.Fatalf("%s and %s must be set to different values for acceptance tests", envvar.DefaultRegion, envvar.AlternateRegion)
	}

	if Partition() != alternateRegionPartition() {
		t.Fatalf("%s partition (%s) does not match %s partition (%s)", envvar.AlternateRegion, alternateRegionPartition(), envvar.DefaultRegion, Partition())
	}

	if regions >= 3 {
		if thirdRegionPartition() == names.USGovCloudPartitionID || Partition() == names.USGovCloudPartitionID {
			t.Skipf("wanted %d regions, partition (%s) only has 2 regions", regions, Partition())
		}

		if Region() == ThirdRegion() {
			t.Fatalf("%s and %s must be set to different values for acceptance tests", envvar.DefaultRegion, envvar.ThirdRegion)
		}

		if AlternateRegion() == ThirdRegion() {
			t.Fatalf("%s and %s must be set to different values for acceptance tests", envvar.AlternateRegion, envvar.ThirdRegion)
		}

		if Partition() != thirdRegionPartition() {
			t.Fatalf("%s partition (%s) does not match %s partition (%s)", envvar.ThirdRegion, thirdRegionPartition(), envvar.DefaultRegion, Partition())
		}
	}

	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), Region()); ok {
		if len(partition.Regions()) < regions {
			t.Skipf("skipping tests; partition includes %d regions, %d expected", len(partition.Regions()), regions)
		}
	}
}

// PreCheckRegion checks that the test region is one of the specified AWS Regions.
func PreCheckRegion(t *testing.T, regions ...string) {
	t.Helper()

	curr := Region()
	var regionOK bool

	for _, region := range regions {
		if curr == region {
			regionOK = true
			break
		}
	}

	if !regionOK {
		t.Skipf("skipping tests; %s (%s) not supported. Supported: [%s]", envvar.DefaultRegion, curr, strings.Join(regions, ", "))
	}
}

// PreCheckRegionNot checks that the test region is not one of the specified AWS Regions.
func PreCheckRegionNot(t *testing.T, regions ...string) {
	t.Helper()

	curr := Region()

	for _, region := range regions {
		if curr == region {
			t.Skipf("skipping tests; %s (%s) not supported", envvar.DefaultRegion, curr)
		}
	}
}

// PreCheckAlternateRegionIs checks that the alternate test region is the specified AWS Region.
func PreCheckAlternateRegionIs(t *testing.T, region string) {
	t.Helper()

	if curr := AlternateRegion(); curr != region {
		t.Skipf("skipping tests; %s (%s) does not equal %s", envvar.AlternateRegion, curr, region)
	}
}

// PreCheckPartition checks that the test partition is the specified partition.
func PreCheckPartition(t *testing.T, partition string) {
	t.Helper()

	if curr := Partition(); curr != partition {
		t.Skipf("skipping tests; current partition (%s) does not equal %s", curr, partition)
	}
}

// PreCheckPartitionNot checks that the test partition is not one of the specified partitions.
func PreCheckPartitionNot(t *testing.T, partitions ...string) {
	t.Helper()

	for _, partition := range partitions {
		if curr := Partition(); curr == partition {
			t.Skipf("skipping tests; current partition (%s) not supported", curr)
		}
	}
}

func PreCheckCognitoIdentityProvider(ctx context.Context, t *testing.T) {
	t.Helper()

	conn := Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

	input := &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int32(1),
	}

	_, err := conn.ListUserPools(ctx, input)

	if PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func PreCheckInspector2(ctx context.Context, t *testing.T) {
	t.Helper()

	conn := Provider.Meta().(*conns.AWSClient).Inspector2Client(ctx)

	_, err := conn.ListDelegatedAdminAccounts(ctx, &inspector2.ListDelegatedAdminAccountsInput{})

	if errs.IsA[*inspector2types.AccessDeniedException](err) {
		t.Skipf("Amazon Inspector not available: %s", err)
	}

	if err != nil {
		t.Fatalf("listing Inspector2 delegated administrators: %s", err)
	}
}

func PreCheckOrganizationsAccount(ctx context.Context, t *testing.T) {
	t.Helper()

	_, err := tforganizations.FindOrganization(ctx, Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx))

	if tfresource.NotFound(err) {
		return
	}

	if err != nil {
		t.Fatalf("describing AWS Organization: %s", err)
	}

	t.Skip("skipping tests; this AWS account must not be an existing member of an AWS Organization")
}

func PreCheckOrganizationsEnabled(ctx context.Context, t *testing.T) *organizationstypes.Organization {
	t.Helper()

	return PreCheckOrganizationsEnabledWithProvider(ctx, t, func() *schema.Provider { return Provider })
}

func PreCheckOrganizationsEnabledWithProvider(ctx context.Context, t *testing.T, providerF ProviderFunc) *organizationstypes.Organization {
	t.Helper()

	organization, err := tforganizations.FindOrganization(ctx, providerF().Meta().(*conns.AWSClient).OrganizationsClient(ctx))

	if tfresource.NotFound(err) {
		t.Skip("this AWS account must be an existing member of an AWS Organization")
	}

	if err != nil {
		t.Fatalf("describing AWS Organization: %s", err)
	}

	return organization
}

func PreCheckOrganizationManagementAccount(ctx context.Context, t *testing.T) {
	t.Helper()

	PreCheckOrganizationManagementAccountWithProvider(ctx, t, func() *schema.Provider { return Provider })
}

func PreCheckOrganizationManagementAccountWithProvider(ctx context.Context, t *testing.T, providerF ProviderFunc) {
	t.Helper()

	organization := PreCheckOrganizationsEnabledWithProvider(ctx, t, providerF)

	callerIdentity, err := tfsts.FindCallerIdentity(ctx, providerF().Meta().(*conns.AWSClient).STSClient(ctx))

	if err != nil {
		t.Fatalf("getting current identity: %s", err)
	}

	if aws.StringValue(organization.MasterAccountId) != aws.StringValue(callerIdentity.Account) {
		t.Skip("this AWS account must be the management account of an AWS Organization")
	}
}

func PreCheckOrganizationMemberAccount(ctx context.Context, t *testing.T) {
	t.Helper()

	PreCheckOrganizationMemberAccountWithProvider(ctx, t, func() *schema.Provider { return Provider })
}

func PreCheckOrganizationMemberAccountWithProvider(ctx context.Context, t *testing.T, providerF ProviderFunc) {
	t.Helper()

	organization := PreCheckOrganizationsEnabledWithProvider(ctx, t, providerF)

	callerIdentity, err := tfsts.FindCallerIdentity(ctx, providerF().Meta().(*conns.AWSClient).STSClient(ctx))

	if err != nil {
		t.Fatalf("getting current identity: %s", err)
	}

	if aws.StringValue(organization.MasterAccountId) == aws.StringValue(callerIdentity.Account) {
		t.Skip("this AWS account must not be the management account of an AWS Organization")
	}
}

func PreCheckPinpointApp(ctx context.Context, t *testing.T) {
	conn := Provider.Meta().(*conns.AWSClient).PinpointConn(ctx)

	input := &pinpoint.GetAppsInput{}

	_, err := conn.GetAppsWithContext(ctx, input)

	if PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func PreCheckRegionOptIn(ctx context.Context, t *testing.T, region string) {
	t.Helper()

	output, err := tfaccount.FindRegionOptStatus(ctx, Provider.Meta().(*conns.AWSClient).AccountClient(ctx), "", region)

	if err != nil {
		t.Fatalf("reading Region (%s) opt-in status: %s", region, err)
	}

	if status := output.RegionOptStatus; status != accounttypes.RegionOptStatusEnabled && status != accounttypes.RegionOptStatusEnabledByDefault {
		t.Skipf("Region (%s) opt-in status: %s", region, status)
	}
}

func PreCheckSSOAdminInstances(ctx context.Context, t *testing.T) {
	t.Helper()

	conn := Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)
	input := &ssoadmin.ListInstancesInput{}
	var instances []ssoadmintypes.InstanceMetadata

	paginator := ssoadmin.NewListInstancesPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if PreCheckSkipError(err) {
			t.Skipf("skipping tests: %s", err)
		}
		if err != nil {
			t.Fatalf("listing SSO Instances: %s", err)
		}

		if page != nil {
			instances = append(instances, page.Instances...)
		}
	}

	if len(instances) == 0 {
		t.Skip("skipping tests; no SSO Instances found.")
	}
}

func PreCheckHasIAMRole(ctx context.Context, t *testing.T, roleName string) {
	t.Helper()

	_, err := tfiam.FindRoleByName(ctx, Provider.Meta().(*conns.AWSClient).IAMClient(ctx), roleName)

	if tfresource.NotFound(err) {
		t.Skipf("skipping acceptance test: required IAM role %q not found", roleName)
	}

	if PreCheckSkipError(err) {
		t.Skipf("skipping acceptance test: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func PreCheckIAMServiceLinkedRole(ctx context.Context, t *testing.T, pathPrefix string) {
	t.Helper()

	PreCheckIAMServiceLinkedRoleWithProvider(ctx, t, func() *schema.Provider { return Provider }, pathPrefix)
}

func PreCheckIAMServiceLinkedRoleWithProvider(ctx context.Context, t *testing.T, providerF ProviderFunc, pathPrefix string) {
	t.Helper()

	conn := providerF().Meta().(*conns.AWSClient).IAMClient(ctx)
	input := &iam.ListRolesInput{
		PathPrefix: aws.String(pathPrefix),
	}
	var role iamtypes.Role

	pages := iam.NewListRolesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if PreCheckSkipError(err) {
			t.Skipf("skipping tests: %s", err)
		}
		if err != nil {
			t.Fatalf("listing IAM roles: %s", err)
		}

		for _, r := range page.Roles {
			role = r
			break
		}
	}

	if reflect.ValueOf(role).IsZero() {
		t.Skipf("skipping tests; missing IAM service-linked role %s. Please create the role and retry", pathPrefix)
	}
}

func PreCheckDirectoryService(ctx context.Context, t *testing.T) {
	t.Helper()

	conn := Provider.Meta().(*conns.AWSClient).DSClient(ctx)
	input := &directoryservice.DescribeDirectoriesInput{}

	_, err := conn.DescribeDirectories(ctx, input)

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
func PreCheckDirectoryServiceSimpleDirectory(ctx context.Context, t *testing.T) {
	t.Helper()

	conn := Provider.Meta().(*conns.AWSClient).DSClient(ctx)
	input := &directoryservice.CreateDirectoryInput{
		Name:     aws.String("corp.example.com"),
		Password: aws.String("PreCheck123"),
		Size:     dstypes.DirectorySizeSmall,
	}

	_, err := conn.CreateDirectory(ctx, input)

	if errs.IsAErrorMessageContains[*dstypes.ClientException](err, "Simple AD directory creation is currently not supported in this region") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil && !errs.IsAErrorMessageContains[*dstypes.InvalidParameterException](err, "VpcSettings must be specified") {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func PreCheckOutpostsOutposts(ctx context.Context, t *testing.T) {
	t.Helper()

	conn := Provider.Meta().(*conns.AWSClient).OutpostsConn(ctx)
	input := &outposts.ListOutpostsInput{}

	output, err := conn.ListOutpostsWithContext(ctx, input)

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

func PreCheckWAFV2CloudFrontScope(ctx context.Context, t *testing.T) {
	t.Helper()

	switch Partition() {
	case names.StandardPartitionID:
		PreCheckRegion(t, names.USEast1RegionID)
	case names.ChinaPartitionID:
		PreCheckRegion(t, names.CNNorthwest1RegionID)
	}

	conn := Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)
	input := &wafv2.ListWebACLsInput{
		Scope: wafv2types.ScopeCloudfront,
	}

	_, err := conn.ListWebACLs(ctx, input)

	if PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func ConfigAlternateAccountProvider() string {
	//lintignore:AT004
	return ConfigNamedAccountProvider(
		ProviderNameAlternate,
		os.Getenv(envvar.AlternateAccessKeyId),
		os.Getenv(envvar.AlternateProfile),
		os.Getenv(envvar.AlternateSecretAccessKey),
	)
}

func ConfigMultipleAccountProvider(t *testing.T, accounts int) string {
	t.Helper()

	var config strings.Builder

	if accounts > 3 {
		t.Fatalf("invalid number of Account configurations: %d", accounts)
	}

	if accounts >= 2 {
		config.WriteString(
			ConfigNamedAccountProvider(
				ProviderNameAlternate,
				os.Getenv(envvar.AlternateAccessKeyId),
				os.Getenv(envvar.AlternateProfile),
				os.Getenv(envvar.AlternateSecretAccessKey),
			),
		)
	}
	if accounts == 3 {
		config.WriteString(
			ConfigNamedAccountProvider(
				ProviderNameThird,
				os.Getenv(envvar.ThirdAccessKeyId),
				os.Getenv(envvar.ThirdProfile),
				os.Getenv(envvar.ThirdSecretAccessKey),
			),
		)
	}

	return config.String()
}

// ConfigNamedAccountProvider creates a new provider named configuration with a region.
//
// This can be used to build multiple provider configuration testing.
func ConfigNamedAccountProvider(providerName, accessKey, profile, secretKey string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider %[1]q {
  access_key = %[2]q
  profile    = %[3]q
  secret_key = %[4]q
}
`, providerName, accessKey, profile, secretKey)
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

// ConfigNamedRegionalProvider creates a new named provider configuration with a region.
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

func ConfigDefaultAndIgnoreTagsKeyPrefixes1(key1, value1, keyPrefix1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  default_tags {
    tags = {
      %[1]q = %[2]q
    }
  }
  ignore_tags {
    key_prefixes = [%[3]q]
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
      %[1]q = %[2]q
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

// ConfigRegionalProvider creates a new provider configuration with a region.
//
// This can only be used for single provider configuration testing as it
// overwrites the "aws" provider configuration.
func ConfigRegionalProvider(region string) string {
	return ConfigNamedRegionalProvider(ProviderName, region)
}

func ConfigAlternateAccountAlternateRegionProvider() string {
	return ConfigNamedAlternateAccountAlternateRegionProvider(ProviderNameAlternate)
}

func ConfigNamedAlternateAccountAlternateRegionProvider(providerName string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider %[1]q {
  access_key = %[2]q
  profile    = %[3]q
  region     = %[4]q
  secret_key = %[5]q
}
`, providerName, os.Getenv(envvar.AlternateAccessKeyId), os.Getenv(envvar.AlternateProfile), AlternateRegion(), os.Getenv(envvar.AlternateSecretAccessKey))
}

func RegionProviderFunc(region string, providers *[]*schema.Provider) ProviderFunc {
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

func NamedProviderFunc(name string, providers map[string]*schema.Provider) ProviderFunc {
	return func() *schema.Provider {
		return NamedProvider(name, providers)
	}
}

func NamedProvider(name string, providers map[string]*schema.Provider) *schema.Provider {
	if name == "" {
		log.Printf("[ERROR] No name passed")
	}

	p, ok := providers[name]
	if !ok {
		log.Printf("[ERROR] No provider named %q found", name)
		return nil
	}

	return p
}

func DeleteResource(ctx context.Context, resource *schema.Resource, d *schema.ResourceData, meta interface{}) error {
	if resource.DeleteContext != nil || resource.DeleteWithoutTimeout != nil {
		var diags diag.Diagnostics

		if resource.DeleteContext != nil {
			diags = resource.DeleteContext(ctx, d, meta) // nosemgrep:ci.semgrep.migrate.direct-CRUD-calls
		} else {
			diags = resource.DeleteWithoutTimeout(ctx, d, meta) // nosemgrep:ci.semgrep.migrate.direct-CRUD-calls
		}

		for i := range diags {
			if diags[i].Severity == diag.Error {
				return fmt.Errorf("deleting resource: %s", diags[i].Summary)
			}
		}

		return nil
	}

	return resource.Delete(d, meta) // nosemgrep:ci.semgrep.migrate.direct-CRUD-calls
}

func CheckResourceDisappears(ctx context.Context, provider *schema.Provider, resource *schema.Resource, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID missing: %s", n)
		}

		var state terraformsdk.InstanceState
		err := mapstructure.Decode(rs.Primary, &state)
		if err != nil {
			return err
		}

		return DeleteResource(ctx, resource, resource.Data(&state), provider.Meta())
	}
}

type TestCheckWithProviderFunc func(*terraform.State, *schema.Provider) error

func CheckWithProviders(f TestCheckWithProviderFunc, providers *[]*schema.Provider) resource.TestCheckFunc {
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

func CheckWithNamedProviders(f TestCheckWithProviderFunc, providers map[string]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		numberOfProviders := len(providers)
		for k, provo := range providers {
			if provo.Meta() == nil {
				log.Printf("[DEBUG] Skipping empty provider %q (total: %d)", k, numberOfProviders)
				continue
			}
			log.Printf("[DEBUG] Calling check with provider %q (total: %d)", k, numberOfProviders)
			if err := f(s, provo); err != nil {
				return err
			}
		}
		return nil
	}
}

// ErrorCheckSkipMessagesContaining skips tests based on error messages that contain one of the specified needles.
func ErrorCheckSkipMessagesContaining(t *testing.T, needles ...string) resource.ErrorCheckFunc {
	t.Helper()

	return func(err error) error {
		if err == nil {
			return nil
		}

		for _, needle := range needles {
			errorMessage := err.Error()
			if strings.Contains(errorMessage, needle) {
				t.Skipf("skipping test for %s/%s: %s", Partition(), Region(), errorMessage)
			}
		}

		return err
	}
}

// ErrorCheckSkipMessagesMatches skips tests based on error messages that match one of the specified regular expressions.
func ErrorCheckSkipMessagesMatches(t *testing.T, rs ...*regexp.Regexp) resource.ErrorCheckFunc {
	t.Helper()

	return func(err error) error {
		if err == nil {
			return nil
		}

		for _, r := range rs {
			errorMessage := err.Error()
			if r.MatchString(errorMessage) {
				t.Skipf("skipping test for %s/%s: %s", Partition(), Region(), errorMessage)
			}
		}

		return err
	}
}

type ServiceErrorCheckFunc func(*testing.T) resource.ErrorCheckFunc

var serviceErrorCheckFuncs map[string]ServiceErrorCheckFunc

func RegisterServiceErrorCheckFunc(serviceID string, f ServiceErrorCheckFunc) {
	if serviceErrorCheckFuncs == nil {
		serviceErrorCheckFuncs = make(map[string]ServiceErrorCheckFunc)
	}

	if _, ok := serviceErrorCheckFuncs[serviceID]; ok {
		// already registered
		panic(fmt.Sprintf("Cannot re-register a service! ServiceErrorCheckFunc exists for %s", serviceID)) //lintignore:R009
	}

	serviceErrorCheckFuncs[serviceID] = f
}

func ErrorCheck(t *testing.T, serviceIDs ...string) resource.ErrorCheckFunc {
	t.Helper()

	return func(err error) error {
		if err == nil {
			return nil
		}

		for _, serviceID := range serviceIDs {
			if f, ok := serviceErrorCheckFuncs[serviceID]; ok {
				ef := f(t)
				err = ef(err)
			}

			if err == nil {
				break
			}
		}

		if errorCheckCommon(err) {
			t.Skipf("skipping test for %s/%s: %s", Partition(), Region(), err.Error())
		}

		return err
	}
}

// NOTE: This function cannot use the standard tfawserr helpers
// as it is receiving error strings from the SDK testing framework,
// not actual error types from the resource logic.
func errorCheckCommon(err error) bool {
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
	if tfawserr.ErrCodeEquals(err, "AccessDeniedException") {
		return true
	}
	// Ignore missing API endpoints
	if tfawserr.ErrMessageContains(err, "RequestError", "send request failed") {
		return true
	}
	// Ignore unsupported API calls
	if tfawserr.ErrCodeEquals(err, "UnknownOperationException") {
		return true
	}
	if tfawserr.ErrCodeEquals(err, "UnsupportedOperation") {
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
	// ignore when not authorized to call API from account
	if tfawserr.ErrCodeEquals(err, "ForbiddenException") {
		return true
	}
	// Ignore missing API endpoints
	if errs.IsA[*net.DNSError](err) {
		return true
	}
	return false
}

func ConfigDefaultTags_Tags0() string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
		`
provider "aws" {
  skip_credentials_validation = true
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
      %[1]q = %[2]q
    }
  }

  skip_credentials_validation = true
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
      %[1]q = %[2]q
      %[3]q = %[4]q
    }
  }

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`, tag1, value1, tag2, value2))
}

func PreCheckAssumeRoleARN(t *testing.T) {
	t.Helper()

	envvar.SkipIfEmpty(t, envvar.AccAssumeRoleARN, "Amazon Resource Name (ARN) of existing IAM Role to assume for testing restricted permissions")
}

func ConfigAssumeRolePolicy(policy string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  assume_role {
    role_arn = %[1]q
    policy   = %[2]q
  }
}
`, os.Getenv(envvar.AccAssumeRoleARN), policy)
}

const testAccProviderConfigBase = `
data "aws_region" "provider_test" {}

# Required to initialize the provider.
data "aws_service" "provider_test" {
  region     = data.aws_region.provider_test.name
  service_id = "s3"
}
`

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
	return d.Subdomain(sdkacctest.RandString(8)) //nolint:mnd // standard length of 8
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

const (
	// ACM domain names cannot be longer than 64 characters
	// Other resources, e.g. Cognito User Pool Domains, limit this to 63
	acmCertificateDomainMaxLen = 63

	acmRandomSubDomainPrefix    = "tf-acc-"
	acmRandomSubDomainPrefixLen = len(acmRandomSubDomainPrefix)

	// Max length (63)
	// Subtract "tf-acc-" prefix (7)
	// Subtract "." between prefix and root domain (1)
	acmRandomSubDomainRemainderLen = acmCertificateDomainMaxLen - acmRandomSubDomainPrefixLen - 1
)

func ACMCertificateDomainFromEnv(t *testing.T) string {
	t.Helper()

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

	if len(rootDomain) > acmRandomSubDomainRemainderLen {
		t.Skipf(
			"Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is too long. "+
				"The domain must be %d characters or shorter to allow for "+
				"subdomain randomization in the testing.", acmRandomSubDomainRemainderLen)
	}

	return rootDomain
}

// ACM domain names cannot be longer than 64 characters
// Other resources, e.g. Cognito User Pool Domains, limit this to 63
func ACMCertificateRandomSubDomain(rootDomain string) string {
	return fmt.Sprintf(
		acmRandomSubDomainPrefix+"%s.%s",
		sdkacctest.RandString(acmRandomSubDomainRemainderLen-len(rootDomain)),
		rootDomain)
}

func CheckACMPCACertificateAuthorityActivateRootCA(ctx context.Context, certificateAuthority *acmpcatypes.CertificateAuthority) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := Provider.Meta().(*conns.AWSClient).ACMPCAClient(ctx)

		if v := certificateAuthority.Type; v != acmpcatypes.CertificateAuthorityTypeRoot {
			return fmt.Errorf("attempting to activate ACM PCA %s Certificate Authority", v)
		}

		arn := aws.StringValue(certificateAuthority.Arn)

		getCsrOutput, err := conn.GetCertificateAuthorityCsr(ctx, &acmpca.GetCertificateAuthorityCsrInput{
			CertificateAuthorityArn: aws.String(arn),
		})

		if err != nil {
			return fmt.Errorf("getting ACM PCA Certificate Authority (%s) CSR: %w", arn, err)
		}

		issueCertOutput, err := conn.IssueCertificate(ctx, &acmpca.IssueCertificateInput{
			CertificateAuthorityArn: aws.String(arn),
			Csr:                     []byte(aws.StringValue(getCsrOutput.Csr)),
			IdempotencyToken:        aws.String(id.UniqueId()),
			SigningAlgorithm:        certificateAuthority.CertificateAuthorityConfiguration.SigningAlgorithm,
			TemplateArn:             aws.String(fmt.Sprintf("arn:%s:acm-pca:::template/RootCACertificate/V1", Partition())),
			Validity: &acmpcatypes.Validity{
				Type:  acmpcatypes.ValidityPeriodTypeYears,
				Value: aws.Int64(10),
			},
		})

		if err != nil {
			return fmt.Errorf("issuing ACM PCA Certificate Authority (%s) Root CA certificate from CSR: %w", arn, err)
		}

		// Wait for certificate status to become ISSUED.
		waiter := acmpca.NewCertificateIssuedWaiter(conn)
		params := &acmpca.GetCertificateInput{
			CertificateAuthorityArn: aws.String(arn),
			CertificateArn:          issueCertOutput.CertificateArn,
		}

		err = waiter.Wait(ctx, params, CertificateIssueTimeout)

		if err != nil {
			return fmt.Errorf("waiting for ACM PCA Certificate Authority (%s) Root CA certificate to become ISSUED: %w", arn, err)
		}

		getCertOutput, err := conn.GetCertificate(ctx, &acmpca.GetCertificateInput{
			CertificateAuthorityArn: aws.String(arn),
			CertificateArn:          issueCertOutput.CertificateArn,
		})

		if err != nil {
			return fmt.Errorf("getting ACM PCA Certificate Authority (%s) issued Root CA certificate: %w", arn, err)
		}

		_, err = conn.ImportCertificateAuthorityCertificate(ctx, &acmpca.ImportCertificateAuthorityCertificateInput{
			CertificateAuthorityArn: aws.String(arn),
			Certificate:             []byte(aws.StringValue(getCertOutput.Certificate)),
		})

		if err != nil {
			return fmt.Errorf("importing ACM PCA Certificate Authority (%s) Root CA certificate: %w", arn, err)
		}

		return err
	}
}

func CheckACMPCACertificateAuthorityActivateSubordinateCA(ctx context.Context, rootCertificateAuthority, certificateAuthority *acmpcatypes.CertificateAuthority) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := Provider.Meta().(*conns.AWSClient).ACMPCAClient(ctx)

		if v := certificateAuthority.Type; v != acmpcatypes.CertificateAuthorityTypeSubordinate {
			return fmt.Errorf("attempting to activate ACM PCA %s Certificate Authority", v)
		}

		arn := aws.StringValue(certificateAuthority.Arn)

		getCsrOutput, err := conn.GetCertificateAuthorityCsr(ctx, &acmpca.GetCertificateAuthorityCsrInput{
			CertificateAuthorityArn: aws.String(arn),
		})

		if err != nil {
			return fmt.Errorf("getting ACM PCA Certificate Authority (%s) CSR: %w", arn, err)
		}

		rootCertificateAuthorityArn := aws.StringValue(rootCertificateAuthority.Arn)

		issueCertOutput, err := conn.IssueCertificate(ctx, &acmpca.IssueCertificateInput{
			CertificateAuthorityArn: aws.String(rootCertificateAuthorityArn),
			Csr:                     []byte(aws.StringValue(getCsrOutput.Csr)),
			IdempotencyToken:        aws.String(id.UniqueId()),
			SigningAlgorithm:        certificateAuthority.CertificateAuthorityConfiguration.SigningAlgorithm,
			TemplateArn:             aws.String(fmt.Sprintf("arn:%s:acm-pca:::template/SubordinateCACertificate_PathLen0/V1", Partition())),
			Validity: &acmpcatypes.Validity{
				Type:  acmpcatypes.ValidityPeriodTypeYears,
				Value: aws.Int64(3),
			},
		})

		if err != nil {
			return fmt.Errorf("issuing ACM PCA Certificate Authority (%s) Subordinate CA certificate from CSR: %w", arn, err)
		}

		// Wait for certificate status to become ISSUED.
		waiter := acmpca.NewCertificateIssuedWaiter(conn)
		params := &acmpca.GetCertificateInput{
			CertificateAuthorityArn: aws.String(rootCertificateAuthorityArn),
			CertificateArn:          issueCertOutput.CertificateArn,
		}

		err = waiter.Wait(ctx, params, CertificateIssueTimeout)

		if err != nil {
			return fmt.Errorf("waiting for ACM PCA Certificate Authority (%s) Subordinate CA certificate (%s) to become ISSUED: %w", arn, aws.StringValue(issueCertOutput.CertificateArn), err)
		}

		getCertOutput, err := conn.GetCertificate(ctx, &acmpca.GetCertificateInput{
			CertificateAuthorityArn: aws.String(rootCertificateAuthorityArn),
			CertificateArn:          issueCertOutput.CertificateArn,
		})

		if err != nil {
			return fmt.Errorf("getting ACM PCA Certificate Authority (%s) issued Subordinate CA certificate: %w", arn, err)
		}

		_, err = conn.ImportCertificateAuthorityCertificate(ctx, &acmpca.ImportCertificateAuthorityCertificateInput{
			CertificateAuthorityArn: aws.String(arn),
			Certificate:             []byte(aws.StringValue(getCertOutput.Certificate)),
			CertificateChain:        []byte(aws.StringValue(getCertOutput.CertificateChain)),
		})

		if err != nil {
			return fmt.Errorf("importing ACM PCA Certificate Authority (%s) Subordinate CA certificate: %w", arn, err)
		}

		return err
	}
}

func CheckACMPCACertificateAuthorityDisableCA(ctx context.Context, certificateAuthority *acmpcatypes.CertificateAuthority) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := Provider.Meta().(*conns.AWSClient).ACMPCAClient(ctx)

		_, err := conn.UpdateCertificateAuthority(ctx, &acmpca.UpdateCertificateAuthorityInput{
			CertificateAuthorityArn: certificateAuthority.Arn,
			Status:                  acmpcatypes.CertificateAuthorityStatusDisabled,
		})

		return err
	}
}

func CheckACMPCACertificateAuthorityExists(ctx context.Context, n string, certificateAuthority *acmpcatypes.CertificateAuthority) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ACM PCA Certificate Authority ID is set")
		}

		conn := Provider.Meta().(*conns.AWSClient).ACMPCAClient(ctx)

		output, err := tfacmpca.FindCertificateAuthorityByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*certificateAuthority = *output

		return nil
	}
}

// PreCheckAPIGatewayTypeEDGE checks if endpoint config type EDGE can be used in a test and skips test if not (i.e., not in standard partition).
func PreCheckAPIGatewayTypeEDGE(t *testing.T) {
	t.Helper()

	if Partition() != names.StandardPartitionID {
		t.Skipf("skipping test; Endpoint Configuration type EDGE is not supported in this partition (%s)", Partition())
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
//   - data.aws_availability_zones.available.names[0]
//   - aws_subnet.test.availability_zone
//   - us-west-2a
//
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

func configLatestAmazonLinux2HVMEBSAMI(architecture ec2types.ArchitectureValues) string {
	return fmt.Sprintf(`
data "aws_ami" "amzn2-ami-minimal-hvm-ebs-%[1]s" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "architecture"
    values = [%[1]q]
  }
}
`, architecture)
}

// ConfigLatestAmazonLinux2HVMEBSX8664AMI returns the configuration for a data source that
// describes the latest Amazon Linux 2 x86_64 AMI using HVM virtualization and an EBS root device.
// The data source is named 'amzn2-ami-minimal-hvm-ebs-x86_64'.
func ConfigLatestAmazonLinux2HVMEBSX8664AMI() string {
	return configLatestAmazonLinux2HVMEBSAMI(ec2types.ArchitectureValuesX8664)
}

// ConfigLatestAmazonLinux2HVMEBSARM64AMI returns the configuration for a data source that
// describes the latest Amazon Linux 2 arm64 AMI using HVM virtualization and an EBS root device.
// The data source is named 'amzn2-ami-minimal-hvm-ebs-arm64'.
func ConfigLatestAmazonLinux2HVMEBSARM64AMI() string {
	return configLatestAmazonLinux2HVMEBSAMI(ec2types.ArchitectureValuesArm64)
}

func ConfigLambdaBase(policyName, roleName, sgName string) string {
	return ConfigCompose(ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = %[1]q
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
        "ec2:DeleteNetworkInterface",
        "ec2:AssignPrivateIpAddresses",
        "ec2:UnassignPrivateIpAddresses"
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
  name = %[2]q

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
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[3]q
  }
}

resource "aws_subnet" "subnet_for_lambda" {
  vpc_id                          = aws_vpc.vpc_for_lambda.id
  cidr_block                      = cidrsubnet(aws_vpc.vpc_for_lambda.cidr_block, 8, 1)
  availability_zone               = data.aws_availability_zones.available.names[1]
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.vpc_for_lambda.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[3]q
  }
}

# This is defined here, rather than only in test cases where it's needed is to
# prevent a timeout issue when fully removing Lambda Filesystems
resource "aws_subnet" "subnet_for_lambda_az2" {
  vpc_id                          = aws_vpc.vpc_for_lambda.id
  cidr_block                      = cidrsubnet(aws_vpc.vpc_for_lambda.cidr_block, 8, 2)
  availability_zone               = data.aws_availability_zones.available.names[1]
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.vpc_for_lambda.ipv6_cidr_block, 8, 2)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[3]q
  }
}

resource "aws_security_group" "sg_for_lambda" {
  name        = %[3]q
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

  tags = {
    Name = %[3]q
  }
}
`, policyName, roleName, sgName))
}

func ConfigVPCWithSubnets(rName string, subnetCount int) string {
	return ConfigCompose(
		ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = %[2]d

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}
`, rName, subnetCount),
	)
}

func ConfigVPCWithSubnetsEnableDNSHostnames(rName string, subnetCount int) string {
	return ConfigCompose(
		ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = %[2]d

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}
`, rName, subnetCount),
	)
}

func ConfigVPCWithSubnetsIPv6(rName string, subnetCount int) string {
	return ConfigCompose(
		ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = %[2]d

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[1]q
  }
}
`, rName, subnetCount),
	)
}

func CheckVPCExists(ctx context.Context, n string, v *ec2types.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no VPC ID is set")
		}

		conn := Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindVPCByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func CheckCallerIdentityAccountID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find AccountID resource: %s", n)
		}

		expected := Provider.Meta().(*conns.AWSClient).AccountID
		if rs.Primary.Attributes["account_id"] != expected {
			return fmt.Errorf("incorrect Account ID: expected %q, got %q", expected, rs.Primary.Attributes["account_id"])
		}

		if rs.Primary.Attributes["user_id"] == "" {
			return fmt.Errorf("user_id expected to not be nil")
		}

		if rs.Primary.Attributes["arn"] == "" {
			return fmt.Errorf("attribute ARN expected to not be nil")
		}

		return nil
	}
}

func CheckResourceAttrGreaterThanValue(n, key string, val int) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(n, key, func(value string) error {
		v, err := strconv.Atoi(value)

		if err != nil {
			return err
		}

		if v <= val {
			return fmt.Errorf("got %d, want > %d", v, val)
		}

		return nil
	})
}

func CheckResourceAttrGreaterThanOrEqualValue(n, key string, val int) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(n, key, func(value string) error {
		v, err := strconv.Atoi(value)

		if err != nil {
			return err
		}

		if v < val {
			return fmt.Errorf("got %d, want >= %d", v, val)
		}

		return nil
	})
}

func CheckResourceAttrIsJSONString(n, key string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(n, key, func(value string) error {
		var m map[string]*json.RawMessage

		if err := json.Unmarshal([]byte(value), &m); err != nil {
			return err
		}

		if len(m) == 0 {
			return errors.New(`empty JSON string`)
		}

		return nil
	})
}

// SkipIfEnvVarNotSet skips the current test if the specified environment variable is not set.
// The variable's value is returned.
func SkipIfEnvVarNotSet(t *testing.T, key string) string {
	t.Helper()

	v := os.Getenv(key)
	if v == "" {
		t.Skipf("Environment variable %s is not set, skipping test", key)
	}
	return v
}

// SkipIfExeNotOnPath skips the current test if the specified executable is not found in the directories named by the PATH environment variable.
// The absolute path to the executable is returned.
func SkipIfExeNotOnPath(t *testing.T, file string) string {
	t.Helper()

	v, err := exec.LookPath(file)
	if err != nil {
		t.Skipf("File %s not found on PATH, skipping test: %s", v, err)
	}
	return v
}

// RunSerialTests1Level runs test cases in parallel, optionally sleeping between each.
func RunSerialTests1Level(t *testing.T, testCases map[string]func(*testing.T), d time.Duration) {
	t.Helper()

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
			time.Sleep(d)
		})
	}
}

// RunSerialTests2Levels runs test cases in parallel, optionally sleeping between each.
func RunSerialTests2Levels(t *testing.T, testCases map[string]map[string]func(*testing.T), d time.Duration) {
	t.Helper()

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			RunSerialTests1Level(t, m, d)
		})
	}
}

// RunLimitedConcurrencyTests2Levels runs test cases with concurrency limited via `semaphore`.
func RunLimitedConcurrencyTests2Levels(t *testing.T, semaphore tfsync.Semaphore, testCases map[string]map[string]func(*testing.T, tfsync.Semaphore)) {
	t.Helper()

	for group, m := range testCases {
		m := m
		for name, tc := range m {
			tc := tc
			t.Run(fmt.Sprintf("%s_%s", group, name), func(t *testing.T) {
				t.Cleanup(func() {
					if os.Getenv(resource.EnvTfAcc) != "" {
						semaphore.Notify()
					}
				})
				tc(t, semaphore)
			})
		}
	}
}

// TestNoMatchResourceAttr ensures a value matching a regular expression is
// NOT stored in state for the given name and key combination. Same as resource.TestMatchResourceAttr()
// except negative.
func TestNoMatchResourceAttr(name, key string, r *regexp.Regexp) resource.TestCheckFunc {
	return checkIfIndexesIntoTypeSet(key, func(s *terraform.State) error {
		is, err := primaryInstanceState(s, name)
		if err != nil {
			return err
		}

		return testNoMatchResourceAttr(is, name, key, r)
	})
}

// testNoMatchResourceAttr is same as testMatchResourceAttr in
// github.com/hashicorp/terraform-plugin-testing/helper/resource
// except negative.
func testNoMatchResourceAttr(is *terraform.InstanceState, name string, key string, r *regexp.Regexp) error {
	if r.MatchString(is.Attributes[key]) {
		return fmt.Errorf(
			"%s: Attribute '%s' did match %q and should not, got %#v",
			name,
			key,
			r.String(),
			is.Attributes[key])
	}

	return nil
}

// checkIfIndexesIntoTypeSet is copied from
// https://github.com/hashicorp/terraform-plugin-testing/blob/dee4bfbbfd4911cf69a6c9917a37ecd8faa41ae9/helper/resource/testing.go#L1689
func checkIfIndexesIntoTypeSet(key string, f resource.TestCheckFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		err := f(s)
		if err != nil && indexesIntoTypeSet(key) {
			return fmt.Errorf("Error in test check: %s\nTest check address %q likely indexes into TypeSet\nThis is currently not possible in the SDK", err, key)
		}
		return err
	}
}

// indexesIntoTypeSet is copied from
// https://github.com/hashicorp/terraform-plugin-testing/blob/dee4bfbbfd4911cf69a6c9917a37ecd8faa41ae9/helper/resource/testing.go#L1680
func indexesIntoTypeSet(key string) bool {
	for _, part := range strings.Split(key, ".") {
		if i, err := strconv.Atoi(part); err == nil && i > 100 {
			return true
		}
	}
	return false
}

// primaryInstanceState is copied from
// https://github.com/hashicorp/terraform-plugin-testing/blob/dee4bfbbfd4911cf69a6c9917a37ecd8faa41ae9/helper/resource/testing.go#L1672
func primaryInstanceState(s *terraform.State, name string) (*terraform.InstanceState, error) {
	ms := s.RootModule()
	return modulePrimaryInstanceState(ms, name)
}

// modulePrimaryInstanceState is copied from
// https://github.com/hashicorp/terraform-plugin-testing/blob/dee4bfbbfd4911cf69a6c9917a37ecd8faa41ae9/helper/resource/testing.go#L1645
func modulePrimaryInstanceState(ms *terraform.ModuleState, name string) (*terraform.InstanceState, error) {
	rs, ok := ms.Resources[name]
	if !ok {
		return nil, fmt.Errorf("Not found: %s in %s", name, ms.Path)
	}

	is := rs.Primary
	if is == nil {
		return nil, fmt.Errorf("No primary instance: %s in %s", name, ms.Path)
	}

	return is, nil
}

func ExpectErrorAttrAtLeastOneOf(attrs ...string) *regexp.Regexp {
	return regexache.MustCompile(fmt.Sprintf("one of\\s+`%s`\\s+must be specified", strings.Join(attrs, ",")))
}

func ExpectErrorAttrMinItems(attr string, expected, actual int) *regexp.Regexp {
	return regexache.MustCompile(fmt.Sprintf(`Attribute %s requires %d\s+item minimum, but config has only %d declared`, attr, expected, actual))
}
