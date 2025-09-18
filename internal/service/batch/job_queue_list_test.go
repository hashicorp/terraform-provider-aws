// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

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

func TestAccBatchJobQueue_List_Basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_batch_job_queue.test[0]"
	resourceName2 := "aws_batch_job_queue.test[1]"
	resourceName3 := "aws_batch_job_queue.test[2]"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.BatchServiceID),
		CheckDestroy: testAccCheckJobQueueDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/JobQueue/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("batch", "job-queue/"+rName+"-0")),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("batch", "job-queue/"+rName+"-1")),
					statecheck.ExpectKnownValue(resourceName3, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("batch", "job-queue/"+rName+"-2")),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/JobQueue/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigQueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_batch_job_queue.test", map[string]knownvalue.Check{
						names.AttrARN: tfknownvalue.RegionalARNExact("batch", "job-queue/"+rName+"-0"),
					}),
					querycheck.ExpectIdentity("aws_batch_job_queue.test", map[string]knownvalue.Check{
						names.AttrARN: tfknownvalue.RegionalARNExact("batch", "job-queue/"+rName+"-1"),
					}),
					querycheck.ExpectIdentity("aws_batch_job_queue.test", map[string]knownvalue.Check{
						names.AttrARN: tfknownvalue.RegionalARNExact("batch", "job-queue/"+rName+"-2"),
					}),
				},
			},
		},
	})
}

func TestAccBatchJobQueue_List_RegionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_batch_job_queue.test[0]"
	resourceName2 := "aws_batch_job_queue.test[1]"
	resourceName3 := "aws_batch_job_queue.test[2]"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.BatchServiceID),
		CheckDestroy: testAccCheckJobQueueDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/JobQueue/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionExact("batch", "job-queue/"+rName+"-0")),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionExact("batch", "job-queue/"+rName+"-1")),
					statecheck.ExpectKnownValue(resourceName3, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionExact("batch", "job-queue/"+rName+"-2")),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/JobQueue/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigQueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_batch_job_queue.test", map[string]knownvalue.Check{
						names.AttrARN: tfknownvalue.RegionalARNAlternateRegionExact("batch", "job-queue/"+rName+"-0"),
					}),
					querycheck.ExpectIdentity("aws_batch_job_queue.test", map[string]knownvalue.Check{
						names.AttrARN: tfknownvalue.RegionalARNAlternateRegionExact("batch", "job-queue/"+rName+"-1"),
					}),
					querycheck.ExpectIdentity("aws_batch_job_queue.test", map[string]knownvalue.Check{
						names.AttrARN: tfknownvalue.RegionalARNAlternateRegionExact("batch", "job-queue/"+rName+"-2"),
					}),
				},
			},
		},
	})
}
