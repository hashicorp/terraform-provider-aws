package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSEcrDataSource_LoginTokenBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcrLoginTokenDataSourceBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccAWSEcrDataSourc_ecrLoginTokenMeta("data.aws_ecr_login_token.default"),
				),
			},
		},
	})
}

func TestAccAWSEcrDataSource_LoginTokenMultipleRepos(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcrLoginTokenDataSourceMultipleConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccAWSEcrDataSourc_ecrLoginTokenMultipeMeta("data.aws_ecr_login_token.onefrommany", "1"),
				),
			},
		},
	})
}

func testAccAWSEcrDataSourc_ecrLoginTokenMeta(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No id set in data source")
		}

		if _, ok := rs.Primary.Attributes["token.0"]; !ok {
			return fmt.Errorf("token.0 is missing, should be set.")
		}
		if _, ok := rs.Primary.Attributes["proxy_endpoint.0"]; !ok {
			return fmt.Errorf("proxy_endpoint.0 is missing, should be set.")
		}

		if _, ok := rs.Primary.Attributes["expires_at.0"]; !ok {
			return fmt.Errorf("expires_at.0 is missing, should be set.")
		}

		if rs.Primary.Attributes["token.#"] != "1" {
			return fmt.Errorf("token.# attribute should have 1 element, has: %s", rs.Primary.Attributes["token.#"])
		}

		if rs.Primary.Attributes["proxy_endpoint.#"] != "1" {
			return fmt.Errorf("proxy_endpoint.# attribute should have 1 element, has: %s", rs.Primary.Attributes["proxy_endpoint.#"])
		}

		if rs.Primary.Attributes["expires_at.#"] != "1" {
			return fmt.Errorf("expires_at.# attribute should have 1 element, has: %s", rs.Primary.Attributes["expires_at.#"])
		}
		return nil
	}
}

func testAccAWSEcrDataSourc_ecrLoginTokenMultipeMeta(n string, c string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No id set in data source")
		}

		if _, ok := rs.Primary.Attributes["token.#"]; !ok {
			return fmt.Errorf("token.# is missing, should be set.")
		}

		if rs.Primary.Attributes["token.#"] != c {
			return fmt.Errorf("token.# attribute should have %s element, has: %s", c, rs.Primary.Attributes["token.#"])
		}
		return nil
	}
}

var testAccCheckAwsEcrLoginTokenDataSourceBasicConfig = `
resource "aws_ecr_repository" "repo" {
  name = "testrepobasic"
}

data "aws_ecr_login_token" "default" { }
`

var testAccCheckAwsEcrLoginTokenDataSourceMultipleConfig = `
resource "aws_ecr_repository" "repo01" {
  name = "testrepobasic01"
}

resource "aws_ecr_repository" "repo02" {
  name = "testrepobasic02"
}

data "aws_ecr_login_token" "onefrommany" {
	registry_ids = [
		"${aws_ecr_repository.repo02.registry_id}"
	]
}
`
