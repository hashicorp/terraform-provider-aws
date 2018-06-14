package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsAppsyncSchema(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncSchemaConfig(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncSchemaExists("aws_appsync_schema.test"),
				),
			},
		},
	})
}

func TestAccAwsAppsyncSchema_update(t *testing.T) {
	rName := acctest.RandString(5)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncSchemaConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncSchemaExists("aws_appsync_schema.test"),
				),
			},
			{
				Config: testAccAppsyncSchemaConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncSchemaExists("aws_appsync_schema.test"),
				),
			},
		},
	})
}

func testAccCheckAwsAppsyncSchemaDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appsyncconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appsync_schema" {
			continue
		}

		input := &appsync.GetSchemaCreationStatusInput{
			ApiId: aws.String(rs.Primary.Attributes["api_id"]),
		}

		_, err := conn.GetSchemaCreationStatus(input)
		if err != nil {
			if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckAwsAppsyncSchemaExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAppsyncSchemaConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name = "tf_appsync_%s"
}

resource "aws_appsync_schema" "test" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  definition = <<EOF
schema {
	query: Query
}
type Query {
  test: Int
}
EOF
}
`, rName)
}

func testAccAppsyncSchemaConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name = "tf_appsync_%s"
}

resource "aws_appsync_schema" "test" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  definition = <<EOF
schema {
    query:Query
}

type Query {
    getTodos: [String]
}
EOF
}
`, rName)
}
