// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecretVersionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resource1Name := "aws_secretsmanager_secret_version.test"
	dataSourceName := "data.aws_secretsmanager_secret_versions.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "secret_id", resource1Name, "secret_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "versions.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "versions.0.%"),
					resource.TestCheckResourceAttrSet(dataSourceName, "versions.0.created_time"),
					resource.TestCheckResourceAttrSet(dataSourceName, "versions.0.version_id"),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretVersionsDataSource_emptyVer(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_secretsmanager_secret_versions.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionsDataSourceConfig_emptyVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "versions.#", "0"),
				),
			},
		},
	})
}

func testAccSecretVersionsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"
}

resource "aws_secretsmanager_secret_version" "test2" {
  depends_on    = [aws_secretsmanager_secret_version.test]
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string2"
}

data "aws_secretsmanager_secret_versions" "test" {
  depends_on = [aws_secretsmanager_secret_version.test]
  secret_id  = aws_secretsmanager_secret.test.id
}
`, rName)
}

func testAccSecretVersionsDataSourceConfig_emptyVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}
data "aws_secretsmanager_secret_versions" "test" {
  secret_id = aws_secretsmanager_secret.test.id
}
`, rName)
}
