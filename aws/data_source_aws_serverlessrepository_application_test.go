package aws

import (
	"testing"

	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsServerlessRepositoryApplication_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServerlessRepositoryApplicationDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServerlessRepositoryApplicationDataSourceID("data.aws_serverlessrepository_application.secrets_manager_postgres_single_user_rotator"),
					resource.TestCheckResourceAttr("data.aws_serverlessrepository_application.secrets_manager_postgres_single_user_rotator", "name", "SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttrSet("data.aws_serverlessrepository_application.secrets_manager_postgres_single_user_rotator", "semantic_version"),
					resource.TestCheckResourceAttrSet("data.aws_serverlessrepository_application.secrets_manager_postgres_single_user_rotator", "source_code_url"),
					resource.TestCheckResourceAttrSet("data.aws_serverlessrepository_application.secrets_manager_postgres_single_user_rotator", "template_url"),
				),
			},
		},
	})
}

func testAccCheckAwsServerlessRepositoryApplicationDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Serverless Repository Application data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("AMI data source ID not set")
		}
		return nil
	}
}

const testAccCheckAwsServerlessRepositoryApplicationDataSourceConfig = `
data "aws_serverlessrepository_application" "secrets_manager_postgres_single_user_rotator" {
	application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
}
`
