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

func TestAccEC2ApplicationStatusCheck_List_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceTypeName := "aws_ec2_application_status_check.test"
	resourceName1 := "aws_ec2_application_status_check.test[0]"
	resourceName2 := "aws_ec2_application_status_check.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
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
		CheckDestroy:             testAccCheckApplicationStatusCheckDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationStatusCheck/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationStatusCheck/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(resourceTypeName, identity1.Checks()),
					querycheck.ExpectResourceDisplayName(resourceTypeName, tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^`+rName+`-0 \(asc-[a-z0-9]+\)$`))),
					tfquerycheck.ExpectNoResourceObject(resourceTypeName, tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),
					tfquerycheck.ExpectIdentityFunc(resourceTypeName, identity2.Checks()),
					querycheck.ExpectResourceDisplayName(resourceTypeName, tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^`+rName+`-1 \(asc-[a-z0-9]+\)$`))),
					tfquerycheck.ExpectNoResourceObject(resourceTypeName, tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccEC2ApplicationStatusCheck_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_application_status_check.test"
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
		CheckDestroy:             testAccCheckApplicationStatusCheckDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationStatusCheck/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationStatusCheck/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(resourceName, identity.Checks()),
					querycheck.ExpectResourceDisplayName(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^asc-[a-z0-9]+$`))),
					querycheck.ExpectResourceKnownValues(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("aggregation"), knownvalue.StringExact("included")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("device_index"), knownvalue.Int64Exact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("failure_threshold"), knownvalue.Int64Exact(2)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("health_check_path"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.StringRegexp(regexache.MustCompile(`^asc-[a-z0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("initialization_grace_period_seconds"), knownvalue.Int64Exact(300)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrInterval), knownvalue.Int64Exact(60)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ip_scope"), knownvalue.StringExact("private")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ip_version"), knownvalue.StringExact("ipv4")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPath), knownvalue.StringExact("/")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPort), knownvalue.Int64Exact(80)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrProtocol), knownvalue.StringExact("http")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("status_code_matcher"), knownvalue.StringExact("200")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("success_threshold"), knownvalue.Int64Exact(2)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTimeout), knownvalue.Int64Exact(6)),
					}),
				},
			},
		},
	})
}

func TestAccEC2ApplicationStatusCheck_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)
	resourceTypeName := "aws_ec2_application_status_check.test"
	resourceName1 := "aws_ec2_application_status_check.test[0]"
	resourceName2 := "aws_ec2_application_status_check.test[1]"
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
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationStatusCheck/list_region_override/"),
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
				ConfigDirectory: config.StaticDirectory("testdata/ApplicationStatusCheck/list_region_override/"),
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
