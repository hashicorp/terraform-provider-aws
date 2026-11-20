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

func TestAccSSOAdminManagedPolicyAttachment_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ssoadmin_managed_policy_attachment.test[0]"
	resourceName2 := "aws_ssoadmin_managed_policy_attachment.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstancesWithRegion(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		CheckDestroy:             testAccCheckManagedPolicyAttachmentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ManagedPolicyAttachment/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ManagedPolicyAttachment/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_managed_policy_attachment.test", identity1.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_ssoadmin_managed_policy_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_managed_policy_attachment.test", identity2.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_ssoadmin_managed_policy_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccSSOAdminManagedPolicyAttachment_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity := tfstatecheck.Identity()
	attachmentID := tfstatecheck.StateValue()
	instanceARN := tfstatecheck.StateValue()
	managedPolicyARN := tfstatecheck.StateValue()
	permissionSetARN := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstancesWithRegion(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		CheckDestroy:             testAccCheckManagedPolicyAttachmentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ManagedPolicyAttachment/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					attachmentID.GetStateValue(resourceName, tfjsonpath.New(names.AttrID)),
					instanceARN.GetStateValue(resourceName, tfjsonpath.New("instance_arn")),
					managedPolicyARN.GetStateValue(resourceName, tfjsonpath.New("managed_policy_arn")),
					permissionSetARN.GetStateValue(resourceName, tfjsonpath.New("permission_set_arn")),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ManagedPolicyAttachment/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_managed_policy_attachment.test", identity.Checks()),
					querycheck.ExpectResourceKnownValues("aws_ssoadmin_managed_policy_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), attachmentID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("instance_arn"), instanceARN.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("managed_policy_arn"), managedPolicyARN.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("managed_policy_name"), knownvalue.StringExact("AmazonCognitoReadOnly")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("permission_set_arn"), permissionSetARN.ValueCheck()),
					}),
				},
			},
		},
	})
}

func TestAccSSOAdminManagedPolicyAttachment_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ssoadmin_managed_policy_attachment.test[0]"
	resourceName2 := "aws_ssoadmin_managed_policy_attachment.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

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
				ConfigDirectory: config.StaticDirectory("testdata/ManagedPolicyAttachment/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ManagedPolicyAttachment/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_managed_policy_attachment.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_ssoadmin_managed_policy_attachment.test", identity2.Checks()),
				},
			},
		},
	})
}
