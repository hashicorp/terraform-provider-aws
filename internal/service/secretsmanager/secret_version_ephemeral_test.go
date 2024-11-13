// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecretVersionEphemeral_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_version.test"
	datasourceName := "ephemeral.aws_secretsmanager_secret_version.test"

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

func testAccSecretVersionEphemeralResourceConfig_basic(rName string) string {
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
