// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerRandomPasswordEphemeral_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck: acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(ctx, acctest.ProviderNameEcho),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccRandomPasswordEphemeralResourceConfig_basic(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("random_password"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccRandomPasswordEphemeralResourceConfig_basic() string {
	return acctest.ConfigCompose(
		acctest.ConfigWithEchoProvider("ephemeral.aws_secretsmanager_random_password.test"),
		`
ephemeral "aws_secretsmanager_random_password" "test" {
  password_length     = 40
  exclude_characters  = "abc"
  exclude_numbers     = true
  exclude_uppercase   = true
  exclude_punctuation = true
}
`)
}
