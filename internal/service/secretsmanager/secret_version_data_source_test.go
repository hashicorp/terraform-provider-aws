// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecretVersionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	datasourceName := "data.aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSecretVersionDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`couldn't find resource`),
			},
			{
				Config: testAccSecretVersionDataSourceConfig_stageDefault(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccSecretVersionCheckDataSource(datasourceName, resourceName),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretVersionDataSource_versionID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	datasourceName := "data.aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionDataSourceConfig_id(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccSecretVersionCheckDataSource(datasourceName, resourceName),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretVersionDataSource_versionStage(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	datasourceName := "data.aws_secretsmanager_secret_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionDataSourceConfig_stageCustom(rName),
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

const testAccSecretVersionDataSourceConfig_nonExistent = `
data "aws_secretsmanager_secret_version" "test" {
  secret_id = "tf-acc-test-does-not-exist"
}
`

func testAccSecretVersionDataSourceConfig_id(rName string) string {
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

func testAccSecretVersionDataSourceConfig_stageCustom(rName string) string {
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

func testAccSecretVersionDataSourceConfig_stageDefault(rName string) string {
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
