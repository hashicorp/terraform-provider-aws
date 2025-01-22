// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKMSSecretsEphemeral_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")
	plaintext := "my-plaintext-string"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(ctx, t) },
		ErrorCheck: acctest.ErrorCheck(t, names.KMSServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(ctx, acctest.ProviderNameEcho),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretsEphemeralResourceConfig_basic(rName, plaintext),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("plaintext").AtMapKey(rName), knownvalue.StringExact(plaintext)),
				},
			},
		},
	})
}

func testAccSecretsEphemeralResourceConfig_basic(rName, secretString string) string {
	return acctest.ConfigCompose(
		acctest.ConfigWithEchoProvider("ephemeral.aws_kms_secrets.test"),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
  is_enabled  = true
}

resource "aws_kms_ciphertext" "test" {
  key_id = aws_kms_key.test.key_id

  context = {
    foo = "bar"
  }

  plaintext = %[2]q
}

ephemeral "aws_kms_secrets" "test" {
  secret {
    name    = %[1]q
    payload = aws_kms_ciphertext.test.ciphertext_blob
    context = aws_kms_ciphertext.test.context
  }

  depends_on = [aws_kms_key.test]
}
`, rName, secretString))
}
