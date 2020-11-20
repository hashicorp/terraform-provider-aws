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
		version1 = "1.0.13"
		version2 = "1.1.36"
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

const testAccCheckAwsServerlessRepositoryApplicationDataSourceConfig = testAccCheckAwsServerlessRepositoryPostgresSingleUserRotatorApplication + `
data "aws_serverlessrepository_application" "secrets_manager_postgres_single_user_rotator" {
  application_id = local.postgres_single_user_rotator_arn
}
`

const testAccCheckAwsServerlessRepositoryApplicationDataSourceConfig_NonExistent = `
data "aws_serverlessrepository_application" "no_such_function" {
  application_id = "arn:${data.aws_partition.current.partition}:serverlessrepo:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:applications/ThisFunctionDoesNotExist"
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}
`

func testAccCheckAwsServerlessRepositoryApplicationDataSourceConfig_Versioned(version string) string {
	return composeConfig(
		testAccCheckAwsServerlessRepositoryPostgresSingleUserRotatorApplication,
		fmt.Sprintf(`
data "aws_serverlessrepository_application" "secrets_manager_postgres_single_user_rotator" {
  application_id   = local.postgres_single_user_rotator_arn
  semantic_version = "%[1]s"
}
`, version))
}

const testAccCheckAwsServerlessRepositoryApplicationDataSourceConfig_Versioned_NonExistent = testAccCheckAwsServerlessRepositoryPostgresSingleUserRotatorApplication + `
data "aws_serverlessrepository_application" "secrets_manager_postgres_single_user_rotator" {
  application_id   = local.postgres_single_user_rotator_arn
  semantic_version = "42.13.7"
}
`
