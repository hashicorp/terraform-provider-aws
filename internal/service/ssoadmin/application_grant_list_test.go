// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

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
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminApplicationGrant_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	authCodeResourceName := "aws_ssoadmin_application_grant.auth_code"
	refreshTokenResourceName := "aws_ssoadmin_application_grant.refresh_token"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	authCodeIdentity := tfstatecheck.Identity()
	refreshTokenIdentity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		CheckDestroy:             testAccCheckApplicationGrantDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationGrant/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					authCodeIdentity.GetIdentity(authCodeResourceName),
					refreshTokenIdentity.GetIdentity(refreshTokenResourceName),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationGrant/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_application_grant.test", authCodeIdentity.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_ssoadmin_application_grant.test", tfqueryfilter.ByResourceIdentityFunc(authCodeIdentity.Checks())),
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_application_grant.test", refreshTokenIdentity.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_ssoadmin_application_grant.test", tfqueryfilter.ByResourceIdentityFunc(refreshTokenIdentity.Checks())),
				},
			},
		},
	})
}

func TestAccSSOAdminApplicationGrant_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_ssoadmin_application_grant.auth_code"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity := tfstatecheck.Identity()
	grantID := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		CheckDestroy:             testAccCheckApplicationGrantDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationGrant/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					grantID.GetStateValue(resourceName, tfjsonpath.New(names.AttrID)),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationGrant/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_application_grant.test", identity.Checks()),
					querycheck.ExpectResourceKnownValues("aws_ssoadmin_application_grant.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), grantID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("grant_type"), knownvalue.StringExact("authorization_code")),
					}),
				},
			},
		},
	})
}

func TestAccSSOAdminApplicationGrant_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	authCodeResourceName := "aws_ssoadmin_application_grant.auth_code"
	refreshTokenResourceName := "aws_ssoadmin_application_grant.refresh_token"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	authCodeIdentity := tfstatecheck.Identity()
	refreshTokenIdentity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationGrant/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					authCodeIdentity.GetIdentity(authCodeResourceName),
					refreshTokenIdentity.GetIdentity(refreshTokenResourceName),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationGrant/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_application_grant.test", authCodeIdentity.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_application_grant.test", refreshTokenIdentity.Checks()),
				},
			},
		},
	})
}
