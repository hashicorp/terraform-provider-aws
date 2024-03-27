// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecretVersionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_secretsmanager_secret_version.test1"
	resource2Name := "aws_secretsmanager_secret_version.test2"
	datasourceName := "data.aws_secretsmanager_secret_versions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSecretVersionsDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`couldn't find resource`),
			},
			{
				Config: testAccSecretVersionsDataSourceConfig_oneVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccSecretVersionsCheckDataSource(datasourceName, resource1Name, "", 1),
				),
			},
			{
				Config: testAccSecretVersionsDataSourceConfig_twoVersions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccSecretVersionsCheckDataSource(datasourceName, resource1Name, resource2Name, 2),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretVersionsDataSource_maxResults(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_secretsmanager_secret_version.test1"
	datasourceName := "data.aws_secretsmanager_secret_versions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionsDataSourceConfig_maxResults(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccSecretVersionsCheckDataSource(datasourceName, resource1Name, "", 1),
				),
			},
		},
	})
}

func testAccSecretVersionsCheckDataSource(datasourceName, resource1Name, resource2Name string, versionCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dataSource, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no data source called %s", datasourceName)
		}

		resource1, ok := s.RootModule().Resources[resource1Name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resource1Name)
		}

		resource2, ok := s.RootModule().Resources[resource2Name]
		if resource2Name != "" && !ok {
			return fmt.Errorf("root module has no resource called %s", resource2Name)
		}

		if dataSource.Primary.Attributes["versions.#"] != strconv.Itoa(versionCount) {
			return fmt.Errorf(
				"versions.# is %s; want %d",
				dataSource.Primary.Attributes["versions.#"],
				versionCount,
			)
		}

		checkAttributes := func(attrNames []string, resourceAttributes map[string]string) error {
			for _, attrName := range attrNames {
				if resourceAttributes[attrName] != dataSource.Primary.Attributes[attrName] {
					return fmt.Errorf(
						"%s is %s; want %s",
						attrName,
						resourceAttributes[attrName],
						dataSource.Primary.Attributes[attrName],
					)
				}
			}
			return nil
		}

		attrNamesResource1 := []string{
			"arn",
			"secret_id",
			"versions.0,created_date",
			"versions.0,version_id",
			"versions.0,version_stages.#",
			"versions.0,last_accessed_date",
		}
		err := checkAttributes(attrNamesResource1, resource1.Primary.Attributes)
		if err != nil {
			return err
		}

		if resource2Name != "" {
			attrNamesResource2 := []string{
				"arn",
				"secret_id",
				"versions.1,created_date",
				"versions.1,version_id",
				"versions.1,version_stages.#",
				"versions.1,last_accessed_date",
			}
			err = checkAttributes(attrNamesResource2, resource2.Primary.Attributes)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

const testAccSecretVersionsDataSourceConfig_nonExistent = `
data "aws_secretsmanager_secret_versions" "test" {
  secret_id = "tf-acc-test-does-not-exist"
}
`

func testAccSecretVersionsDataSourceConfig_oneVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

resource "aws_secretsmanager_secret_version" "test1" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"
}
  
data "aws_secretsmanager_secret_versions" "test" {
  depends_on = [aws_secretsmanager_secret_version.test1]
  secret_id = aws_secretsmanager_secret.test.id
}
`, rName)
}

func testAccSecretVersionsDataSourceConfig_twoVersions(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

resource "aws_secretsmanager_secret_version" "test1" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"
}

resource "aws_secretsmanager_secret_version" "test2" {
  depends_on = [aws_secretsmanager_secret_version.test1]
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string2"
}

data "aws_secretsmanager_secret_versions" "test" {
  depends_on = [aws_secretsmanager_secret_version.test1, aws_secretsmanager_secret_version.test2]
  secret_id = aws_secretsmanager_secret.test.id
}
`, rName)
}

func testAccSecretVersionsDataSourceConfig_maxResults(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

resource "aws_secretsmanager_secret_version" "test1" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"
}

resource "aws_secretsmanager_secret_version" "test2" {
  depends_on = [aws_secretsmanager_secret_version.test1]
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string2"
}

data "aws_secretsmanager_secret_versions" "test" {
  depends_on = [aws_secretsmanager_secret_version.test1, aws_secretsmanager_secret_version.test2]
  secret_id = aws_secretsmanager_secret.test.id
  max_results = 1
}
`, rName)
}
