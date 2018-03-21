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
	dateAfterOneYear := time.Now().Add(time.Hour * 24 * time.Duration(364))
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAppsyncApiKeyConfigBasic(dateAfterOneYear.Format(time.RFC3339)),
				ResourceName: "aws_appsync_api_key.test",
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncApiKeyExistsTillDate(
						"aws_appsync_api_key.test",
						dateAfterOneYear.Unix(),
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
		ApiId, _, er := decodeAppSyncApiKeyId(rs.Primary.ID)
		if er != nil {
			return er
		}
		describe, err := conn.ListApiKeys(&appsync.ListApiKeysInput{ApiId: aws.String(ApiId)})
		if err == nil {
			if len(describe.ApiKeys) != 0 {
				return fmt.Errorf("Appsync ApiKey still exists")
			}
		}

		return nil

	}
	return nil
}

func testAccCheckAwsAppsyncApiKeyExistsTillDate(ApiKeyName string, date int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rsApiKey, ok := s.RootModule().Resources[ApiKeyName]
		if !ok {
			return fmt.Errorf("Key Not found in state: %s", ApiKeyName)
		}
		ApiId, Id, er := decodeAppSyncApiKeyId(rsApiKey.Primary.ID)
		if er != nil {
			return er
		}

		conn := testAccProvider.Meta().(*AWSClient).appsyncconn
		input := &appsync.ListApiKeysInput{
			ApiId: aws.String(ApiId),
		}

		resp, err := conn.ListApiKeys(input)
		if err != nil {
			return err
		}

		var key appsync.ApiKey
		for _, v := range resp.ApiKeys {
			if *v.Id == *aws.String(Id) {
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

func testAccAppsyncApiKeyConfigBasic(rDate string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name = "tf_appsync_test"
}
resource "aws_appsync_api_key" "test" {
	api_id = "${aws_appsync_graphql_api.test.id}"
	expires = "%sT00:00:00Z"
}

`, strings.Split(rDate, "T")[0])
}
