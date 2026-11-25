// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfmeta "github.com/hashicorp/terraform-provider-aws/internal/service/meta"
)

func TestAccCredentialsEphemeral_basic(t *testing.T) {
	ctx := acctest.Context(t)
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(ctx, t) },
		ErrorCheck: acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(ctx, acctest.ProviderNameEcho),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCredentialsEphemeralConfig_basic(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key_id"), knownvalue.StringRegexp(regexache.MustCompile(`^A[KS]IA[0-7A-Z]{16}$`))),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("session_token"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func TestAccCredentialsEphemeral_region(t *testing.T) {
	ctx := acctest.Context(t)
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(ctx, t) },
		ErrorCheck: acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(ctx, acctest.ProviderNameEcho),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCredentialsEphemeralConfig_region(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("access_key_id"), knownvalue.StringRegexp(regexache.MustCompile(`^A[KS]IA[0-7A-Z]{16}$`))),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("secret_access_key"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("session_token"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccCredentialsEphemeralConfig_basic() string {
	return acctest.ConfigCompose(
		acctest.ConfigWithEchoProvider("ephemeral.aws_credentials.test"),
		`
ephemeral "aws_credentials" "test" {}
`)
}

func testAccCredentialsEphemeralConfig_region() string {
	return acctest.ConfigCompose(
		acctest.ConfigWithEchoProvider("ephemeral.aws_credentials.test"),
		fmt.Sprintf(`
ephemeral "aws_credentials" "test" {
  region = %[1]q
}
`, acctest.Region()))
}
