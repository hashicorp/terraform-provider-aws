package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAthenaNamedQuery(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaNamedQueryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaNamedQueryConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaNamedQueryExists("aws_athena_named_query.foo"),
				),
			},
		},
	})
}

func testAccCheckAWSAthenaNamedQueryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).athenaconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_athena_named_query" {
			continue
		}

		input := &athena.GetNamedQueryInput{
			NamedQueryId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetNamedQuery(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case athena.ErrCodeInvalidRequestException:
					return nil
				default:
					return err
				}
			}
			return err
		}
	}
	return nil
}

func testAccCheckAWSAthenaNamedQueryExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

const testAccAthenaNamedQueryConfig = `
resource "aws_athena_named_query" "foo" {
	name = "bar"
	database = "users"
	query = "SELECT * FROM users limit 10;"
	description = "Fetch latest users"
}
`
