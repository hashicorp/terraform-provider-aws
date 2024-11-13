// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")
	secretString := "super-secret"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck: acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"echo": echoprovider.NewProviderServer(),
		},
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
	return fmt.Sprintf(`
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

provider "echo" {
  data = ephemeral.aws_secretsmanager_secret_version.test
}

resource "echo" "test" {}
`, rName, secretString)
}
