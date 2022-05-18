package secretsmanager_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSecretsManagerSecretVersionDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	datasourceName := "data.aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSecretVersionDataSourceConfig_NonExistent,
				ExpectError: regexp.MustCompile(`not found`),
			},
			{
				Config: testAccSecretVersionDataSourceConfig_VersionStage_Default(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccSecretVersionCheckDataSource(datasourceName, resourceName),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretVersionDataSource_versionID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	datasourceName := "data.aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionDataSourceConfig_VersionID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccSecretVersionCheckDataSource(datasourceName, resourceName),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretVersionDataSource_versionStage(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	datasourceName := "data.aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionDataSourceConfig_VersionStage_Custom(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccSecretVersionCheckDataSource(datasourceName, resourceName),
				),
			},
		},
	})
}

func testAccSecretVersionCheckDataSource(datasourceName, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		dataSource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		attrNames := []string{
			"secret_value",
			"version_stages.#",
		}

		for _, attrName := range attrNames {
			if resource.Primary.Attributes[attrName] != dataSource.Primary.Attributes[attrName] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrName,
					resource.Primary.Attributes[attrName],
					dataSource.Primary.Attributes[attrName],
				)
			}
		}

		return nil
	}
}

const testAccSecretVersionDataSourceConfig_NonExistent = `
data "aws_secretsmanager_secret_version" "test" {
  secret_id = "tf-acc-test-does-not-exist"
}
`

func testAccSecretVersionDataSourceConfig_VersionID(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"
}

data "aws_secretsmanager_secret_version" "test" {
  secret_id  = aws_secretsmanager_secret.test.id
  version_id = aws_secretsmanager_secret_version.test.version_id
}
`, rName)
}

func testAccSecretVersionDataSourceConfig_VersionStage_Custom(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id      = aws_secretsmanager_secret.test.id
  secret_string  = "test-string"
  version_stages = ["test-stage", "AWSCURRENT"]
}

data "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret_version.test.secret_id
  version_stage = "test-stage"
}
`, rName)
}

func testAccSecretVersionDataSourceConfig_VersionStage_Default(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"
}

data "aws_secretsmanager_secret_version" "test" {
  secret_id = aws_secretsmanager_secret_version.test.secret_id
}
`, rName)
}
