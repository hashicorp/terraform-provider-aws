package aws

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsAppsyncApiKey_basic(t *testing.T) {
	// sample date to test
	dateAfterOneYear := time.Now().Add(time.Hour * 24 * time.Duration(360)).Format("02/01/2006")
	// test sample date against time of expiry
	layout := "02/01/2006 15:04:05 -0700 MST"
	tx := strings.Split(time.Now().Format(layout), " ")
	tx[0] = dateAfterOneYear
	timeAfterOneYear, _ := time.Parse(layout, strings.Join(tx, " "))

	thirtyDays := "30"
	timeAfterThirdyDays := time.Now().Add(time.Hour * 24 * 30)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncApiKeyConfigValidTillDate(dateAfterOneYear),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncApiKeyExistsTillDate(
						"aws_appsync_graphql_api.test1",
						"aws_appsync_api_key.test_valid_till_date",
						timeAfterOneYear.Unix(),
					),
				),
			},
			{
				Config: testAccAppsyncApiKeyConfigValidityPeriodDays(thirtyDays),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncApiKeyExistsTillDate(
						"aws_appsync_graphql_api.test2",
						"aws_appsync_api_key.test_validity_period_days",
						timeAfterThirdyDays.Unix(),
					),
				),
			},
		},
	})
}

func testAccCheckAwsAppsyncApiKeyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appsyncconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appsync_api_key" {
			continue
		}

		describe, err := conn.ListApiKeys(&appsync.ListApiKeysInput{})

		if err == nil {
			if len(describe.ApiKeys) != 0 &&
				*describe.ApiKeys[0].Id == rs.Primary.ID {
				return fmt.Errorf("Appsync ApiKey still exists")
			}
			return err
		}

	}
	return nil
}

func testAccCheckAwsAppsyncApiKeyExistsTillDate(GqlApiName string, ApiKeyName string, date int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rsGql, ok := s.RootModule().Resources[GqlApiName]
		if !ok {
			return fmt.Errorf("Gql Not found in state: %s", GqlApiName)
		}

		rsApiKey, ok := s.RootModule().Resources[ApiKeyName]
		if !ok {
			return fmt.Errorf("Key Not found in state: %s", ApiKeyName)
		}

		conn := testAccProvider.Meta().(*AWSClient).appsyncconn

		input := &appsync.ListApiKeysInput{
			ApiId: aws.String(rsGql.Primary.ID),
		}

		resp, err := conn.ListApiKeys(input)
		if err != nil {
			return err
		}
		var key appsync.ApiKey
		for _, v := range resp.ApiKeys {
			if *v.Id == *aws.String(rsApiKey.Primary.ID) {
				key = *v
			}
		}
		if key.Id == nil {
			return fmt.Errorf("Key Not found: %s  %s", ApiKeyName, *aws.String(rsApiKey.Primary.ID))
		}
		// aws when they create, slight difference will be in the minutes, so better check date
		if time.Unix(*key.Expires, 0).Format("02/01/2006") != time.Unix(date, 0).Format("02/01/2006") {

			return fmt.Errorf("Expiry date got is: %s and expected is %s", time.Unix(*key.Expires, 0).Format("02/01/2006"),
				time.Unix(date, 0).Format("02/01/2006"))
		}

		return nil
	}
}

func testAccAppsyncApiKeyConfigValidTillDate(rDate string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test1" {
  authentication_type = "API_KEY"
  name = "tf_appsync_test1"
}
resource "aws_appsync_api_key" "test_valid_till_date" {
	appsync_api_id = "${aws_appsync_graphql_api.test1.id}"
	valid_till_date = "%s"
}

`, rDate)
}

func testAccAppsyncApiKeyConfigValidityPeriodDays(rDays string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test2" {
  authentication_type = "API_KEY"
  name = "tf_appsync_test2"
}

resource "aws_appsync_api_key" "test_validity_period_days" {
	appsync_api_id = "${aws_appsync_graphql_api.test2.id}"
	validity_period_days = %s
}

`, rDays)
}
