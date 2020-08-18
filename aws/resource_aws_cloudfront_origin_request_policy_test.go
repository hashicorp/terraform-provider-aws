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

func testAccCheckAwsCloudFrontOriginRequestPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_origin_request_policy" {
			continue
		}

		id := rs.Primary.ID

		switch _, _, ok, err := getAwsCloudFrontOriginRequestPolicy(context.Background(), conn, id); {
		case err != nil:
			return err
		case ok:
			return fmt.Errorf("origin request policy %s still exists", id)
		}
	}

	return nil
}

func testAccAwsCloudFrontOriginRequestPolicyExists(
	name string,
	out *cloudfront.OriginRequestPolicy,
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

		policy, et, ok, err := getAwsCloudFrontOriginRequestPolicy(context.Background(), conn, id)
		switch {
		case err != nil:
			return err
		case !ok:
			return fmt.Errorf("resource %s (%s) has not been created", name, id)
		}

		if out != nil {
			*out = *policy
			*etag = et
		}

		return nil
	}
}

func TestAccAwsCloudFrontOriginRequestPolicy_basic(t *testing.T) {
	resourceName := "aws_cloudfront_origin_request_policy.test"
	randomName := "Terraform-AccTest-" + acctest.RandString(8)
	policy, etag := cloudfront.OriginRequestPolicy{}, ""

	checkAttributes := func(*terraform.State) error {
		sortStringPtrs := func(slice []*string) {
			sort.Slice(slice, func(i, j int) bool {
				return *slice[i] < *slice[j]
			})
		}

		actual := *policy.OriginRequestPolicyConfig
		sortStringPtrs(actual.CookiesConfig.Cookies.Items)
		sortStringPtrs(actual.HeadersConfig.Headers.Items)
		sortStringPtrs(actual.QueryStringsConfig.QueryStrings.Items)

		expect := cloudfront.OriginRequestPolicyConfig{
			Comment: aws.String("Greetings, programs!"),
			Name:    aws.String(randomName + "-Suffix"),
			CookiesConfig: &cloudfront.OriginRequestPolicyCookiesConfig{
				CookieBehavior: aws.String("whitelist"),
				Cookies: &cloudfront.CookieNames{
					Items:    aws.StringSlice([]string{"Cookie1", "Cookie2"}),
					Quantity: aws.Int64(2),
				},
			},
			HeadersConfig: &cloudfront.OriginRequestPolicyHeadersConfig{
				HeaderBehavior: aws.String("whitelist"),
				Headers: &cloudfront.Headers{
					Items:    aws.StringSlice([]string{"X-Header-1", "X-Header-2"}),
					Quantity: aws.Int64(2),
				},
			},
			QueryStringsConfig: &cloudfront.OriginRequestPolicyQueryStringsConfig{
				QueryStringBehavior: aws.String("whitelist"),
				QueryStrings: &cloudfront.QueryStringNames{
					Items:    aws.StringSlice([]string{"test"}),
					Quantity: aws.Int64(1),
				},
			},
		}

		if !reflect.DeepEqual(actual, expect) {
			return fmt.Errorf("Expected OriginRequestPolicyConfig:\n%#v\nGot:\n%#v", expect, actual)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudFrontOriginRequestPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsCloudFrontOriginRequestPolicyConfig_basic_create(randomName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsCloudFrontOriginRequestPolicyExists(resourceName, &policy, &etag),
					resource.TestCheckResourceAttr(resourceName, "comment", "Greetings, programs!"),
					resource.TestCheckResourceAttr(resourceName, "cookie_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "cookie_names.#", "0"),
					resource.TestCheckResourceAttrPtr(resourceName, "etag", &etag),
					resource.TestCheckResourceAttr(resourceName, "header_behavior", "none"),
					resource.TestCheckResourceAttr(resourceName, "header_names.#", "0"),
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
				Config:       testAccAwsCloudFrontOriginRequestPolicyConfig_basic_update(randomName + "-Suffix"),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsCloudFrontOriginRequestPolicyExists(resourceName, &policy, &etag),
					checkAttributes,
					resource.TestCheckResourceAttr(resourceName, "comment", "Greetings, programs!"),
					resource.TestCheckResourceAttr(resourceName, "cookie_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "cookie_names.#", "2"),
					resource.TestCheckResourceAttrPtr(resourceName, "etag", &etag),
					resource.TestCheckResourceAttr(resourceName, "header_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "header_names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", randomName+"-Suffix"),
					resource.TestCheckResourceAttr(resourceName, "query_string_behavior", "whitelist"),
					resource.TestCheckResourceAttr(resourceName, "query_string_names.#", "1"),
				),
			},
		},
	})
}

func testAccAwsCloudFrontOriginRequestPolicyConfig_basic_create(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_request_policy" "test" {
	comment               = "Greetings, programs!"
	cookie_behavior       = "none"
	header_behavior       = "none"
	name                  = %[1]q
	query_string_behavior = "none"
}
`, name)

}

func testAccAwsCloudFrontOriginRequestPolicyConfig_basic_update(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_request_policy" "test" {
	comment                     = "Greetings, programs!"
	cookie_behavior             = "whitelist"
	cookie_names                = ["Cookie1", "Cookie2"]
	header_behavior             = "whitelist"
	header_names                = ["X-Header-1", "X-Header-2"]
	name                        = %[1]q
	query_string_behavior       = "whitelist"
	query_string_names          = ["test"]
}
`, name)
}

func TestAccAwsCloudFrontOriginRequestPolicy_disappears(t *testing.T) {
	resourceName := "aws_cloudfront_origin_request_policy.test"
	randomName := "Terraform-AccTest-" + acctest.RandString(8)
	policy, etag := cloudfront.OriginRequestPolicy{}, ""

	checkDisappears := func(*terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn
		input := cloudfront.DeleteOriginRequestPolicyInput{
			Id:      aws.String(*policy.Id),
			IfMatch: aws.String(etag),
		}
		_, err := conn.DeleteOriginRequestPolicy(&input)
		return err
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudFrontOriginRequestPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsCloudFrontOriginRequestPolicyConfig_disappears(randomName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsCloudFrontOriginRequestPolicyExists(resourceName, &policy, &etag),
					checkDisappears,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsCloudFrontOriginRequestPolicyConfig_disappears(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_origin_request_policy" "test" {
	comment               = "Greetings, programs!"
	cookie_behavior       = "none"
	header_behavior       = "none"
	name                  = %[1]q
	query_string_behavior = "none"
}
`, name)

}
