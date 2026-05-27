// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
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

func TestAccS3ControlMultiRegionAccessPoint_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_s3control_multi_region_access_point.test[0]"
	resourceName2 := "aws_s3control_multi_region_access_point.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		CheckDestroy:             testAccCheckMultiRegionAccessPointDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/MultiRegionAccessPoint/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-1")),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/MultiRegionAccessPoint/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_s3control_multi_region_access_point.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_s3control_multi_region_access_point.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_s3control_multi_region_access_point.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_s3control_multi_region_access_point.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_s3control_multi_region_access_point.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"-1")),
					tfquerycheck.ExpectNoResourceObject("aws_s3control_multi_region_access_point.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccS3ControlMultiRegionAccessPoint_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_s3control_multi_region_access_point.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	id1 := tfstatecheck.StateValue()
	alias1 := tfstatecheck.StateValue()
	arn1 := tfstatecheck.StateValue()
	domainName1 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		CheckDestroy:             testAccCheckMultiRegionAccessPointDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/MultiRegionAccessPoint/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),

					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					alias1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrAlias)),
					arn1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrARN)),
					domainName1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrDomainName)),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/MultiRegionAccessPoint/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_s3control_multi_region_access_point.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_s3control_multi_region_access_point.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					querycheck.ExpectResourceKnownValues("aws_s3control_multi_region_access_point.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), id1.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAccountID), tfknownvalue.AccountID()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAlias), alias1.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), arn1.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDomainName), domainName1.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("READY")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("details"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact(rName + "-0"),
								"public_access_block": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"block_public_acls":       knownvalue.Bool(true),
										"block_public_policy":     knownvalue.Bool(true),
										"ignore_public_acls":      knownvalue.Bool(true),
										"restrict_public_buckets": knownvalue.Bool(true),
									}),
								}),
								names.AttrRegion: knownvalue.SetExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										names.AttrBucket:    knownvalue.StringExact(rName + "-0"),
										"bucket_account_id": tfknownvalue.AccountID(),
										names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
									}),
								}),
							}),
						})),
					}),
				},
			},
		},
	})
}
