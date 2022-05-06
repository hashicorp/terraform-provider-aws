package appsync_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
)

func testAccAPIKey_basic(t *testing.T) {
	var apiKey appsync.ApiKey
	dateAfterSevenDays := time.Now().UTC().Add(time.Hour * 24 * time.Duration(7)).Truncate(time.Hour)
	resourceName := "aws_appsync_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(resourceName, &apiKey),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					testAccCheckAPIKeyExpiresDate(&apiKey, dateAfterSevenDays),
					resource.TestMatchResourceAttr(resourceName, "key", regexp.MustCompile(`.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAPIKey_description(t *testing.T) {
	var apiKey appsync.ApiKey
	resourceName := "aws_appsync_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(resourceName, &apiKey),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccAPIKeyConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(resourceName, &apiKey),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAPIKey_expires(t *testing.T) {
	var apiKey appsync.ApiKey
	dateAfterTenDays := time.Now().UTC().Add(time.Hour * 24 * time.Duration(10)).Truncate(time.Hour)
	dateAfterTwentyDays := time.Now().UTC().Add(time.Hour * 24 * time.Duration(20)).Truncate(time.Hour)
	resourceName := "aws_appsync_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_expires(rName, dateAfterTenDays.Format(time.RFC3339)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(resourceName, &apiKey),
					testAccCheckAPIKeyExpiresDate(&apiKey, dateAfterTenDays),
				),
			},
			{
				Config: testAccAPIKeyConfig_expires(rName, dateAfterTwentyDays.Format(time.RFC3339)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(resourceName, &apiKey),
					testAccCheckAPIKeyExpiresDate(&apiKey, dateAfterTwentyDays),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAPIKeyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appsync_api_key" {
			continue
		}

		apiID, keyID, err := tfappsync.DecodeAPIKeyID(rs.Primary.ID)
		if err != nil {
			return err
		}

		apiKey, err := tfappsync.GetAPIKey(apiID, keyID, conn)
		if err == nil {
			if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
				return nil
			}
			return err
		}

		if apiKey != nil && aws.StringValue(apiKey.Id) == keyID {
			return fmt.Errorf("Appsync API Key ID %q still exists", rs.Primary.ID)
		}

		return nil

	}
	return nil
}

func testAccCheckAPIKeyExists(resourceName string, apiKey *appsync.ApiKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Appsync API Key Not found in state: %s", resourceName)
		}

		apiID, keyID, err := tfappsync.DecodeAPIKeyID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn
		key, err := tfappsync.GetAPIKey(apiID, keyID, conn)
		if err != nil {
			return err
		}

		if key == nil || key.Id == nil {
			return fmt.Errorf("Appsync API Key %q not found", rs.Primary.ID)
		}

		*apiKey = *key

		return nil
	}
}

func testAccCheckAPIKeyExpiresDate(apiKey *appsync.ApiKey, expectedTime time.Time) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiKeyExpiresTime := time.Unix(aws.Int64Value(apiKey.Expires), 0)
		if !apiKeyExpiresTime.Equal(expectedTime) {
			return fmt.Errorf("Appsync API Key expires difference: got %s and expected %s", apiKeyExpiresTime.Format(time.RFC3339), expectedTime.Format(time.RFC3339))
		}

		return nil
	}
}

func testAccAPIKeyConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}

resource "aws_appsync_api_key" "test" {
  api_id      = aws_appsync_graphql_api.test.id
  description = %q
}
`, rName, description)
}

func testAccAPIKeyConfig_expires(rName, expires string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}

resource "aws_appsync_api_key" "test" {
  api_id  = aws_appsync_graphql_api.test.id
  expires = %q
}
`, rName, expires)
}

func testAccAPIKeyConfig_required(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}

resource "aws_appsync_api_key" "test" {
  api_id = aws_appsync_graphql_api.test.id
}
`, rName)
}
