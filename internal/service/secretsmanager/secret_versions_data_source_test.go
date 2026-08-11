// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecretVersionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resource1Name := "aws_secretsmanager_secret_version.test"
	resource2Name := "aws_secretsmanager_secret_version.test2"
	secretName := "aws_secretsmanager_secret.test"
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
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrARN), dataSourceName, tfjsonpath.New("secret_arn"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("include_deprecated"), knownvalue.Null()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrName), dataSourceName, tfjsonpath.New("secret_name"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("secret_arn"), resource1Name, tfjsonpath.New("secret_arn"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("secret_id"), resource1Name, tfjsonpath.New("secret_id"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("secret_name"), secretName, tfjsonpath.New(names.AttrName), compare.ValuesSame()),
					// Order is not guaranteed for `versions`
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("versions"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrCreatedTime: knownvalue.StringRegexp(regexache.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
							"last_accessed_date":  knownvalue.StringRegexp(regexache.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
							"version_id":          knownvalue.NotNull(),
							"version_stages":      knownvalue.ListSizeExact(1),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrCreatedTime: knownvalue.StringRegexp(regexache.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
							"last_accessed_date":  knownvalue.StringRegexp(regexache.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)),
							"version_id":          knownvalue.NotNull(),
							"version_stages":      knownvalue.ListSizeExact(1),
						}),
					})),
					statecheck.CompareValueCollection(dataSourceName, []tfjsonpath.Path{tfjsonpath.New("versions"), tfjsonpath.New("version_id")}, resource1Name, tfjsonpath.New("version_id"), compare.ValuesSame()),
					statecheck.CompareValueCollection(dataSourceName, []tfjsonpath.Path{tfjsonpath.New("versions"), tfjsonpath.New("version_id")}, resource2Name, tfjsonpath.New("version_id"), compare.ValuesSame()),
				},
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
data "aws_secretsmanager_secret_versions" "test" {
  secret_id = aws_secretsmanager_secret.test.id

  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_secretsmanager_secret_version.test2,
  ]
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"
}

resource "aws_secretsmanager_secret_version" "test2" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string2"

  depends_on = [aws_secretsmanager_secret_version.test]
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
