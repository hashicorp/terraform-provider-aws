// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccEC2ApplicationStatusCheckAssociation_List_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceTypeName := "aws_ec2_application_status_check_association.test"
	resourceName1 := "aws_ec2_application_status_check_association.test[0]"
	resourceName2 := "aws_ec2_application_status_check_association.test[1]"
	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckApplicationStatusCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckApplicationStatusCheckAssociationDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationStatusCheckAssociation/list_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationStatusCheckAssociation/list_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(resourceTypeName, identity1.Checks()),
					querycheck.ExpectResourceDisplayName(resourceTypeName, tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^asc-[a-z0-9]+,tag,Environment,production$`))),
					tfquerycheck.ExpectNoResourceObject(resourceTypeName, tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),
					tfquerycheck.ExpectIdentityFunc(resourceTypeName, identity2.Checks()),
					querycheck.ExpectResourceDisplayName(resourceTypeName, tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^asc-[a-z0-9]+,tag,Environment,production$`))),
					tfquerycheck.ExpectNoResourceObject(resourceTypeName, tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccEC2ApplicationStatusCheckAssociation_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_application_status_check_association.test"
	identity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckApplicationStatusCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckApplicationStatusCheckAssociationDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationStatusCheckAssociation/list_include_resource/"),
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationStatusCheckAssociation/list_include_resource/"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(resourceName, identity.Checks()),
					querycheck.ExpectResourceDisplayName(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^asc-[a-z0-9]+,tag,Environment,production$`))),
					querycheck.ExpectResourceKnownValues(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("application_status_check_id"), knownvalue.StringRegexp(regexache.MustCompile(`^asc-[a-z0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrInstanceID), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("target_tag_key"), knownvalue.StringExact("Environment")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("target_tag_value"), knownvalue.StringExact("production")),
					}),
				},
			},
		},
	})
}

func TestAccEC2ApplicationStatusCheckAssociation_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)
	resourceTypeName := "aws_ec2_application_status_check_association.test"
	resourceName1 := "aws_ec2_application_status_check_association.test[0]"
	resourceName2 := "aws_ec2_application_status_check_association.test[1]"
	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckApplicationStatusCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationStatusCheckAssociation/list_region_override/"),
				ConfigVariables: config.Variables{
					"region":         config.StringVariable(acctest.AlternateRegion()),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationStatusCheckAssociation/list_region_override/"),
				ConfigVariables: config.Variables{
					"region":         config.StringVariable(acctest.AlternateRegion()),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(resourceTypeName, identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc(resourceTypeName, identity2.Checks()),
				},
			},
		},
	})
}
