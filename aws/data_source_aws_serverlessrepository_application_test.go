package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func TestAccDataSourceAwsServerlessRepositoryApplication_Basic(t *testing.T) {
	datasourceName := "data.aws_serverlessrepository_application.secrets_manager_postgres_single_user_rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServerlessRepositoryApplicationDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServerlessRepositoryApplicationDataSourceID(datasourceName),
					resource.TestCheckResourceAttr(datasourceName, "name", "SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttrSet(datasourceName, "semantic_version"),
					resource.TestCheckResourceAttrSet(datasourceName, "source_code_url"),
					resource.TestCheckResourceAttrSet(datasourceName, "template_url"),
					resource.TestCheckResourceAttrSet(datasourceName, "required_capabilities.#"),
				),
			},
			{
				Config:      testAccCheckAwsServerlessRepositoryApplicationDataSourceConfig_NonExistent,
				ExpectError: regexp.MustCompile(`error reading Serverless Application Repository application`),
			},
		},
	})
}
func TestAccDataSourceAwsServerlessRepositoryApplication_Versioned(t *testing.T) {
	datasourceName := "data.aws_serverlessrepository_application.secrets_manager_postgres_single_user_rotator"

	const (
		version1 = "1.0.15"
		version2 = "1.1.78"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServerlessRepositoryApplicationDataSourceConfig_Versioned(version1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServerlessRepositoryApplicationDataSourceID(datasourceName),
					resource.TestCheckResourceAttr(datasourceName, "name", "SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr(datasourceName, "semantic_version", version1),
					resource.TestCheckResourceAttrSet(datasourceName, "source_code_url"),
					resource.TestCheckResourceAttrSet(datasourceName, "template_url"),
					resource.TestCheckResourceAttr(datasourceName, "required_capabilities.#", "0"),
				),
			},
			{
				Config: testAccCheckAwsServerlessRepositoryApplicationDataSourceConfig_Versioned(version2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServerlessRepositoryApplicationDataSourceID(datasourceName),
					resource.TestCheckResourceAttr(datasourceName, "name", "SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr(datasourceName, "semantic_version", version2),
					resource.TestCheckResourceAttrSet(datasourceName, "source_code_url"),
					resource.TestCheckResourceAttrSet(datasourceName, "template_url"),
					resource.TestCheckResourceAttr(datasourceName, "required_capabilities.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttr(datasourceName, "required_capabilities.*", "CAPABILITY_IAM"),
					tfawsresource.TestCheckTypeSetElemAttr(datasourceName, "required_capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
				),
			},
			{
				Config:      testAccCheckAwsServerlessRepositoryApplicationDataSourceConfig_Versioned_NonExistent,
				ExpectError: regexp.MustCompile(`error reading Serverless Application Repository application`),
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

const testAccCheckAwsServerlessRepositoryApplicationDataSourceConfig_NonExistent = `
data "aws_serverlessrepository_application" "no_such_function" {
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/ThisFunctionDoesNotExist"
}
`

func testAccCheckAwsServerlessRepositoryApplicationDataSourceConfig_Versioned(version string) string {
	return fmt.Sprintf(`
data "aws_serverlessrepository_application" "secrets_manager_postgres_single_user_rotator" {
  application_id   = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  semantic_version = "%[1]s"
}
`, version)
}

const testAccCheckAwsServerlessRepositoryApplicationDataSourceConfig_Versioned_NonExistent = `
data "aws_serverlessrepository_application" "secrets_manager_postgres_single_user_rotator" {
  application_id   = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  semantic_version = "42.13.7"
}
`
