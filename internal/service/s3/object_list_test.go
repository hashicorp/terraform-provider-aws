// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3Object_List_Basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_s3_object.test[0]"
	resourceName2 := "aws_s3_object.test[1]"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Object/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrKey), knownvalue.StringExact(rName+"-0")),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrKey), knownvalue.StringExact(rName+"-1")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Object/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_s3_object.test", map[string]knownvalue.Check{
						names.AttrBucket:    knownvalue.StringExact(rName),
						names.AttrKey:       knownvalue.StringExact(rName + "-0"),
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
					}),
					querycheck.ExpectIdentity("aws_s3_object.test", map[string]knownvalue.Check{
						names.AttrBucket:    knownvalue.StringExact(rName),
						names.AttrKey:       knownvalue.StringExact(rName + "-1"),
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
					}),
				},
			},
		},
	})
}

func TestAccS3Object_List_Prefix(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_s3_object.test[0]"
	resourceName2 := "aws_s3_object.test[1]"
	resourceName3 := "aws_s3_object.other[0]"
	resourceName4 := "aws_s3_object.other[1]"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Object/list_prefix/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrKey), knownvalue.StringExact("prefix-"+rName+"-0")),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrKey), knownvalue.StringExact("prefix-"+rName+"-1")),
					statecheck.ExpectKnownValue(resourceName3, tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName3, tfjsonpath.New(names.AttrKey), knownvalue.StringExact("other-"+rName+"-0")),
					statecheck.ExpectKnownValue(resourceName4, tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName4, tfjsonpath.New(names.AttrKey), knownvalue.StringExact("other-"+rName+"-1")),
				},
			},

			// Step 2: Query only for prefix- objects
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Object/list_prefix/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_s3_object.test", map[string]knownvalue.Check{
						names.AttrBucket:    knownvalue.StringExact(rName),
						names.AttrKey:       knownvalue.StringExact("prefix-" + rName + "-0"),
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
					}),
					querycheck.ExpectIdentity("aws_s3_object.test", map[string]knownvalue.Check{
						names.AttrBucket:    knownvalue.StringExact(rName),
						names.AttrKey:       knownvalue.StringExact("prefix-" + rName + "-1"),
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
					}),
				},
			},
		},
	})
}
