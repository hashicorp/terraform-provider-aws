// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidentity_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-uuid"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIdentityOpenIDTokenForDeveloperIdentityEphemeral_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	uuid, err := uuid.GenerateUUID()
	developerProviderName := sdkacctest.RandString(10)
	echoResourceName := "echo.test"
	dataPath := tfjsonpath.New("data")
	if err != nil {
		t.Logf("error generating uuid: %s", err.Error())
		t.Fail()
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CognitoIdentityEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck: acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(ctx, acctest.ProviderNameEcho),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDTokenForDeveloperIdentityEphemeralConfig_basic(rName, developerProviderName, uuid),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dataPath.AtMapKey("token"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccOpenIDTokenForDeveloperIdentityEphemeralConfig_basic(rName, developerProviderName, uuid string) string {
	return acctest.ConfigCompose(
		acctest.ConfigWithEchoProvider("ephemeral.aws_cognito_identity_openid_token_for_developer_identity.test"),
		testAccPoolConfig_developerProviderName(rName, developerProviderName),
		fmt.Sprintf(`
data "aws_region" "current" {}

ephemeral "aws_cognito_identity_openid_token_for_developer_identity" "test" {
  identity_pool_id = aws_cognito_identity_pool.test.id

  logins = {
    %[2]q = "user123"
  }
}
`, rName, developerProviderName, uuid))
}
