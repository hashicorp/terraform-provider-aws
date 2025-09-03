// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBatchJobQueue_List_Basic(t *testing.T) {
	ctx := acctest.Context(t)
	// var v1, v2, v3 awstypes.JobQueueDetail
	resourceName1 := "aws_batch_job_queue.test[0]"
	resourceName2 := "aws_batch_job_queue.test[1]"
	resourceName3 := "aws_batch_job_queue.test[2]"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.BatchServiceID),
		CheckDestroy: testAccCheckJobQueueDestroy(ctx),
		// ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				// ConfigDirectory:          config.StaticDirectory("testdata/JobQueue/list_basic/"),
				// ConfigVariables: config.Variables{
				// 	acctest.CtRName: config.StringVariable(rName),
				// },
				Config: fmt.Sprintf(`
provider "aws" {}

resource "aws_batch_job_queue" "test" {
  count = 3

  name     = "%[1]s-${count.index}"
  priority = 1
  state    = "DISABLED"

  compute_environment_order {
    compute_environment = aws_batch_compute_environment.test.arn
    order               = 1
  }
}

resource "aws_batch_compute_environment" "test" {
  name         = "%[1]s"
  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"

  depends_on = [aws_iam_role_policy_attachment.batch_service]
}

data "aws_partition" "current" {}

resource "aws_iam_role" "batch_service" {
  name = "%[1]s-batch-service"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": "batch.${data.aws_partition.current.dns_suffix}"
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "batch_service" {
  role       = aws_iam_role.batch_service.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_iam_role" "ecs_instance" {
  name = "%[1]s-ecs-instance"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
        }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ecs_instance" {
  role       = aws_iam_role.ecs_instance.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance" {
  name = aws_iam_role.ecs_instance.name
  role = aws_iam_role_policy_attachment.ecs_instance.role
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
				// testAccCheckJobQueueExists(ctx, resourceName1, &v1),
				// testAccCheckJobQueueExists(ctx, resourceName2, &v2),
				// testAccCheckJobQueueExists(ctx, resourceName3, &v3),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "batch", "job-queue/{name}"),
					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "batch", "job-queue/{name}"),
					tfstatecheck.ExpectRegionalARNFormat(resourceName3, tfjsonpath.New(names.AttrARN), "batch", "job-queue/{name}"),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				// ConfigDirectory:          config.StaticDirectory("testdata/JobQueue/list_basic/"),
				// ConfigVariables: config.Variables{
				// 	acctest.CtRName: config.StringVariable(rName),
				// },
				Config: `
// provider "aws" {}

list "aws_batch_job_queue" "test" {
	provider = aws
}
`,
				ConfigQueryChecks: []querycheck.QueryCheck{},
			},
		},
	})
}

// func TestAccBatchJobQueue_List_RegionOverride(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	// var v1, v2, v3 awstypes.JobQueueDetail
// 	resourceName1 := "aws_batch_job_queue.test[0]"
// 	resourceName2 := "aws_batch_job_queue.test[1]"
// 	resourceName3 := "aws_batch_job_queue.test[2]"
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

// 	resource.ParallelTest(t, resource.TestCase{
// 		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
// 			tfversion.SkipBelow(tfversion.Version1_14_0),
// 		},
// 		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
// 		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
// 		CheckDestroy:             testAccCheckJobQueueDestroy(ctx),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		Steps: []resource.TestStep{
// 			// Step 1: Setup
// 			{
// 				ConfigDirectory: config.StaticDirectory("testdata/JobQueue/list_region_override/"),
// 				ConfigVariables: config.Variables{
// 					acctest.CtRName: config.StringVariable(rName),
// 					"region":        config.StringVariable(acctest.AlternateRegion()),
// 				},
// 				// Check: resource.ComposeAggregateTestCheckFunc(
// 				// 	testAccCheckJobQueueExists(ctx, resourceName1, &v1),
// 				// 	testAccCheckJobQueueExists(ctx, resourceName2, &v2),
// 				// 	testAccCheckJobQueueExists(ctx, resourceName3, &v3),
// 				// ),
// 				ConfigStateChecks: []statecheck.StateCheck{
// 					tfstatecheck.ExpectRegionalARNFormat(resourceName1, tfjsonpath.New(names.AttrARN), "batch", "job-queue/{name}"),
// 					tfstatecheck.ExpectRegionalARNFormat(resourceName2, tfjsonpath.New(names.AttrARN), "batch", "job-queue/{name}"),
// 					tfstatecheck.ExpectRegionalARNFormat(resourceName3, tfjsonpath.New(names.AttrARN), "batch", "job-queue/{name}"),
// 				},
// 			},

// 			// Step 2: Query
// 			{
// 				Query:           true,
// 				ConfigDirectory: config.StaticDirectory("testdata/JobQueue/list_region_override/"),
// 				ConfigVariables: config.Variables{
// 					acctest.CtRName: config.StringVariable(rName),
// 					"region":        config.StringVariable(acctest.AlternateRegion()),
// 				},
// 			},
// 		},
// 	})
// }
