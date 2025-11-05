// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr_test

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

func TestAccECRAuthorizationTokenEphemeral_basic(t *testing.T) {
	ctx := acctest.Context(t)
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(ctx, t) },
		ErrorCheck: acctest.ErrorCheck(t, names.ECRServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(ctx, acctest.ProviderNameEcho),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizationTokenEphemeralResourceConfig_basic(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("authorization_token"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("proxy_endpoint"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("expires_at"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey(names.AttrUserName), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey(names.AttrPassword), knownvalue.NotNull()),
				},
			},
		},
	})
}

func TestAccECRAuthorizationTokenEphemeral_registryID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(ctx, t) },
		ErrorCheck: acctest.ErrorCheck(t, names.ECRServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(ctx, acctest.ProviderNameEcho),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizationTokenEphemeralResourceConfig_registryID(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("authorization_token"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("proxy_endpoint"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("expires_at"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey(names.AttrUserName), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey(names.AttrPassword), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccAuthorizationTokenEphemeralResourceConfig_basic() string {
	return acctest.ConfigCompose(
		acctest.ConfigWithEchoProvider("ephemeral.aws_ecr_authorization_token.test"),
		`
ephemeral "aws_ecr_authorization_token" "test" {}
`)
}

func testAccAuthorizationTokenEphemeralResourceConfig_registryID(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigWithEchoProvider("ephemeral.aws_ecr_authorization_token.test"),
		fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q
}

ephemeral "aws_ecr_authorization_token" "test" {
  registry_id = aws_ecr_repository.test.registry_id
}
`, rName))
}
