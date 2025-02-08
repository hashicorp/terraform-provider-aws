// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBatchJobQueue_migrateFromComputeEnvironments(t *testing.T) {
	ctx := acctest.Context(t)
	var jobQueue1 awstypes.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.BatchServiceID),
		CheckDestroy: testAccCheckJobQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.23.0",
					},
				},
				Config: testAccJobQueueConfig_computeEnvironments_initial(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobQueueExists(ctx, resourceName, &jobQueue1),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccJobQueueConfig_computeEnvironments_updated(rName),
				PlanOnly:                 true,
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccJobQueueConfig_computeEnvironments_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobQueueExists(ctx, resourceName, &jobQueue1),
				),
			},
		},
	})
}

func testAccJobQueueConfig_computeEnvironments_initial(rName string) string {
	return acctest.ConfigCompose(
		testAccJobQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environments = [aws_batch_compute_environment.test.arn]
  name                 = %[1]q
  priority             = 1
  state                = "ENABLED"
}
`, rName))
}

func testAccJobQueueConfig_computeEnvironments_updated(rName string) string {
	return acctest.ConfigCompose(
		testAccJobQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environment_order {
    compute_environment = aws_batch_compute_environment.test.arn
    order               = 0
  }
  name     = %[1]q
  priority = 1
  state    = "ENABLED"
}
`, rName))
}
