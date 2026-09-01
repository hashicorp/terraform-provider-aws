// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecretVersionEphemeral_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")
	secretString := "super-secret"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck: acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(ctx, acctest.ProviderNameEcho),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretVersionEphemeralResourceConfig_basic(rName, secretString),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey(names.AttrCreatedDate), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_string"), knownvalue.StringExact(secretString)),
				},
			},
		},
	})
}

func testAccSecretVersionEphemeralResourceConfig_basic(rName, secretString string) string {
	return acctest.ConfigCompose(
		acctest.ConfigWithEchoProvider("ephemeral.aws_secretsmanager_secret_version.test"),
		fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = %[2]q
}

ephemeral "aws_secretsmanager_secret_version" "test" {
  secret_id  = aws_secretsmanager_secret.test.id
  version_id = aws_secretsmanager_secret_version.test.version_id
}
`, rName, secretString))
}
