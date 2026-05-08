// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53_test

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
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53Record_List_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName1 := "aws_route53_record.test[0]"
	resourceName2 := "aws_route53_record.test[1]"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Record/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("zone_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrName), knownvalue.StringRegexp(regexache.MustCompile(rName+`-0\..*\.com\.?`))),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrType), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("ttl"), knownvalue.Int64Exact(300)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("records"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("10.0.0.0"),
					})),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("zone_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrName), knownvalue.StringRegexp(regexache.MustCompile(rName+`-1\..*\.com\.?`))),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrType), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("ttl"), knownvalue.Int64Exact(300)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("records"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("10.0.0.1"),
					})),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Record/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_route53_record.test", map[string]knownvalue.Check{
						"zone_id":           knownvalue.NotNull(),
						names.AttrName:      knownvalue.StringRegexp(regexache.MustCompile(rName + `-0\.` + rName + `\.com\.?`)),
						names.AttrType:      knownvalue.StringExact("A"),
						"set_identifier":    knownvalue.Null(),
						names.AttrAccountID: knownvalue.NotNull(),
					}),
					querycheck.ExpectIdentity("aws_route53_record.test", map[string]knownvalue.Check{
						"zone_id":           knownvalue.NotNull(),
						names.AttrName:      knownvalue.StringRegexp(regexache.MustCompile(rName + `-1\.` + rName + `\.com\.?`)),
						names.AttrType:      knownvalue.StringExact("A"),
						"set_identifier":    knownvalue.Null(),
						names.AttrAccountID: knownvalue.NotNull(),
					}),
				},
			},
		},
	})
}
