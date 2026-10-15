// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sts_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSTSWebIdentityTokenEphemeral_basic(t *testing.T) {
	ctx := acctest.Context(t)
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")
	domain := acctest.RandomDomain(t).String()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccWebIdentityTokenPreCheck(ctx, t)
		},
		ErrorCheck: acctest.ErrorCheck(t, names.STSServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(ctx, acctest.ProviderNameEcho),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccWebIdentityTokenEphemeralConfig_basic(domain),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("web_identity_token"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("expiration"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func TestAccSTSWebIdentityTokenEphemeral_full(t *testing.T) {
	ctx := acctest.Context(t)
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")
	domain1, domain2 := acctest.RandomDomain(t).String(), acctest.RandomDomain(t).String()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccWebIdentityTokenPreCheck(ctx, t)
		},
		ErrorCheck: acctest.ErrorCheck(t, names.STSServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(ctx, acctest.ProviderNameEcho),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccWebIdentityTokenEphemeralConfig_full(domain1, domain2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("web_identity_token"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("expiration"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccWebIdentityTokenPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).STSClient(ctx)

	input := sts.GetWebIdentityTokenInput{
		Audience:         []string{"test"},
		SigningAlgorithm: aws.String("RS256"),
	}
	_, err := conn.GetWebIdentityToken(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance test: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccWebIdentityTokenEphemeralConfig_basic(domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigWithEchoProvider("ephemeral.aws_sts_web_identity_token.test"),
		fmt.Sprintf(`
ephemeral "aws_sts_web_identity_token" "test" {
  audience          = [%[1]q]
  signing_algorithm = "RS256"
}
`, domain))
}

func testAccWebIdentityTokenEphemeralConfig_full(domain1, domain2 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigWithEchoProvider("ephemeral.aws_sts_web_identity_token.test"),
		fmt.Sprintf(`
ephemeral "aws_sts_web_identity_token" "test" {
  audience          = [%[1]q, %[2]q]
  signing_algorithm = "ES384"
  duration_seconds  = 600

  tags = {
    environment = "test"
    purpose     = "acceptance-testing"
  }
}
`, domain1, domain2))
}
