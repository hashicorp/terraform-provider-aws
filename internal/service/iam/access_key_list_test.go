// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMAccessKey_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_iam_access_key.test[0]"
	resourceName2 := "aws_iam_access_key.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()
	keyID1 := tfstatecheck.StateValue()
	keyID2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		CheckDestroy:             testAccCheckAccessKeyDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessKey/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					keyID1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("user"), knownvalue.StringExact(rName)),

					identity2.GetIdentity(resourceName2),
					keyID2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("user"), knownvalue.StringExact(rName)),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/AccessKey/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_iam_access_key.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_iam_access_key.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), newAccessKeyDisplayNameCheck(rName, &keyID1)),
					tfquerycheck.ExpectNoResourceObject("aws_iam_access_key.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_iam_access_key.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_iam_access_key.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), newAccessKeyDisplayNameCheck(rName, &keyID2)),
					tfquerycheck.ExpectNoResourceObject("aws_iam_access_key.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccIAMAccessKey_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_iam_access_key.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	keyID1 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		CheckDestroy:             testAccCheckAccessKeyDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccessKey/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					keyID1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("user"), knownvalue.StringExact(rName)),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/AccessKey/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_iam_access_key.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_iam_access_key.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), newAccessKeyDisplayNameCheck(rName, &keyID1)),
					querycheck.ExpectResourceKnownValues("aws_iam_access_key.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("create_date"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("Inactive")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("user"), knownvalue.StringExact(rName)),
					}),
				},
			},
		},
	})
}

type accessKeyDisplayNameCheck struct {
	username string
	keyID    interface{ Value() string }
}

func newAccessKeyDisplayNameCheck(username string, keyID interface{ Value() string }) knownvalue.Check {
	return accessKeyDisplayNameCheck{
		username: username,
		keyID:    keyID,
	}
}

func (c accessKeyDisplayNameCheck) CheckValue(other any) error {
	actual, ok := other.(string)
	if !ok {
		return fmt.Errorf("expected string display name, got: %T", other)
	}

	expected := fmt.Sprintf("User: %s - Access Key: %s", c.username, c.keyID.Value())
	if actual != expected {
		return fmt.Errorf("expected display name %q, got %q", expected, actual)
	}

	return nil
}

func (c accessKeyDisplayNameCheck) String() string {
	return fmt.Sprintf("User: %s - Access Key: <state id>", c.username)
}
