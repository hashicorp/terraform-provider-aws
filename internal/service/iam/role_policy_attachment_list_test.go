// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMRolePolicyAttachment_List_Basic(t *testing.T) {
	ctx := acctest.Context(t)

	customerManagedName1 := "aws_iam_role_policy_attachment.customer_managed[0]"
	customerManagedName2 := "aws_iam_role_policy_attachment.customer_managed[1]"
	awsManagedName1 := "aws_iam_role_policy_attachment.aws_managed[0]"
	awsManagedName2 := "aws_iam_role_policy_attachment.aws_managed[1]"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()
	identity3 := tfstatecheck.Identity()
	identity4 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.IAMServiceID),
		CheckDestroy: testAccCheckRoleDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/RolePolicyAttachment/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(customerManagedName1),
					statecheck.ExpectKnownValue(customerManagedName1, tfjsonpath.New(names.AttrRole), knownvalue.StringExact(rName+"-0")),
					statecheck.CompareValuePairs(customerManagedName1, tfjsonpath.New("policy_arn"), "aws_iam_policy.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),

					identity2.GetIdentity(customerManagedName2),
					statecheck.ExpectKnownValue(customerManagedName2, tfjsonpath.New(names.AttrRole), knownvalue.StringExact(rName+"-1")),
					statecheck.CompareValuePairs(customerManagedName2, tfjsonpath.New("policy_arn"), "aws_iam_policy.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),

					identity3.GetIdentity(awsManagedName1),
					statecheck.ExpectKnownValue(awsManagedName1, tfjsonpath.New(names.AttrRole), knownvalue.StringExact(rName+"-0")),
					statecheck.CompareValuePairs(awsManagedName1, tfjsonpath.New("policy_arn"), "data.aws_iam_policy.AmazonDynamoDBReadOnlyAccess", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),

					identity4.GetIdentity(awsManagedName2),
					statecheck.ExpectKnownValue(awsManagedName2, tfjsonpath.New(names.AttrRole), knownvalue.StringExact(rName+"-1")),
					statecheck.CompareValuePairs(awsManagedName2, tfjsonpath.New("policy_arn"), "data.aws_iam_policy.AmazonDynamoDBReadOnlyAccess", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/RolePolicyAttachment/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_iam_role_policy_attachment.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_iam_role_policy_attachment.test", identity2.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_iam_role_policy_attachment.test", identity3.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_iam_role_policy_attachment.test", identity4.Checks()),
				},
			},
		},
	})
}
