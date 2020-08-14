package aws

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccCheckAwsCloudFrontCachePolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_cache_policy" {
			continue
		}

		id := rs.Primary.ID

		switch _, _, ok, err := getAwsCloudFrontCachePolicy(context.Background(), conn, id); {
		case err != nil:
			return err
		case ok:
			return fmt.Errorf("cache policy %s still exists", id)
		}
	}

	return nil
}

func testAccAwsCloudFrontCachePolicyExists(
	name string,
	out *cloudfront.CachePolicy,
	etag *string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		switch {
		case !ok:
			return fmt.Errorf("resource %s not found", name)
		case rs.Primary.ID == "":
			return fmt.Errorf("resource %s has not set its id", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn
		id := rs.Primary.ID

		cachePolicy, et, ok, err := getAwsCloudFrontCachePolicy(context.Background(), conn, id)
		switch {
		case err != nil:
			return err
		case !ok:
			return fmt.Errorf("resource %s (%s) has not been created", name, id)
		}

		if out != nil {
			*out = *cachePolicy
			*etag = et
		}

		return nil
	}
}

func TestAccAwsCloudFrontCachePolicy_basic(t *testing.T) {
	resourceName := "aws_cloudfront_cache_policy.test"
	randomName := "Terraform-AccTest-" + acctest.RandString(8)
	cachePolicy, etag := cloudfront.CachePolicy{}, ""

	checkAttributes := func(*terraform.State) error {
		sortStringPtrs := func(slice []*string) {
			sort.Slice(slice, func(i, j int) bool {
				return *slice[i] < *slice[j]
			})
		}

		actual := *cachePolicy.CachePolicyConfig
		sortStringPtrs(actual.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies.Items)
		sortStringPtrs(actual.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.Headers.Items)
		sortStringPtrs(actual.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings.Items)

		expect := cloudfront.CachePolicyConfig{
			Comment:    aws.String("Greetings, programs!"),
			DefaultTTL: aws.Int64(3600),
			MaxTTL:     aws.Int64(86400),
			MinTTL:     aws.Int64(0),
			Name:       aws.String(randomName + "-Suffix"),
			ParametersInCacheKeyAndForwardedToOrigin: &cloudfront.ParametersInCacheKeyAndForwardedToOrigin{
				CookiesConfig: &cloudfront.CachePolicyCookiesConfig{
					CookieBehavior: aws.String("whitelist"),
					Cookies: &cloudfront.CookieNames{
						Items:    aws.StringSlice([]string{"Cookie1", "Cookie2"}),
						Quantity: aws.Int64(2),
					},
				},
				EnableAcceptEncodingGzip: aws.Bool(true),
				HeadersConfig: &cloudfront.CachePolicyHeadersConfig{
					HeaderBehavior: aws.String("whitelist"),
					Headers: &cloudfront.Headers{
						Items:    aws.StringSlice([]string{"X-Header-1", "X-Header-2"}),
						Quantity: aws.Int64(2),
					},
				},
				QueryStringsConfig: &cloudfront.CachePolicyQueryStringsConfig{
					QueryStringBehavior: aws.String("whitelist"),
					QueryStrings: &cloudfront.QueryStringNames{
						Items:    aws.StringSlice([]string{"test"}),
						Quantity: aws.Int64(1),
					},
				},
			},
		}

		if !reflect.DeepEqual(actual, expect) {
			return fmt.Errorf("Expected CachePolicyConfig:\n%#v\nGot:\n%#v", expect, actual)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudFrontCachePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsCloudFrontCachePolicyConfig_basic_create(randomName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsCloudFrontCachePolicyExists(resourceName, &cachePolicy, &etag),
					resource.TestCheckResourceAttr(resourceName, "comment", "Greetings, programs!"),
					resource.TestCheckResourceAttr(resourceName, "cookie_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "cookie_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_ttl", "86400"),
					resource.TestCheckResourceAttr(resourceName, "enable_accept_encoding_gzip", "false"),
					resource.TestCheckResourceAttrPtr(resourceName, "etag", &etag),
					resource.TestCheckResourceAttr(resourceName, "header_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "header_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_ttl", "31536000"),
					resource.TestCheckResourceAttr(resourceName, "min_ttl", "5"),
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "query_string_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "query_string_names.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsCloudFrontCachePolicyConfig_basic_update(randomName + "-Suffix"),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsCloudFrontCachePolicyExists(resourceName, &cachePolicy, &etag),
					checkAttributes,
					resource.TestCheckResourceAttr(resourceName, "comment", "Greetings, programs!"),
					resource.TestCheckResourceAttr(resourceName, "cookie_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "cookie_names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_ttl", "3600"),
					resource.TestCheckResourceAttr(resourceName, "enable_accept_encoding_gzip", "true"),
					resource.TestCheckResourceAttrPtr(resourceName, "etag", &etag),
					resource.TestCheckResourceAttr(resourceName, "header_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "header_names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "max_ttl", "86400"),
					resource.TestCheckResourceAttr(resourceName, "min_ttl", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", randomName+"-Suffix"),
					resource.TestCheckResourceAttr(resourceName, "query_string_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "query_string_names.#", "1"),
				),
			},
		},
	})
}

func testAccAwsCloudFrontCachePolicyConfig_basic_create(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "test" {
	comment = "Greetings, programs!"
	min_ttl = 5
	name    = %[1]q
}
`, name)

}

func testAccAwsCloudFrontCachePolicyConfig_basic_update(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "test" {
	comment                     = "Greetings, programs!"
	cookie_behavior             = "whitelist"
	cookie_names                = ["Cookie1", "Cookie2"]
	default_ttl                 = 3600
	enable_accept_encoding_gzip = true
	header_behavior             = "whitelist"
	header_names                = ["X-Header-1", "X-Header-2"]
	max_ttl                     = 86400
	min_ttl                     = 0
	name                        = %[1]q
	query_string_behavior       = "whitelist"
	query_string_names          = ["test"]
}
`, name)
}

func TestAccAwsCloudFrontCachePolicy_disappears(t *testing.T) {
	resourceName := "aws_cloudfront_cache_policy.test"
	randomName := "Terraform-AccTest-" + acctest.RandString(8)
	cachePolicy, etag := cloudfront.CachePolicy{}, ""

	checkDisappears := func(*terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn
		input := cloudfront.DeleteCachePolicyInput{
			Id:      aws.String(*cachePolicy.Id),
			IfMatch: aws.String(etag),
		}
		_, err := conn.DeleteCachePolicy(&input)
		return err
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudFrontCachePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsCloudFrontCachePolicyConfig_disappears(randomName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsCloudFrontCachePolicyExists(resourceName, &cachePolicy, &etag),
					checkDisappears,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsCloudFrontCachePolicyConfig_disappears(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_cache_policy" "test" {
	comment = "Greetings, programs!"
	min_ttl = 5
	name    = %[1]q
}
`, name)

}
