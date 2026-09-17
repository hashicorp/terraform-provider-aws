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
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminAccountAssignment_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceNameGroup := "aws_ssoadmin_account_assignment.group"
	resourceNameUser := "aws_ssoadmin_account_assignment.user"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	groupName := acctest.SkipIfEnvVarNotSet(t, "AWS_IDENTITY_STORE_GROUP_NAME")
	userName := acctest.SkipIfEnvVarNotSet(t, "AWS_IDENTITY_STORE_USER_NAME")

	groupIdentity := tfstatecheck.Identity()
	userIdentity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstancesWithRegion(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		CheckDestroy:             testAccCheckAccountAssignmentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccountAssignment/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:                 config.StringVariable(rName),
					"AWS_IDENTITY_STORE_GROUP_NAME": config.StringVariable(groupName),
					"AWS_IDENTITY_STORE_USER_NAME":  config.StringVariable(userName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					groupIdentity.GetIdentity(resourceNameGroup),
					userIdentity.GetIdentity(resourceNameUser),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/AccountAssignment/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:                 config.StringVariable(rName),
					"AWS_IDENTITY_STORE_GROUP_NAME": config.StringVariable(groupName),
					"AWS_IDENTITY_STORE_USER_NAME":  config.StringVariable(userName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_account_assignment.test", groupIdentity.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_ssoadmin_account_assignment.test", tfqueryfilter.ByResourceIdentityFunc(groupIdentity.Checks())),
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_account_assignment.test", userIdentity.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_ssoadmin_account_assignment.test", tfqueryfilter.ByResourceIdentityFunc(userIdentity.Checks())),
				},
			},
		},
	})
}

func TestAccSSOAdminAccountAssignment_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_ssoadmin_account_assignment.group"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	groupName := acctest.SkipIfEnvVarNotSet(t, "AWS_IDENTITY_STORE_GROUP_NAME")

	identity := tfstatecheck.Identity()
	assignmentID := tfstatecheck.StateValue()
	instanceARN := tfstatecheck.StateValue()
	permissionSetARN := tfstatecheck.StateValue()
	principalID := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstancesWithRegion(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		CheckDestroy:             testAccCheckAccountAssignmentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccountAssignment/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:                 config.StringVariable(rName),
					"AWS_IDENTITY_STORE_GROUP_NAME": config.StringVariable(groupName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					assignmentID.GetStateValue(resourceName, tfjsonpath.New(names.AttrID)),
					instanceARN.GetStateValue(resourceName, tfjsonpath.New("instance_arn")),
					permissionSetARN.GetStateValue(resourceName, tfjsonpath.New("permission_set_arn")),
					principalID.GetStateValue(resourceName, tfjsonpath.New("principal_id")),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/AccountAssignment/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:                 config.StringVariable(rName),
					"AWS_IDENTITY_STORE_GROUP_NAME": config.StringVariable(groupName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_account_assignment.test", identity.Checks()),
					querycheck.ExpectResourceKnownValues("aws_ssoadmin_account_assignment.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), assignmentID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("instance_arn"), instanceARN.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("permission_set_arn"), permissionSetARN.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("principal_id"), principalID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("principal_type"), knownvalue.StringExact("GROUP")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("target_id"), tfknownvalue.AccountID()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("target_type"), knownvalue.StringExact("AWS_ACCOUNT")),
					}),
				},
			},
		},
	})
}

func TestAccSSOAdminAccountAssignment_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceNameGroup := "aws_ssoadmin_account_assignment.group"
	resourceNameUser := "aws_ssoadmin_account_assignment.user"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	groupName := acctest.SkipIfEnvVarNotSet(t, "AWS_IDENTITY_STORE_GROUP_NAME")
	userName := acctest.SkipIfEnvVarNotSet(t, "AWS_IDENTITY_STORE_USER_NAME")

	groupIdentity := tfstatecheck.Identity()
	userIdentity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckSSOAdminInstancesWithRegion(ctx, t, acctest.AlternateRegion())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AccountAssignment/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:                 config.StringVariable(rName),
					"region":                        config.StringVariable(acctest.AlternateRegion()),
					"AWS_IDENTITY_STORE_GROUP_NAME": config.StringVariable(groupName),
					"AWS_IDENTITY_STORE_USER_NAME":  config.StringVariable(userName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					groupIdentity.GetIdentity(resourceNameGroup),
					userIdentity.GetIdentity(resourceNameUser),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/AccountAssignment/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:                 config.StringVariable(rName),
					"region":                        config.StringVariable(acctest.AlternateRegion()),
					"AWS_IDENTITY_STORE_GROUP_NAME": config.StringVariable(groupName),
					"AWS_IDENTITY_STORE_USER_NAME":  config.StringVariable(userName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_account_assignment.test", groupIdentity.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_account_assignment.test", userIdentity.Checks()),
				},
			},
		},
	})
}
