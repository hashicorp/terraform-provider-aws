// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecretVersion_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_secretsmanager_secret_version.test[0]"
	resourceName2 := "aws_secretsmanager_secret_version.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	secretID := tfstatecheck.StateValue()
	versionID1 := tfstatecheck.StateValue()
	versionID2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/SecretVersion/list_basic"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					secretID.GetStateValue("aws_secretsmanager_secret.test", tfjsonpath.New(names.AttrID)),
					versionID1.GetStateValue(resourceName1, tfjsonpath.New("version_id")),
					versionID2.GetStateValue(resourceName2, tfjsonpath.New("version_id")),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/SecretVersion/list_basic"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_secretsmanager_secret_version.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						"secret_id":         secretID.Value(),
						"version_id":        versionID1.Value(),
					}),
					querycheck.ExpectIdentity("aws_secretsmanager_secret_version.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						"secret_id":         secretID.Value(),
						"version_id":        versionID2.Value(),
					}),
				},
			},
		},
	})
}
