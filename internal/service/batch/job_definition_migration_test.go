// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBatchJobDefinition_MigrateFromSDKToPluginFramework(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_batch_job_definition.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.BatchServiceID),
		CheckDestroy: testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.23.0", // Ensure to use the latest provider version
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobDefinitionExists(ctx, resourceName),
				),
				Config: testAccJobDefinitionConfig_initial(rName),
			},
			{
				Config:                   testAccJobDefinitionConfig_updated(rName),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
			},
			{
				Config:                   testAccJobDefinitionConfig_updated(rName),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "container_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container_properties.0.memory", "1024"),
				),
			},
		},
	})
}

func testAccJobDefinitionConfig_initial(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = "%s"
  type = "container"

  container_properties = jsonencode({
    image   = "my-image"
    vcpus   = 1
    memory  = 1024
    command = ["echo", "hello"]
  })
}`, rName)
}

func testAccJobDefinitionConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  name = "%s"
  type = "container"

  container_properties {
    image   = "my-image"
    vcpus   = 1
    memory  = 1024
    command = ["echo", "world"]
  }
}`, rName)
}
